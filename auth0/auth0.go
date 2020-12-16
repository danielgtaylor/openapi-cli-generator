package auth0

import (
	"github.com/rigetti/openapi-cli-generator/cli"
	"github.com/rigetti/openapi-cli-generator/oauth"
)

type config struct {
	typeName string
	extra    []string
	scopes   []string
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

// Type defines the type name of this auth mechanism.
func Type(name string) func(*config) error {
	return func(c *config) error {
		c.typeName = name
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

	handler := oauth.NewClientCredentialsHandler(issuer+"oauth/token", append([]string{"audience"}, c.extra...), []string{"audience"}, c.scopes)

	cli.UseAuth(c.typeName, handler)
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

	cli.UseAuth(c.typeName, &oauth.AuthCodeHandler{
		ClientID:     clientID,
		AuthorizeURL: issuer + "authorize",
		TokenURL:     issuer + "oauth/token",
		Keys:         append([]string{"audience"}, c.extra...),
		Params:       []string{"audience"},
		Scopes:       c.scopes,
	})
}
