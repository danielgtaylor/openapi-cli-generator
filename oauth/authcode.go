package oauth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"context"

	"github.com/rigetti/openapi-cli-generator/cli"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

// open opens the specified URL in the default browser regardless of OS.
func open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

// authHandler is an HTTP handler that takes a channel and sends the `code`
// query param when it gets a request.
type authHandler struct {
	codeChan      chan string
	errorChan     chan error
	expectedState string
}

// We use a simple HTTP server to receive the redirected call from the auth server containing the
// authorization code. The call may not include a token; it may include only an error message
// explaining the reason for the call's failure.
func (h authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var code string
	var message string

	state := r.URL.Query().Get("state")

	if h.expectedState != state {
		message = fmt.Sprintf("Incorrect state parameter in response; expected %s, received %s", h.expectedState, state)
		h.errorChan <- errors.New(message)
	} else {
		code = r.URL.Query().Get("code")

		if code == "" {
			errorDescription := r.URL.Query().Get("error_description")
			message = fmt.Sprintf("Authentication failed. %s", errorDescription)

			h.errorChan <- errors.New(errorDescription)
		} else {
			h.codeChan <- code
			message = "Login successful. Please return to the terminal. You may now close this window."
		}
	}

	w.Header().Set("Content-Type", "text/html")
	body := fmt.Sprintf("<html><body><p>%s</p></body></html>", message)
	w.Write([]byte(body))
}

// AuthorizationCodeTokenSource with PKCE as described in:
// https://www.oauth.com/oauth2-servers/pkce/
// This works by running a local HTTP server on a configurable port and then having the
// user log in through a web browser, which redirects to the local server with
// an authorization code. That code is then used to make another HTTP request
// to fetch an auth token (and refresh token). That token may then be
// used to make requests against the API.
type AuthorizationCodeTokenSource struct {
	ClientID       string
	AuthorizeURL   string
	TokenURL       string
	RedirectURI    *url.URL
	State          string
	EndpointParams *url.Values
	Scopes         []string
}

// Token generates a new token using an authorization code.
func (ac *AuthorizationCodeTokenSource) Token() (*oauth2.Token, error) {
	// Generate a random code verifier string
	verifierBytes := make([]byte, 32)
	if _, err := rand.Read(verifierBytes); err != nil {
		return nil, err
	}

	verifier := base64.RawURLEncoding.EncodeToString(verifierBytes)

	// Generate a code challenge. Only the challenge is sent when requesting a
	// code which allows us to keep it secret for now.
	shaBytes := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(shaBytes[:])

	redirectURI := ac.RedirectURI

	if redirectURI == nil {
		redirectURI, _ = url.Parse("http://localhost:8484")
	}

	state := uuid.New().String()

	// Generate a URL with the challenge to have the user log in.
	url := fmt.Sprintf("%s?response_type=code&code_challenge=%s&code_challenge_method=S256&client_id=%s&redirect_uri=%s&scope=%s&state=%s", ac.AuthorizeURL, challenge, ac.ClientID, redirectURI.String(), strings.Join(ac.Scopes, `%20`), state)

	if len(*ac.EndpointParams) > 0 {
		url += "&" + ac.EndpointParams.Encode()
	}

	// Run server before opening the user's browser so we are ready for any redirect.
	codeChan := make(chan string)
	errorChan := make(chan error)
	handler := authHandler{
		codeChan:      codeChan,
		errorChan:     errorChan,
		expectedState: string(state),
	}

	s := &http.Server{
		Addr:           fmt.Sprintf(":%s", redirectURI.Port()),
		Handler:        handler,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1024,
	}

	go func() {
		// Run in a goroutine until the server is closed or we get an error.
		if err := s.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Authentication failed; server exited unexpectedly")
		}
	}()

	// Open auth URL in browser, print for manual use in case open fails.
	fmt.Println("Open your browser to log in using the URL:")
	fmt.Println(url)
	open(url)

	// Get code from handler, exchange it for a token, and then return it. This
	// channel read blocks until one code becomes available.
	// There is currently no timeout.
	var code string

	select {
	case code = <-codeChan:
		if code == "" {
			log.Fatal().Msg("Authentication failed: no authorization code returned from server")
		}
	case err := <-errorChan:
		log.Fatal().Err(err).Msg("Authentication failed")
	}
	s.Shutdown(context.Background())

	payload := fmt.Sprintf("grant_type=authorization_code&client_id=%s&code_verifier=%s&code=%s&redirect_uri=%s", ac.ClientID, verifier, code, redirectURI.String())

	return requestToken(ac.TokenURL, payload)
}

// AuthCodeHandler sets up the OAuth 2.0 authorization code with PKCE authentication
// flow.
type AuthCodeHandler struct {
	ClientID     string
	AuthorizeURL string
	TokenURL     string
	RedirectURI  *url.URL
	Keys         []string
	Params       []string
	Scopes       []string

	getParamsFunc func(profile map[string]string) url.Values
}

// ProfileKeys returns the key names for fields to store in the profile.
func (h *AuthCodeHandler) ProfileKeys() []string {
	return h.Keys
}

func (h *AuthCodeHandler) getRefreshTokenSource(log *zerolog.Logger) RefreshTokenSource {
	// No auth is set, so let's get the token either from a cache
	// or generate a new one from the issuing server.
	params := url.Values{}

	source := &AuthorizationCodeTokenSource{
		ClientID:       h.ClientID,
		AuthorizeURL:   h.AuthorizeURL,
		TokenURL:       h.TokenURL,
		RedirectURI:    h.RedirectURI,
		EndpointParams: &params,
		Scopes:         h.Scopes,
	}

	// Try to get a cached refresh token from the current profile and use
	// it to wrap the auth code token source with a refreshing source.
	return RefreshTokenSource{
		ClientID:       h.ClientID,
		TokenURL:       h.TokenURL,
		EndpointParams: &params,
		RefreshToken:   cli.RunConfig.GetCredentials().TokenPayload.RefreshToken,
		TokenSource:    source,
	}
}

// ExecuteFlow gets run before the request goes out on the wire.
func (h *AuthCodeHandler) ExecuteFlow(log *zerolog.Logger) (*oauth2.Token, error) {
	source := h.getRefreshTokenSource(log)
	return getOauth2Token(source, log)
}

// OnRequest gets run before the request goes out on the wire.
func (h *AuthCodeHandler) OnRequest(log *zerolog.Logger, request *http.Request) error {
	if request.Header.Get("Authorization") == "" {
		source := h.getRefreshTokenSource(log)
		return TokenHandler(source, log, request)
	}

	return nil
}

// InitAuthCode sets up the OAuth 2.0 authorization code with PKCE authentication
// flow. Must be called *after* you have called `cli.Init()`. The endpoint
// params allow you to pass additional info to the token URL. Pass in
// profile-related extra variables to store them alongside the default profile
// information.
func InitAuthCode(clientID string, authorizeURL string, tokenURL string, options ...func(*config) error) {
	var c config

	for _, option := range options {
		if err := option(&c); err != nil {
			panic(err)
		}
	}

	handler := &AuthCodeHandler{
		ClientID:     clientID,
		AuthorizeURL: authorizeURL,
		TokenURL:     tokenURL,
		Scopes:       c.scopes,
		Keys:         c.extra,

		// Since you can pass a function to get params, we can't use the normal
		// preset `Params` field. We use an internal field here for backwards
		// compatibility only.
		getParamsFunc: c.getParams,
	}

	cli.UseAuth("", handler)
}
