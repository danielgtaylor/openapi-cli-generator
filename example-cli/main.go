package main

import (
	"github.com/danielgtaylor/openapi-cli-generator/cli"
)

func main() {
	cli.Init(&cli.Config{
		AppName:   "example",
		EnvPrefix: "EXAMPLE",
		Version:   "1.0.0",
	})

	openapiRegister(false)

	cli.Root.Execute()
}
