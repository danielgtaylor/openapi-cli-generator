package main

import (
	"github.com/rigetti/openapi-cli-generator/cli"
	"github.com/rigetti/openapi-cli-generator/oauth"
)

func main() {
	cli.Init(&cli.Config{
		AppName:   "example",
		EnvPrefix: "EXAMPLE",
		Version:   "1.0.0",
	})

	cli.UseAuth("bogus", &oauth.AuthCodeHandler{
		ClientID:     "bogus01",
		AuthorizeURL: "https://auth.qa.qcs.rigetti.com/v1/authorize",
		TokenURL:     "https://auth.qa.qcs.rigetti.com/v1/token",
		Keys:         []string{},
		Params:       []string{},
		Scopes:       []string{"offline_access"},
	})
	cli.UseAuth("", &oauth.AuthCodeHandler{
		ClientID:     "bogus02",
		AuthorizeURL: "https://auth.qa.qcs.rigetti.com/v1/authorize",
		TokenURL:     "https://auth.qa.qcs.rigetti.com/v1/token",
		Keys:         []string{},
		Params:       []string{},
		Scopes:       []string{"offline_access"},
	})

	openapiRegister(false)

	cli.Root.Execute()
}
