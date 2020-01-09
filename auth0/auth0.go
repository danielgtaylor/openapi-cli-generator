package auth0

import (
	"net/url"

	"github.com/danielgtaylor/openapi-cli-generator/oauth"
)

type config struct {
	extra  []string
	scopes []string
}

// Extra provides the names of additional parameters to use to store information
// in user profiles. Use `cli.GetProfile("default")["name"]` to access it.
func Extra(names ...string) func(*config) error {
	return func(c *config) error {
		c.extra = names
		return nil
	}
}

// Scopes are used to request additional information or features for the
// returned token.
func Scopes(scopes ...string) func(*config) error {
	return func(c *config) error {
		c.scopes = scopes
		return nil
	}
}

// InitClientCredentials sets up the Auth0 client credentials flow. Must be
// called *after* you have called `cli.Init()`. Pass in profile-related extra
// variables to store them alongside the default profile information.
func InitClientCredentials(issuer string, options ...func(*config) error) {
	var c config

	for _, option := range options {
		if err := option(&c); err != nil {
			panic(err)
		}
	}

	oauth.InitClientCredentials(issuer+"oauth/token",
		oauth.Scopes(c.scopes...),
		oauth.Extra(append([]string{"audience"}, c.extra...)...),
		oauth.GetParams(func(profile map[string]string) url.Values {
			return url.Values{
				"audience": {profile["audience"]},
			}
		}))
}

// InitAuthCode sets up the Auth0 authorization code flow. Must be
// called *after* you have called `cli.Init()`. Pass in profile-related extra
// variables to store them alongside the default profile information.
func InitAuthCode(clientID string, issuer string, options ...func(*config) error) {
	var c config

	for _, option := range options {
		if err := option(&c); err != nil {
			panic(err)
		}
	}

	oauth.InitAuthCode(clientID, issuer+"authorize", issuer+"oauth/token",
		oauth.Scopes(c.scopes...),
		oauth.Extra(append([]string{"audience"}, c.extra...)...),
		oauth.GetParams(func(profile map[string]string) url.Values {
			return url.Values{
				"audience": {profile["audience"]},
			}
		}))
}
