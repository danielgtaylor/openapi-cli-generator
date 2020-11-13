package main

import (
<<<<<<< HEAD
	"github.com/rigetti/openapi-cli-generator/cli"
=======
	"github.com/kalzoo/openapi-cli-generator/cli"
>>>>>>> replace references from danielgtaylor to kalzoo github
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
