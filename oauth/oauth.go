// Package oauth provides authentication profile support for APIs that require
// OAuth 2.0 auth.
package oauth

import (
	"net/url"

	"github.com/danielgtaylor/openapi-cli-generator/cli"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"gopkg.in/h2non/gentleman.v2/context"
)

type config struct {
	getParams func(profile map[string]string) url.Values
	extra     []string
}

// GetParams registers a function to get additional token endpoint parameters
// to include in the request when fetching a new token.
func GetParams(f func(profile map[string]string) url.Values) func(*config) error {
	return func(c *config) error {
		c.getParams = f
		return nil
	}
}

// Extra provides the names of additional parameters to use to store information
// in user profiles. Use `cli.GetProfile("default")["name"]` to access it.
func Extra(names ...string) func(*config) error {
	return func(c *config) error {
		c.extra = names
		return nil
	}
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

	standard := []string{"client-id", "client-secret"}

	cli.InitCredentials(
		cli.ProfileKeys(append(standard, c.extra...)...),
		cli.ProfileListKeys("client-id"))

	cli.Client.UseRequest(func(ctx *context.Context, h context.Handler) {
		if ctx.Request.Header.Get("Authorization") == "" {
			// No auth is set, so let's get the token either from a cache
			// or generate a new one from the issuing server.
			profile := cli.GetProfile()

			var params url.Values
			if c.getParams != nil {
				params = c.getParams(profile)
			}

			source := (&clientcredentials.Config{
				ClientID:       profile["client_id"],
				ClientSecret:   profile["client_secret"],
				TokenURL:       tokenURL,
				EndpointParams: params,
			}).TokenSource(oauth2.NoContext)

			TokenMiddleware(source, ctx, h)
		}

		h.Next(ctx)
	})

	// TODO: retry on 401
	// cli.Client.UseResponse(func(ctx *context.Context, h context.Handler) {
	// 	h.Next(ctx)
	// })
}

// TokenMiddleware takes a token source, gets a token, and modifies a request to
// add the token auth as a header. Uses the CLI cache to store tokens on a per-
// profile basis between runs.
func TokenMiddleware(source oauth2.TokenSource, ctx *context.Context, h context.Handler) {
	var cached *oauth2.Token

	// Setup logger with the current profile.
	log := ctx.Get("log").(*zerolog.Logger).
		With().Str("profile", viper.GetString("profile")).Logger()

	// Load any existing token from the CLI's cache file.
	expiresKey := "profiles." + viper.GetString("profile") + ".expires"
	typeKey := "profiles." + viper.GetString("profile") + ".type"
	tokenKey := "profiles." + viper.GetString("profile") + ".token"
	refreshKey := "profiles." + viper.GetString("profile") + ".refresh"

	expiry := cli.Cache.GetTime(expiresKey)
	if !expiry.IsZero() {
		log.Debug().Msg("Loading token from cache.")
		cached = &oauth2.Token{
			AccessToken:  cli.Cache.GetString(tokenKey),
			RefreshToken: cli.Cache.GetString(refreshKey),
			TokenType:    cli.Cache.GetString(typeKey),
			Expiry:       expiry,
		}
	}

	if cached != nil {
		// Wrap the token source preloaded with our cached token.
		source = oauth2.ReuseTokenSource(cached, source)
	}

	// Get the next available token from the source.
	token, err := source.Token()
	if err != nil {
		h.Error(ctx, err)
		return
	}

	if cached == nil || (token.AccessToken != cached.AccessToken) {
		// Token either didn't exist in the cache or has changed, so let's write
		// the new values to the CLI cache.
		log.Debug().Msg("Token refreshed. Updating cache.")

		cli.Cache.Set(expiresKey, token.Expiry)
		cli.Cache.Set(typeKey, token.Type())
		cli.Cache.Set(tokenKey, token.AccessToken)
		cli.Cache.Set(refreshKey, token.RefreshToken)

		// Save the cache to disk.
		if err := cli.Cache.WriteConfig(); err != nil {
			h.Error(ctx, err)
			return
		}
	}

	// Set the auth header so the request can be made.
	token.SetAuthHeader(ctx.Request)
}
