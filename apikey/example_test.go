package apikey

import (
	"fmt"

	"github.com/danielgtaylor/openapi-cli-generator/cli"
)

func Example() {
	// Initialize the CLI.
	cli.Init(&cli.Config{
		AppName:   "example",
		EnvPrefix: "EXAMPLE",
	})

	// Initialize the API key authentication.
	Init("X-API-Key", LocationHeader)

	// Mock out a profile to be used in the request.
	cli.Creds.Set("profiles.default.api_key", "my-secret")

	// Make a request.
	req := cli.Client.Get().URL("http://example.com/")
	if _, err := req.Do(); err != nil {
		panic(err)
	}

	// Look at the header that was used in the request. It should match the
	// profile's API key value.
	fmt.Println(req.Context.Request.Header.Get("X-API-Key"))
	// Output: my-secret
}
