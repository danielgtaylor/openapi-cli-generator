package cli

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"regexp"
	"strings"
)

func getBindPathsFromProfile(flagName string, args ...interface{}) (paths []string, err error) {
	if len(args) != 1 {
		err = fmt.Errorf("received unexpected arguments (%d) for flag %q", len(args), ToSnakeCase(flagName))
		return
	}
	settings, ok := args[0].(Settings)
	if !ok {
		err = fmt.Errorf("received non-settings arguments for flag %q", flagName)
		return
	}
	paths = make([]string, 0)
	for profileName := range settings.Profiles {
		paths = append(paths, fmt.Sprintf("profiles.%s.%s", profileName, ToSnakeCase(flagName)))
	}
	return
}

type GlobalFlag struct {
	*pflag.Flag
	UseDefault         bool
	customGetBindPaths func(flagName string, args ...interface{}) ([]string, error)
}

var viperBindPaths = map[string]string{
	"profileName": "default_profile_name",
}

func MakeAndParseGlobalFlags() (globalFlags []GlobalFlag, err error) {
	flagSet := pflag.NewFlagSet("global", pflag.ContinueOnError)
	flagSet.ParseErrorsWhitelist = pflag.ParseErrorsWhitelist{
		UnknownFlags: true,
	}

	flagSet.String("profileName", "default", "")
	flagSet.String("authServerName", "default", "")
	flagSet.String("credentialsName", "default", "")
	flagSet.String("apiUrl", "https://http", "")
	flagSet.StringP("outputFormat", "o", "json", "Output format [json, yaml]")
	flagSet.BoolP("help", "h", false, "")
	flagSet.Bool("raw", false, "Output result of query as raw rather than an escaped JSON string or list")
	err = flagSet.Parse(os.Args[1:])
	if err != nil {
		return
	}

	globalFlags = []GlobalFlag{
		{
			UseDefault: true,
			Flag:       flagSet.Lookup("profileName"),
		}, {
			UseDefault:         true,
			customGetBindPaths: getBindPathsFromProfile,
			Flag:               flagSet.Lookup("authServerName"),
		}, {
			UseDefault:         true,
			customGetBindPaths: getBindPathsFromProfile,
			Flag:               flagSet.Lookup("credentialsName"),
		}, {
			UseDefault:         true,
			customGetBindPaths: getBindPathsFromProfile,
			Flag:               flagSet.Lookup("apiUrl"),
		}, {
			UseDefault: true,
			customGetBindPaths: func(flagName string, args ...interface{}) ([]string, error) {
				return getBindPathsFromProfile("applications.cli.output_format", args...)
			},
			Flag: flagSet.Lookup("outputFormat"),
		},
		/*{
			customGetBindPaths: func(flagName string, args ...interface{}) ([]string, error) {
				return getBindPathsFromProfile("applications.cli.query", args...)
			},
			Flag: &pflag.Flag{
				Name:      "query",
				Shorthand: "q",
				Usage:     "Filter / project results using JMESPath",
			},
		},*/
		{
			customGetBindPaths: func(flagName string, args ...interface{}) ([]string, error) {
				return getBindPathsFromProfile("applications.cli.raw", args...)
			},
			Flag: flagSet.Lookup("raw"),
		},
	}
	return
}

func (gf GlobalFlag) getBindPaths(args ...interface{}) (paths []string, err error) {
	if gf.customGetBindPaths != nil {
		return gf.customGetBindPaths(gf.Flag.Name, args...)
	}
	paths = make([]string, 0)
	if viperBindPath, ok := viperBindPaths[gf.Flag.Name]; ok {
		paths = append(paths, viperBindPath)
		return
	}
	paths = append(paths, ToSnakeCase(gf.Flag.Name))
	return
}

func (gf GlobalFlag) bindFlag(v *viper.Viper, args ...interface{}) (err error) {
	paths, err := gf.getBindPaths(args...)
	if err != nil {
		return
	}
	for _, p := range paths {
		err = gf.bindFlagTo(v, p)
		if err != nil {
			return
		}
	}
	return
}

func (gf GlobalFlag) bindFlagTo(v *viper.Viper, viperBindPath string) error {
	if gf.Flag.Changed {
		return v.BindPFlag(viperBindPath, gf.Flag)
	}

	if gf.UseDefault && !v.IsSet(viperBindPath) {
		return v.BindPFlag(viperBindPath, gf.Flag)
	}
	return nil
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
