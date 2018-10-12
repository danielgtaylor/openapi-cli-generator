package apikey

import (
	"testing"

	"github.com/danielgtaylor/openapi-cli-generator/cli"
	"github.com/stretchr/testify/assert"
)

func ExampleInit_header() {
	// Use a custom header for authentication.
	Init("X-API-Key", LocationHeader)
}

func ExampleInit_query() {
	// Use a query parameter for authentication.
	Init("apikey", LocationHeader)
}

func TestHeaderAuth(t *testing.T) {
	cli.Init(&cli.Config{
		AppName:   "test",
		EnvPrefix: "TEST",
	})
	Init("x-auth", LocationHeader)
	cli.Creds.Set("profiles.default.api_key", "test")

	r := cli.Client.Get()
	r.Do()

	assert.Equal(t, "test", r.Context.Request.Header.Get("x-auth"))
}

func TestQueryAuth(t *testing.T) {
	cli.Init(&cli.Config{
		AppName:   "test",
		EnvPrefix: "TEST",
	})
	Init("key", LocationQuery)
	cli.Creds.Set("profiles.default.api_key", "test")

	r := cli.Client.Get()
	r.Do()

	assert.Equal(t, "test", r.Context.Request.URL.Query().Get("key"))
}

func TestCookieAuth(t *testing.T) {
	cli.Init(&cli.Config{
		AppName:   "test",
		EnvPrefix: "TEST",
	})
	Init("key", LocationCookie)
	cli.Creds.Set("profiles.default.api_key", "test")

	r := cli.Client.Get()
	r.Do()

	cookie, err := r.Context.Request.Cookie("key")
	assert.NoError(t, err)
	assert.Equal(t, "test", cookie.Value)
}
