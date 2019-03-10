package main

import (
	"github.com/danielgtaylor/openapi-cli-generator/cli"
)

//go:generate openapi-cli-generator generate openapi.yaml

func main() {
	cli.Init(&cli.Config{
		AppName:   "example",
		EnvPrefix: "EXAMPLE",
		Version:   "1.0.0",
	})

	openapiRegister(false)

	cli.Root.Execute()
}
