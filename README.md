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
- Response filtering & projection by [JMESPath](http://jmespath.org/) plus [enhancements](https://github.com/danielgtaylor/go-jmespath-plus#enhancements)

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

<<<<<<< HEAD
| Name                | Description                                                        |
| ------------------- | ------------------------------------------------------------------ |
| `x-cli-aliases`     | Sets up command aliases for operations.                            |
| `x-cli-description` | Provide an alternate description for the CLI.                      |
| `x-cli-ignore`      | Ignore this path, operation, or parameter.                         |
| `x-cli-hidden`      | Hide this path, or operation.                                      |
| `x-cli-name`        | Provide an alternate name for the CLI.                             |
| `x-cli-waiters`     | Generate commands/params to wait until a certain state is reached. |
| `x-cli-cmd-groups`  | Describe CLI commands groups.                                      |

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

It is possible to exclude paths, operations, and/or parameters from the generated CLI. No code will be generated as they will be completely skipped.

```yaml
paths:
  /included:
    description: I will get included in the CLI.
  /excluded:
    x-cli-ignore: true
    description: I will not be in the CLI :-(
```

Alternatively, you can have the path or operation exist in the UI but be hidden from the standard help list. Specific help is still available via `my-cli my-hidden-operation --help`:

```yaml
paths:
  /hidden:
    x-cli-hidden: true
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

### Waiters

Waiters allow you to declaratively define special commands and parameters that will cause a command to block and wait until a particular condition has been met. This is particularly useful for asyncronous operations. For example, you might submit an order and then wait for that order to have been charged successfully before continuing on.

At a high level, waiters consist of an operation and a set of matchers that select a value and compare it to an expectation. For the example above, you might call the `GetOrder` operation every 30 seconds until the response's JSON `status` field is equal to `charged`. Here is what that would look like in your OpenAPI YAML file:

```yaml
info:
  title: Orders API
paths:
  /order/{id}:
    get:
      operationId: GetOrder
      description: Get an order's details.
      parameters:
        - name: id
          in: path
      responses:
        200:
          content:
            application/json:
              schema:
                type: object
                properties:
                  status:
                    type: string
                    enum: ['placed', 'charged', 'shipped', 'returned']
x-cli-waiters:
  order-charged:
    delay: 30
    attempts: 10
    operationId: GetOrder
    matchers:
      - select: response.body#status
        expected: charged
```

The generated CLI will work like this: `my-cli wait order-charged $ID` where `$ID` corresponds to the `GetOrder` operation's `id` parameter. It will try to get and match the status 10 times, with a pause of 30 seconds between tries. If it matches, it will exit with a zero status code. If it fails, it will exit with a non-zero exit code and log a message.

This is a great start, but we can make this a little bit friendlier to use. Take a look at this modified waiter configuration:

```yaml
x-cli-waiters:
  order-charged:
    short: Short description for CLI `--help`
    long: Long description for CLI `--help`
    delay: 30
    attempts: 10
    operationId: GetOrder
    matchers:
      - select: response.body#status
        expected: charged
      - select: response.status
        expected: 404
        state: failure
    after:
      CreateOrder:
        id: response.body#order_id
```

Here we added two new features:

1. A short-circuit to fail fast. If we type an invalid order ID then we want the command to exit immediately with a non-zero exit code.

2. The `after` block allows us to add a parameter to an _existing_ operation to invoke the waiter. This block says that after a call to `CreateOrder` with a `--wait-order-charged` param, it should call the waiter's `GetOrder` operation with the `id` param set to the result of the `response.body#order_id` selector.

You can now create and wait on an order via `my-cli create-order <order.json --wait-order-charged`.

#### Matchers

The following matcher fields are available:

| Field      | Description                                                                                     | Example           |
| ---------- | ----------------------------------------------------------------------------------------------- | ----------------- |
| `select`   | The value selection criteria. See the selector table below.                                     | `response.status` |
| `test`     | The test to perform. Defaults to `equal` but can be set to `any` and `all` to match list items. | `equal`           |
| `expected` | The expected value                                                                              | `charged`         |
| `state`    | The state to set. Defaults to `success` but can be set to `failure`.                            | `success`         |

The following selectors are available:

| Selector          | Description               | Argument       | Example                         |
| ----------------- | ------------------------- | -------------- | ------------------------------- |
| `request.param`   | Request parameter         | Parameter name | `request.param#id`              |
| `request.body`    | Request body query        | JMESPath query | `request.body#order.id`         |
| `response.status` | Response HTTP status code | -              | `response.status`               |
| `response.header` | Response HTTP header      | Header name    | `response.header#content-type`  |
| `response.body`   | Response body query       | JMESPath query | `response.body#orders[].status` |

### Commands Grouping

By default all CLI commands are exposed at the root level, however it is possible to group contextually related
commands together. It works by defining a `x-cli-cmd-groups` section at the root of the API spec to describe the
groups structure. Optionally, groups can be nested under another group using the `parent` keyword in order to create
a hierarchy.

```yaml
# cli.yaml
x-cli-cmd-groups:
  blog:
    short: Manage blogs
    long: The commands in this group let you manage your blogs
  post:
    short: Manage posts
    long: The commands in this group let you manage your blog posts
    parent: blog
  media:
    short: Manage media
    long: The commands in this group let you manage your media files
    parent: blog
  user:
    short: Manage users
    long: The commands in this group let you manage your users
```

Once you have defined your command groups, add a `x-cli-cmd-group` entry at operation level to assign an operation to
a group:

```yaml
# cli.yaml
paths:
  /blog:
    post:
      operationId: create-blog
      description: Create a blog
      x-cli-cmd-group: blog
    get:
      operationId: list-blogs
      description: List blogs
      x-cli-cmd-group: blog
  /blog/{blogid}:
    put:
      operationId: update-blog
      description: Update a blog
      x-cli-cmd-group: blog
    delete:
      operationId: delete-blog
      description: Delete a blog
      x-cli-cmd-group: blog
  /blog/{blogid}/post:
    get:
      operationId: list-post
      description: List posts
      x-cli-cmd-group: post
    post:
      operationId: create-post
      description: Create a post
      x-cli-cmd-group: post
  /blog/{blogid}/post/{postid}:
    get:
      operationId: show-post
      description: Show a post
      x-cli-cmd-group: post
    put:
      operationId: update-post
      description: Update a post
      x-cli-cmd-group: post
    delete:
      operationId: delete-post
      description: Delete a post
      x-cli-cmd-group: post
  /blog/{blogid}/media:
    get:
      operationId: list-media
      description: List media files
      x-cli-cmd-group: media
    post:
      operationId: upload-media
      description: Upload a media file
      x-cli-cmd-group: media
  /blog/{blogid}/media/{mediaid}:
    get:
      operationId: show-media
      description: Show a media file
      x-cli-cmd-group: media
    delete:
      operationId: delete-media
      description: Delete a media file
      x-cli-cmd-group: media
  /user:
    post:
      operationId: create-user
      description: Create a user
      x-cli-cmd-group: user
    get:
      operationId: list-user
      description: List users
      x-cli-cmd-group: user
  /user/{userid}:
    get:
      operationId: show-user
      description: Show a user
      x-cli-cmd-group: user
    delete:
      operationId: delete-user
      description: Delete a user
      x-cli-cmd-group: user
```

This sample configuration will result in the following CLI commands hierarchy:

```
$ ./cli -h
Usage:
  cli [command]

Available Commands:
  blog        Manage blogs
  help        Help about any command
  help-config Show CLI configuration help
  help-input  Show CLI input help
  user        Manage users

$ ./cli blog -h
The commands in this group let you manage your blogs

Usage:
  cli blog [command]

Available Commands:
  create      Create a blog
  delete      Delete a blog
  list        List blogs
  media       Manage media
  post        Manage posts
  update      Update a blog

$ ./cli blog post -h
The commands in this group let you manage your blog posts

Usage:
  cli blog post [command]

Available Commands:
  create      Create a post
  delete      Delete a post
  list        List posts
  show        Show a post
  update      Update a post

# etc.
```

Without grouping, the resulting commands look like this:

```
$ ./cli
Usage:
  cli [command]

Available Commands:
  create-blog  Create a blog
  create-post  Create a post
  create-user  Create a user
  delete-blog  Delete a blog
  delete-media Delete a media file
  delete-post  Delete a post
  delete-user  Delete a user
  help         Help about any command
  help-config  Show CLI configuration help
  help-input   Show CLI input help
  list-blog    List blogs
  list-media   List media files
  list-post    List posts
  list-user    List users
  show-media   Show a media file
  show-post    Show a post
  update-blog  Update a blog
  update-post  Update a post
  upload-media Upload a media file
```

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

### Authentication & Authorization

See the `apikey` module for a simple example of a pre-shared key.

If instead you use a third party auth system that vends tokens and want your users to be able to log in and use the API, here's an example using Auth0:

```go
func main() {
	cli.Init(&cli.Config{
		AppName:   "example",
		EnvPrefix: "EXAMPLE",
		Version:   "1.0.0",
	})

  // Auth0 requires a client ID, issuer base URL, and audience fields when
  // requesting a token. We set these up here and use the Authorization Code
  // with PKCE flow to log in, which opens a browser for the user to log into.
  clientID := "abc123"
  issuer := "https://mycompany.auth0.com/"

  cli.UseAuth("user", &oauth.AuthCodeHandler{
    ClientID: "clientID",
    AuthorizeURL: issuer+"authorize",
    TokenURL: issuer+"oauth/token",
    Keys: []string{"audience"},
    Params: []string{"audience"},
    Scopes: []string{"offline_access"},
  })

  // TODO: Register API commands here
  // ...

	cli.Root.Execute()
}
```

Note that there is a convenience module when using Auth0 specifically, allowing you to do this:

```go
auth0.InitAuthCode(clientID, issuer,
  auth0.Type("user"),
  auth0.Scopes("offline_access"))
```

The expanded example above is more useful when integrating with other services since it uses basic OAuth 2 primitives.

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
