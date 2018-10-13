# OpenAPI CLI Generator

[![GoDoc](https://godoc.org/github.com/danielgtaylor/openapi-cli-generator?status.svg)](https://godoc.org/github.com/danielgtaylor/openapi-cli-generator)
[![Build Status](https://travis-ci.org/danielgtaylor/openapi-cli-generator.svg?branch=master)](https://travis-ci.org/danielgtaylor/openapi-cli-generator)
[![Go Report Card](https://goreportcard.com/badge/github.com/danielgtaylor/openapi-cli-generator)](https://goreportcard.com/report/github.com/danielgtaylor/openapi-cli-generator)
[![Platforms](https://img.shields.io/badge/platform-win%20%7C%20mac%20%7C%20linux-ligh.svg)](https://github.com/danielgtaylor/openapi-cli-generator/releases)

<img alt="openapi-to-cli" src="https://user-images.githubusercontent.com/106826/46594546-a8bb2480-ca88-11e8-90ec-fb87e51009a8.png">

This project can be used to generate CLIs from OpenAPI 3 specs. The generated CLIs have the following features:

- Authentication support for API keys and [Auth0](https://auth0.com/).
- Commands, subcommands, & flag parsing through [Cobra](https://github.com/spf13/cobra)
- Configuration through [Viper](https://github.com/spf13/viper)
  - JSON, YAML, or TOML config files in `/etc/` and `$HOME`, e.g. `{"verbose": true}` in `~/.my-app/config.json`
  - From environment: `APP_NAME_VERBOSE=1`
  - From flags: `--verbose`
- HTTP middleware through [Gentleman](https://github.com/h2non/gentleman/)
- Command middleware with custom parameters (see customization below)
- Input through `stdin` or [CLI shorthand](https://github.com/danielgtaylor/openapi-cli-generator/tree/master/shorthand)
- Built-in cache to save data between runs
- Fast structured logging via [zerolog](https://github.com/rs/zerolog)
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
`x-cli-name` | Provide an alternate name for the CLI.

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

### Name

You can override the default name for the API, operations, and params:

```yaml
info:
  x-cli-name: foo
paths:
  /items:
    operationId: myOperation
    x-cli-name: my-op
    parameters:
    - name: id
      x-cli-name: item-id
      in: query
```

With the above, you would be able to call `my-cli my-op --item-id=12`.

## Customization

Your `main.go` is the entrypoint to your generated CLI, and may be customized to add additional logic and features. For example, you might set custom headers or handle auth before a request goes out on the wire. The `apikey` module provides a sample implementation.

### Configuration Description

TODO: Show table describing all well-known configuration keys.

### Custom Global Flags

It's possible to supply custom flags and a pre-run function. For example, say your OpenAPI spec has two servers: production and testing. You could add a `--test` flag to select the second server.

```go
func main() {
	// ... init code ...

	// Add a `--test` flag to enable hitting testing.
	cli.AddGlobalFlag("test", "", "Use test endpoint", false)

	cli.PreRun = func(cmd *cobra.Command, args []string) error {
		if viper.GetBool("test") {
			// Use the test server
			viper.Set("server-index", 1)
		}
}
```

### HTTP Middleware

[Gentleman](https://github.com/h2non/gentleman/) provides support for HTTP request and response middleware. Don't forget to call `h.Next(ctx)` in your handler! For example:

```go
// Register a request middleware handler to print the path.
cli.Client.UseRequest(func(ctx *context.Context, h context.Handler) {
	fmt.Printf("Request path: %s\n", ctx.Request.URL.Path)
	h.Next(ctx)
})

// Register a response middleware handler to print the status code.
cli.Client.UseResponse(func(ctx *context.Context, h context.Handler) {
	fmt.Printf("Response status: %d\n", ctx.Response.StatusCode)
	h.Next(ctx)
})
```

### Custom Command Flags & Middleware

While the above HTTP middleware is great for adding headers or logging various things, there are times when you need to modify the behavior of a generated command. You can do so by registering custom command flags and using command middleware.

Flags and middleware are applied to a _command path_, which is a space-separated list of commands in a hierarchy. For example, if you have a command `foo` which has a subcommand `bar`, then the command path to reference `bar` for flags and middleware would be `foo bar`.

Note that any calls to `cli.AddFlag` must be made **before** calling the generated command registration function (e.g. `openapiRegister(...)`) or the flags will not get created properly.

Here's an example showing how a custom flag can change the command response:

```go
// Register a new custom flag for the `foo` command.
cli.AddFlag("foo", "custom", "", "description", "")

cli.RegisterAfter("foo", func(cmd *cobra.Command, params *viper.Viper, resp *gentleman.Response, data interface{}) interface{} {
  m := data.(map[string]interface{})
  m["custom"] = params.GetString("custom")
  return m
})

// Register our generated commands with the CLI after the above.
openapiRegister(false)
```

If the `foo` command would normally return a JSON object like `{"hello": "world"}` it would now return the following if called with `--custom=test`:

```json
{
  "custom": "test",
  "hello": "world"
}
```

### Authentication

See the `apikey` module for an example. More docs coming soon.

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
