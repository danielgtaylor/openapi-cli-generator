package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	isatty "github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	gentleman "gopkg.in/h2non/gentleman.v2"
)

// Root command (entrypoint) of the CLI.
var Root *cobra.Command

// Cache is used to store temporary data between runs.
var Cache *viper.Viper

// Client makes HTTP requests and parses the responses.
var Client *gentleman.Client

// Formatter is the currently configured response output formatter.
var Formatter ResponseFormatter

// PreRun is a function that will run after flags are parsed but before the
// command handler has been called.
var PreRun func(cmd *cobra.Command, args []string) error

var tty bool

// Config is used to pass settings to the CLI.
type Config struct {
	AppName   string
	EnvPrefix string
	Version   string
}

// Init will set up the CLI.
func Init(config *Config) {
	initConfig(config.AppName, config.EnvPrefix)
	initCache(config.AppName)

	// Determine if we are using a TTY or colored output is forced-on.
	tty = false
	if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) || viper.GetBool("color") {
		tty = true
	}

	zerolog.SetGlobalLevel(zerolog.WarnLevel)
	log.Logger = log.Output(ConsoleWriter{Out: os.Stderr, NoColor: !tty}).With().Caller().Logger()

	Client = gentleman.New()
	UserAgentMiddleware()
	LogMiddleware(tty)

	Formatter = NewDefaultFormatter(tty)

	Root = &cobra.Command{
		Use:     filepath.Base(os.Args[0]),
		Version: config.Version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if viper.GetBool("verbose") {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)

				settings := viper.AllSettings()

				// Hide any secret values
				for k := range settings {
					if strings.Contains(k, "secret") || strings.Contains(k, "password") {
						settings[k] = "**HIDDEN**"
					}
				}

				log.Info().Fields(settings).Msg("Configuration")
			}

			if PreRun != nil {
				if err := PreRun(cmd, args); err != nil {
					return err
				}
			}

			return nil
		},
	}

	Root.AddCommand(&cobra.Command{
		Use:   "help-input",
		Short: "Show CLI input help",
		Run:   showHelpInput,
	})

	AddGlobalFlag("verbose", "", "Enable verbose log output", false)
	AddGlobalFlag("output-format", "o", "Output format [json, yaml]", "json")
	AddGlobalFlag("query", "q", "Filter / project results using JMESPath", "")
	AddGlobalFlag("raw", "", "Output result of query as raw rather than an escaped JSON string", false)
	AddGlobalFlag("server", "", "Override server URL", "")
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

func initConfig(appName, envPrefix string) {
	// One-time setup to ensure the path exists so we can write files into it
	// later as needed.
	configDir := path.Join(userHomeDir(), "."+appName)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		panic(err)
	}

	// Load configuration from file(s) if provided.
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/" + appName + "/")
	viper.AddConfigPath("$HOME/." + appName + "/")
	viper.ReadInConfig()

	// Load configuration from the environment if provided. Flags below get
	// transformed automatically, e.g. `client-id` -> `PREFIX_CLIENT_ID`.
	viper.SetEnvPrefix(envPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	// Save a few things that will be useful elsewhere.
	viper.Set("app-name", appName)
	viper.Set("config-directory", configDir)
	viper.SetDefault("server-index", 0)
}

func initCache(appName string) {
	Cache = viper.New()
	Cache.SetConfigName("cache")
	Cache.AddConfigPath("$HOME/." + appName + "/")

	// Write a blank cache if no file is already there. Later you can use
	// cli.Cache.SaveConfig() to write new values.
	filename := path.Join(viper.GetString("config-directory"), "cache.json")
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		if err := ioutil.WriteFile(filename, []byte("{}"), 0600); err != nil {
			panic(err)
		}
	}

	Cache.ReadInConfig()
}

func showHelpInput(cmd *cobra.Command, args []string) {
	help := `# CLI Request Input

Input to the CLI is handled via parameters, arguments, and standard input. The help for an individual command shows the available optional parameters and required arguments. Optional parameters can be passed like ¬--option=value¬ or ¬--option value¬.

For requests that require a body, standard input and a CLI shorthand can complement each other to supply the request data.

## Standard Input

Standard input allows you to send in whatever data is required to make a successful request against the API. For example: ¬my-cli command <input.json¬ or ¬echo '{\"hello\": \"world\"}' | my-cli command¬.

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

See https://github.com/danielgtaylor/openapi-cli-generator/tree/master/shorthand#readme for more info.`

	fmt.Println(Markdown(strings.Replace(help, "¬", "`", -1)))
}
