# OpenAPI CLI Generator

This project can be used to generate CLIs from OpenAPI 3 specs. The generated CLIs have the following features:

- Configuration through [Viper](https://github.com/spf13/viper)
  - JSON, YAML, or TOML config files in `/etc/` and `$HOME`, e.g. `{"verbose": true}` in `~/.my-app/config.json`
  - From environment: `APP_NAME_VERBOSE=1`
  - From flags: `--verbose`
- HTTP middleware through [Gentleman](https://github.com/h2non/gentleman/)
- Input through `stdin` or [CLI shorthand](https://github.com/danielgtaylor/openapi-cli-generator/tree/master/shorthand)
- Built-in cache to save data between runs
- Fast structured logging via [Zap](https://github.com/uber-go/zap)
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
$ openapi-cli-generator init <app-name> | gofmt | goimports >main.go

# Generate the commands
$ openapi-cli-generator generate openapi.yaml | gofmt | goimports >commands.go
```

Last, add a line like the following to your `main.go` file:

```go
myApiRegister(false)
```

If you would like to generate a client for many APIs and have each available under their own namespace, pass `true` instead. Next, build your client:

```sh
# Build & install the generated client.
$ go install

# Test it out!
$ my-cli --help
```

## Customization

Your `main.go` is the entrypoint to your generated CLI, and may be customized to add additional logic and features. For example, you might set custom headers or handle auth before a request goes out on the wire. The `auth0` module provides a sample implementation.

### Configuration Description

TODO: Show table describing all well-known configuration keys.

### Custom Flags

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

## License

https://dgt.mit-license.org/
