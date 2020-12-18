package oauth

import (
	"net/http"
	"net/url"

	"github.com/rigetti/openapi-cli-generator/cli"
	"github.com/rs/zerolog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// NewClientCredentialsHandler creates a new handler.
func NewClientCredentialsHandler(tokenURL string, keys, params, scopes []string) *ClientCredentialsHandler {
	return &ClientCredentialsHandler{
		TokenURL: tokenURL,
		Keys:     append([]string{"client-id", "client-secret"}, keys...),
		Params:   params,
		Scopes:   scopes,
	}
}

// ClientCredentialsHandler implements the Client Credentials OAuth2 flow.
type ClientCredentialsHandler struct {
	TokenURL     string
	Keys         []string
	Params       []string
	Scopes       []string
	ClientID     string
	ClientSecret string
	Values       map[string]interface{}

	getParamsFunc func(profile map[string]string) url.Values
}

// ProfileKeys returns the key names for fields to store in the profile.
func (h *ClientCredentialsHandler) ProfileKeys() []string {
	return h.Keys
}

// ExecuteFlow gets run before the request goes out on the wre.
func (h *ClientCredentialsHandler) getSource(log *zerolog.Logger) oauth2.TokenSource {
	// No auth is set, so let's get the token either from a cache
	// or generate a new one from the issuing server.
	clientID := h.ClientID
	clientSecret := h.ClientSecret

	params := url.Values{}
	for _, name := range h.Params {
		value, _ := h.Values[name]
		if s, ok := value.(string); ok {
			params.Add(name, s)
		}
	}

	return (&clientcredentials.Config{
		ClientID:       clientID,
		ClientSecret:   clientSecret,
		TokenURL:       h.TokenURL,
		EndpointParams: params,
		Scopes:         h.Scopes,
	}).TokenSource(oauth2.NoContext)

}
func (h *ClientCredentialsHandler) ExecuteFlow(log *zerolog.Logger) (*oauth2.Token, error) {
	source := h.getSource(log)
	return getOauth2Token(source, log)
}

// OnRequest gets run before the request goes out on the wire.
func (h *ClientCredentialsHandler) OnRequest(log *zerolog.Logger, request *http.Request) error {
	if request.Header.Get("Authorization") == "" {
		source := h.getSource(log)
		return TokenHandler(source, log, request)
	}

	return nil
}

// InitClientCredentials sets up the OAuth 2.0 client credentials authentication
// flow. Must be called *after* you have called `cli.Init()`. The endpoint
// params allow you to pass additional info to the token URL. Pass in
// profile-related extra variables to store them alongside the default profile
// information.
func InitClientCredentials(tokenURL string, options ...func(*config) error) {
	var c config

	for _, option := range options {
		if err := option(&c); err != nil {
			panic(err)
		}
	}

	handler := NewClientCredentialsHandler(tokenURL, c.extra, []string{}, c.scopes)

	// Since you can pass a function to get params, we can't use the normal
	// preset `Params` field. We use an internal field here for backwards
	// compatibility only.
	handler.getParamsFunc = c.getParams

	cli.UseAuth("", handler)

	// TODO: retry on 401
	// cli.Client.UseResponse(func(ctx *context.Context, h context.Handler) {
	// 	h.Next(ctx)
	// })
}
