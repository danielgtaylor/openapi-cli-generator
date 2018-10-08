# OpenAPI CLI Generator

[![GoDoc](https://godoc.org/github.com/danielgtaylor/openapi-cli-generator?status.svg)](https://godoc.org/github.com/danielgtaylor/openapi-cli-generator)
[![Build Status](https://travis-ci.org/danielgtaylor/openapi-cli-generator.svg?branch=master)](https://travis-ci.org/danielgtaylor/openapi-cli-generator)
[![Go Report Card](https://goreportcard.com/badge/github.com/danielgtaylor/openapi-cli-generator)](https://goreportcard.com/report/github.com/danielgtaylor/openapi-cli-generator)
[![Platforms](https://img.shields.io/badge/platform-win%20%7C%20mac%20%7C%20linux-ligh.svg)](https://github.com/danielgtaylor/openapi-cli-generator/releases)

This project can be used to generate CLIs from OpenAPI 3 specs. The generated CLIs have the following features:

- Commands, subcommands, & flag parsing through [Cobra](https://github.com/spf13/cobra)
- Configuration through [Viper](https://github.com/spf13/viper)
  - JSON, YAML, or TOML config files in `/etc/` and `$HOME`, e.g. `{"verbose": true}` in `~/.my-app/config.json`
  - From environment: `APP_NAME_VERBOSE=1`
  - From flags: `--verbose`
- HTTP middleware through [Gentleman](https://github.com/h2non/gentleman/)
- Input through `stdin` or [CLI shorthand](https://github.com/danielgtaylor/openapi-cli-generator/tree/master/shorthand)
- Built-in cache to save data between runs
- Fast structured logging via [zerolog](https://github.com/rs/zerolog)
- Built-in support for [Auth0 client credentials](https://auth0.com/docs/api-auth/grant/client-credentials)
- Pretty output colored by [Chroma](https://github.com/alecthomas/chroma)
- Response filtering & projection by [JMESPath](http://jmespath.org/)

## Getting Started

First, make sure you have Go installed. Then, you can grab this project:

```sh
$ go get -u github.com/danielgtaylor/openapi-cli-generator
```

Next, make your project directory and generate the commands file.

```sh
# Set up your new project
$ mkdir my-cli && cd my-cli

# Create the default main file. The app name is used for config and env settings.
$ openapi-cli-generator init <app-name>

# Generate the commands
$ openapi-cli-generator generate openapi.yaml
```

Last, add a line like the following to your `main.go` file:

```go
openapiRegister(false)
```

If you would like to generate a client for many APIs and have each available under their own namespace, pass `true` instead. Next, build your client:

```sh
# Build & install the generated client.
$ go install

# Test it out!
$ my-cli --help
```

## OpenAPI Extensions

Several extensions properties may be used to change the behavior of the CLI.

Name | Description
---- | -----------
`x-cli-aliases` | Sets up command aliases for operations.
`x-cli-description` | Provide an alternate description for the CLI.
`x-cli-ignore` | Ignore this path, operation, or parameter.

### Aliases

The following example shows how you would set up a command that can be invoked by either `list-items` or simply `ls`:

```yaml
paths:
  /items:
    get:
      operationId: ListItems
      x-cli-aliases:
      - ls
```

### Description

You can override the default description easily:

```yaml
paths:
  /items:
    description: Some info talking about HTTP headers.
    x-cli-description: Some info talking about command line arguments.
```

### Exclusion

It is possible to exclude paths, operations, and/or parameters from the generated CLI.

```yaml
paths:
  /included:
    description: I will get included in the CLI.
  /excluded:
    x-cli-ignore: true
    description: I will not be in the CLI :-(
```

## Customization

Your `main.go` is the entrypoint to your generated CLI, and may be customized to add additional logic and features. For example, you might set custom headers or handle auth before a request goes out on the wire. The `auth0` module provides a sample implementation.

### Configuration Description

TODO: Show table describing all well-known configuration keys.

### Custom Global Flags

It's possible to supply custom flags and a pre-run function. For example, say your OpenAPI spec has two servers: production and testing. You could add a `--test` flag to select the second server.

```go
func main() {
	// ... init code ...

	// Add a `--test` flag to enable hitting testing.
	cli.AddFlag("test", "", "Use test endpoint", false)

	cli.PreRun = func(cmd *cobra.Command, args []string) error {
		if viper.GetBool("test") {
			// Use the test server
			viper.Set("server-index", 1)
		}
}
```

### Authentication

See the `auth0` module. More docs coming soon.

## Development

### Working with Templates

The code generator is configured to bundle all necessary assets into the final executable by default. If you wish to modify the templates, you can use the `go-bindata` tool to help:

```sh
# One-time setup of the go-bindata tool:
$ go get -u github.com/shuLhan/go-bindata/...

# Set up development mode (load data from actual files in ./templates/)
$ go-bindata -debug ./templates/...

# Now, do all your edits to the templates. You can test with:
$ go run *.go generate my-api.yaml

# Build the final static embedded files and code generator executable.
$ go generate
$ go install
```

## License

https://dgt.mit-license.org/
