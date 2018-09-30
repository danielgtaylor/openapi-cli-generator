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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	gentleman "gopkg.in/h2non/gentleman.v2"
)

// Root command (entrypoint) of the CLI.
var Root *cobra.Command

// Cache is used to store temporary data between runs.
var Cache *viper.Viper

// Log lets you write messages to the console.
var Log *zap.Logger

// Client makes HTTP requests and parses the responses.
var Client *gentleman.Client

// Formatter is the currently configured response output formatter.
var Formatter ResponseFormatter

// PreRun is a function that will run after flags are parsed but before the
// command handler has been called.
var PreRun func(cmd *cobra.Command, args []string) error

// Config is used to pass settings to the CLI.
type Config struct {
	AppName   string
	EnvPrefix string
	Version   string
}

// AddFlag will make a new global flag on the root command.
func AddFlag(name, short, description string, defaultValue interface{}) {
	viper.SetDefault(name, defaultValue)

	flags := Root.PersistentFlags()
	switch v := defaultValue.(type) {
	case bool:
		flags.BoolP(name, short, viper.GetBool(name), description)
	case int, int16, int32, int64, uint16, uint32, uint64:
		flags.IntP(name, short, viper.GetInt(name), description)
	case float32, float64:
		flags.Float64P(name, short, viper.GetFloat64(name), description)
	default:
		flags.StringP(name, short, fmt.Sprintf("%v", v), description)
	}
	viper.BindPFlag(name, flags.Lookup(name))
}

// Init will set up the CLI.
func Init(config *Config) {
	initConfig(config.AppName, config.EnvPrefix)
	initCache(config.AppName)

	// Determine if we are using a TTY or colored output is forced-on.
	tty := false
	if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) || viper.GetBool("color") {
		tty = true
	}

	logCfg := zap.NewDevelopmentConfig()
	logCfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	logCfg.Level = zap.NewAtomicLevelAt(zap.WarnLevel)

	if tty {
		logCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	var err error
	if Log, err = logCfg.Build(); err != nil {
		panic(err)
	}

	Client = gentleman.New()
	LogMiddleware()

	Formatter = NewDefaultFormatter(tty)

	Root = &cobra.Command{
		Use:     filepath.Base(os.Args[0]),
		Version: config.Version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if viper.GetBool("verbose") {
				logCfg.Level.SetLevel(zap.DebugLevel)

				settings := viper.AllSettings()

				// Hide any secret values
				for k := range settings {
					if strings.Contains(k, "secret") || strings.Contains(k, "password") {
						settings[k] = "**HIDDEN**"
					}
				}

				Log.Info(fmt.Sprintf("Configuration: %v", settings))
			}

			if PreRun != nil {
				if err := PreRun(cmd, args); err != nil {
					return err
				}
			}

			return nil
		},
	}

	AddFlag("verbose", "", "Enable verbose log output", false)
	AddFlag("output-format", "o", "Output format [json, yaml]", "json")
	AddFlag("query", "q", "Filter / project results using JMESPath", "")
	AddFlag("server", "", "Override server URL", "")
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
	viper.Set("executable", path.Base(os.Args[0]))
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
