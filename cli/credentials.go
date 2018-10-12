package cli

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// CredentialsFile holds credential-related information.
type CredentialsFile struct {
	*viper.Viper
	keys     []string
	listKeys []string
}

// Creds represents a configuration file storing credential-related information.
// Use this only after `InitCredentials` has been called.
var Creds *CredentialsFile

// GetProfile returns the current profile's configuration.
func GetProfile() map[string]string {
	return Creds.GetStringMapString("profiles." + viper.GetString("profile"))
}

// ProfileKeys lets you specify authentication profile keys to be used in
// the credentials file.
func ProfileKeys(keys ...string) func(*CredentialsFile) error {
	return func(cf *CredentialsFile) error {
		cf.keys = keys
		return nil
	}
}

// ProfileListKeys sets which keys will be shown in the table when calling
// the `auth list-profiles` command.
func ProfileListKeys(keys ...string) func(*CredentialsFile) error {
	return func(cf *CredentialsFile) error {
		cf.listKeys = keys
		return nil
	}
}

// InitCredentials sets up the profile/auth commands. Must be called *after* you
// have called `cli.Init()`.
//
//  // Initialize an API key
//  cli.InitCredentials(cli.ProfileKeys("api-key"))
func InitCredentials(options ...func(*CredentialsFile) error) {
	// Setup a credentials file, kept separate from configuration which might
	// get checked into source control.
	Creds = &CredentialsFile{viper.New(), []string{}, []string{}}

	for _, option := range options {
		option(Creds)
	}

	Creds.SetConfigName("credentials")
	Creds.AddConfigPath("$HOME/." + viper.GetString("app-name") + "/")
	Creds.ReadInConfig()

	// Register a new `--profile` flag.
	AddGlobalFlag("profile", "", "Credentials profile to use for authentication", "default")

	// Register auth management commands to create and list profiles.
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication settings",
	}
	Root.AddCommand(cmd)

	use := "add-profile [flags] <name>"
	for _, name := range Creds.keys {
		use += " <" + strings.Replace(name, "_", "-", -1) + ">"
	}

	cmd.AddCommand(&cobra.Command{
		Use:     use,
		Aliases: []string{"add"},
		Short:   "Add a new named authentication profile",
		Args:    cobra.ExactArgs(1 + len(Creds.keys)),
		Run: func(cmd *cobra.Command, args []string) {
			for i, key := range Creds.keys {
				Creds.Set("profiles."+args[0]+"."+strings.Replace(key, "-", "_", -1), args[i+1])
			}

			filename := path.Join(viper.GetString("config-directory"), "credentials.json")
			if err := Creds.WriteConfigAs(filename); err != nil {
				panic(err)
			}
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:     "list-profiles",
		Aliases: []string{"ls"},
		Short:   "List available configured authentication profiles",
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			profiles := Creds.GetStringMap("profiles")
			if profiles != nil {
				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader(append([]string{"Profile Name"}, Creds.listKeys...))

				for name, profile := range profiles {
					row := []string{name}
					for _, key := range Creds.listKeys {
						row = append(row, profile.(map[string]interface{})[strings.Replace(key, "-", "_", -1)].(string))
					}
					table.Append(row)
				}
				table.Render()
			} else {
				fmt.Printf("No profiles configured. Use `%s auth add-profile` to add one.\n", Root.CommandPath())
			}
		},
	})
}
