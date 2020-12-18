package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	colorable "github.com/mattn/go-colorable"
	isatty "github.com/mattn/go-isatty"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	gentleman "gopkg.in/h2non/gentleman.v2"
)

// Root command (entrypoint) of the CLI.
var Root *cobra.Command

// Client makes HTTP requests and parses the responses.
var Client *gentleman.Client

// Formatter is the currently configured response output formatter.
var Formatter ResponseFormatter

// PreRun is a function that will run after flags are parsed but before the
// command handler has been called.
var PreRun func(cmd *cobra.Command, args []string) error

// Stdout is a cross-platform, color-safe writer if colors are enabled,
// otherwise it defaults to `os.Stdout`.
var Stdout io.Writer = os.Stdout

// Stderr is a cross-platform, color-safe writer if colors are enabled,
// otherwise it defaults to `os.Stderr`.
var Stderr io.Writer = os.Stderr

var tty bool

// Config is used to pass settings to the CLI.
type Config struct {
	AppName   string
	EnvPrefix string
	Version   string
}

// Init will set up the CLI.
func Init(config *Config) {
	// Determine if we are using a TTY or colored output is forced-on.
	tty = false
	if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) || viper.GetBool("color") {
		tty = true
	}

	if viper.GetBool("nocolor") {
		// If forced off, ignore all of the above!
		tty = false
	}

	if tty {
		// Support colored output across operating systems.
		Stdout = colorable.NewColorableStdout()
		Stderr = colorable.NewColorableStderr()
	}

	log.Logger = log.Output(ConsoleWriter{Out: Stderr, NoColor: !tty}).With().Caller().Logger()

	Client = gentleman.New()
	UserAgentMiddleware()
	LogMiddleware(tty)

	Formatter = NewDefaultFormatter(tty)

	Root = &cobra.Command{
		Use:     filepath.Base(os.Args[0]),
		Version: config.Version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if PreRun != nil {
				if err := PreRun(cmd, args); err != nil {
					return err
				}
			}

			return nil
		},
	}

	Root.SetOutput(Stdout)

	Root.AddCommand(&cobra.Command{
		Use:   "help-config",
		Short: "Show CLI configuration help",
		Run:   showHelpConfig,
	})

	Root.AddCommand(&cobra.Command{
		Use:   "help-input",
		Short: "Show CLI input help",
		Run:   showHelpInput,
	})

}

func userHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

var RunConfig ClientConfiguration


func showHelpConfig(cmd *cobra.Command, args []string) {
	help := `# CLI ClientConfiguration

ClientConfiguration for the CLI comes from the following places:

1. Command options
2. Environment variables
3. ClientConfiguration files

## Global Command Options

Command options are passed when invoking the command. For example, ¬--verbose¬ configures the CLI to run with additional output for debugging. Using the top level ¬--help¬ to shows a list of available options:

$flags

## Environment Variables

Environment variables must be capitalized, prefixed with ¬$APP¬, and words are separated by an underscore rather than a dash. For example, setting ¬$APP_VERBOSE=1¬ is equivalent to passing ¬--verbose¬ to the command.

## ClientConfiguration Files

ClientConfiguration files can be used to configure the CLI and can be written using JSON, YAML, or TOML. The CLI searches in your home directory first (e.g. ¬$config-dir/config.json¬) and on Mac/Linux also looks in e.g. ¬/etc/$app/config.json¬. The following is equivalent to passing ¬--verbose¬ to the command:

¬¬¬json
{
  "verbose": true
}
¬¬¬

## Special Cases

Some configuration values are not exposed as command options but can be set via prefixed environment variables or in configuration files. They are documented here.

AuthServerName      | Type   | Description
--------- | ------ | -----------
¬color¬   | ¬bool¬ | Force colorized output.
¬nocolor¬ | ¬bool¬ | Disable colorized output.
`

	help = strings.Replace(help, "¬", "`", -1)
	help = strings.Replace(help, "$APP", strings.ToUpper(viper.GetString("app-name")), -1)
	help = strings.Replace(help, "$app", viper.GetString("app-name"), -1)
	help = strings.Replace(help, "$config-dir", viper.GetString("config-directory"), -1)

	flags := make([]string, 0)
	flags = append(flags, "AuthServerName            | Type     | Description")
	flags = append(flags, "--------------- | -------- | -----------")
	Root.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		flags = append(flags, fmt.Sprintf("%-15s", "`"+f.Name+"`")+" | `"+fmt.Sprintf("%-7s", f.Value.Type()+"`")+" | "+f.Usage)
	})

	help = strings.Replace(help, "$flags", strings.Join(flags, "\n"), -1)

	fmt.Fprintln(Stdout, Markdown(help))
}

func showHelpInput(cmd *cobra.Command, args []string) {
	help := `# CLI Request Input

Input to the CLI is handled via parameters, arguments, and standard input. The help for an individual command shows the available optional parameters and required arguments. Optional parameters can be passed like ¬--option=value¬ or ¬--option value¬.

For requests that require a body, standard input and a CLI shorthand can complement each other to supply the request data.

## Standard Input

Standard input allows you to send in whatever data is required to make a successful request against the API. For example: ¬my-cli command <input.json¬ or ¬echo '{\"hello\": \"world\"}' | my-cli command¬.

Note: Windows PowerShell and other shells that do not support input redirection via ¬<¬ will need to pipe input instead, for example: ¬cat input.json | my-cli command¬. This may load the entire input file into memory.

## CLI Shortand Syntax

Any arguments beyond those that are required for a command are treated as CLI shorthand and used to generate structured data for requests. Shorthand objects are specified as key/value pairs. They complement standard input so can be used to override or to add additional fields as needed. For example: ¬my-cli command <input.json field: value, other: value2¬.

Null, booleans, integers, and floats are automatically coerced into the appropriate type. Use the ¬~¬ modifier after the ¬:¬ to force a string, like ¬field:~ true¬.

Nested objects use a ¬.¬ separator. Properties can be grouped inside of ¬{¬ and ¬}¬. For example, ¬foo.bar{id: 1, count: 5}¬ will become:

¬¬¬json
{
  "foo": {
    "bar": {
      "id": 1,
      "count": 5
    }
  }
}
¬¬¬

Simple scalar arrays use a ¬,¬ to separate values, like ¬key: 1, 2, 3¬. Appending to an array is possible like ¬key[]: 1, key[]: 2, key[]: 3¬. For nested arrays you specify multiple square bracket sets like ¬key[][]: value¬. You can directly reference an index by including one like ¬key[2]: value¬.

Both objects and arrays can use backreferences. An object backref starts with a ¬.¬ and an array backref starts with ¬[¬. For example, ¬foo{id: 1, count: 5}¬ can be rewritten as ¬foo.id: 1, .count: 5¬.

Use an ¬@¬ to load the contents of a file as the value, like ¬key: @filename¬. Use the ¬~¬ modifier to disable this behavior: ¬key:~ @user¬. By default structured data is loaded when recognized. Use the ¬~¬ filename modifier to force a string: ¬key: @~filename¬. Use the ¬%¬ modifier to load as base-64 data: ¬key: @%filename¬.

See https://github.com/rigetti/openapi-cli-generator/tree/master/shorthand#readme for more info.`

	fmt.Fprintln(Stdout, Markdown(strings.Replace(help, "¬", "`", -1)))
}
