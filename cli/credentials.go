package cli

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/h2non/gentleman.v2/context"
	"github.com/dgrijalva/jwt-go"
)

// AuthHandler describes a handler that can be called on a request to inject
// auth information and is agnostic to the type of auth.
type AuthHandler interface {
	// ProfileKeys returns the key names for fields to store in the profile.
	ProfileKeys() []string

	// OnRequest gets run before the request goes out on the wire.
	OnRequest(log *zerolog.Logger, request *http.Request) error
}

// AuthHandlers is the map of registered auth type names to handlers
var AuthHandlers = make(map[string]AuthHandler)

var authInitialized bool
var authCommand *cobra.Command
var authAddCommand *cobra.Command

// initAuth sets up basic commands and the credentials file so that new auth
// handlers can be registered. This is safe to call many times.
func initAuth() {
	if authInitialized {
		return
	}
	authInitialized = true

	// Set up the credentials file
	InitCredentials()

	// Add base auth commands
	authCommand = &cobra.Command{
		Use:   "auth",
		Short: "Authentication settings",
	}
	Root.AddCommand(authCommand)

	authAddCommand = &cobra.Command{
		Use:     "add-profile",
		Aliases: []string{"add"},
		Short:   "Add user profile for authentication",
	}
	authCommand.AddCommand(authAddCommand)

	authCommand.AddCommand(&cobra.Command{
		Use:     "list-profiles",
		Aliases: []string{"ls"},
		Short:   "List available configured authentication profiles",
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			profiles := Creds.Profiles

			if profiles != nil {
				// Use a map as a set to find the available auth type names.
				types := make(map[string]bool)
				for _, profile := range profiles {
					if profile.Info.Type != "" {
						types[profile.Info.Type] = true
					}
				}

				// For each type name, draw a table with the relevant profile keys
				for typeName := range types {
					handler := AuthHandlers[typeName]
					if handler == nil {
						continue
					}

					listKeys := handler.ProfileKeys()

					table := tablewriter.NewWriter(os.Stdout)
					table.SetHeader(append([]string{fmt.Sprintf("%s profile Name", typeName)}, listKeys...))

					for name, profile := range profiles {
						if profile.Info.Type == "" || profile.Info.Type != typeName {
							continue
						}

						row := []string{name}
						for _, key := range listKeys {
							row = append(row, profile.Info.Other[strings.Replace(key, "-", "_", -1)].(string))
						}
						table.Append(row)
					}
					table.Render()
				}
			} else {
				fmt.Printf("No profiles configured. Use `%s auth add-profile` to add one.\n", Root.CommandPath())
			}
		},
	})

	// Install auth middleware
	Client.UseRequest(func(ctx *context.Context, h context.Handler) {
		profile := GetActiveProfile().Info

		handler := AuthHandlers[profile.Type]
		if handler == nil {
			h.Error(ctx, fmt.Errorf("no handler for auth type %s", profile.Type))
			return
		}

		if err := handler.OnRequest(ctx.Get("log").(*zerolog.Logger), ctx.Request); err != nil {
			h.Error(ctx, err)
			return
		}

		h.Next(ctx)
	})
}

// UseAuth registers a new auth handler for a given type name. For backward-
// compatibility, the auth type name can be a blank string. It is recommended
// to always pass a value for the type name.
func UseAuth(typeName string, handler AuthHandler) {
	// Initialize auth system if it isn't already set up.
	initAuth()

	// Register the handler by its type.
	AuthHandlers[typeName] = handler

	// Set up the add-profile command.
	keys := handler.ProfileKeys()

	use := " [flags] <name>"
	for _, name := range keys {
		use += " <" + strings.Replace(name, "_", "-", -1) + ">"
	}

	run := func(cmd *cobra.Command, args []string) {
		name := strings.Replace(args[0], ".", "-", -1)
		Creds.Profiles[name] = Profile{
			Info:         ProfileInfo{
				Type:  typeName,
				Other: map[string]interface{}{},
			},
			TokenPayload: TokenPayload{},
		}

		for i, key := range keys {
			// Replace periods in the name since Viper will create nested structures
			// in the config and this isn't what we want!
			Creds.Profiles[name].Info.Other[strings.Replace(key, "-", "_", -1)] = args[i+1]
		}

		if err := Creds.Write(); err != nil {
			panic(err)
		}
	}

	if typeName == "" {
		// Backward-compatibility use-case without an explicit type. Set up the
		// `add-profile` command as the only way to authenticate.
		if authAddCommand.Run != nil {
			// This fallback code path was already used, so we must be registering
			// a *second* anonymous auth type, which is not allowed.
			panic("register auth type names to use multi-auth")
		}

		authAddCommand.Use = "add-profile" + use
		authAddCommand.Short = "Add a new named authentication profile"
		authAddCommand.Args = cobra.ExactArgs(1 + len(keys))
		authAddCommand.Run = run
	} else {
		// Add a new type-specific `add-profile` subcommand.
		authAddCommand.AddCommand(&cobra.Command{
			Use:   typeName + use,
			Short: "Add a new named " + typeName + " authentication profile",
			Args:  cobra.ExactArgs(1 + len(keys)),
			Run:   run,
		})
	}
}

type TokenPayload struct {
	ExpiresIn    int    `mapstructure:"expires_in"`
	RefreshToken string `mapstructure:"refresh_token"`
	AccessToken  string `mapstructure:"access_token"`
	IDToken      string `mapstructure:"id_token"`
	Scope        string `mapstructure:"scope"`
	TokenType string `mapstructure:"token_type"`
}

func (tp TokenPayload) ExpiresAt() time.Time {
	token, _, _ := new(jwt.Parser).ParseUnverified(tp.AccessToken, jwt.MapClaims{})
	claims, _ := token.Claims.(jwt.MapClaims)
	iat := claims["iat"].(int)
	return time.Unix(int64(iat), 0)
}

func (tp TokenPayload) Issuer() string {
	token, _, _ := new(jwt.Parser).ParseUnverified(tp.AccessToken, jwt.MapClaims{})
	claims, _ := token.Claims.(jwt.MapClaims)
	iss := claims["iss"].(string)
	return iss
}

type ProfileInfo struct {
	Type string `mapstructure:"type"`
	Other map[string]interface{} `mapstructure:,remain`
}

func (pi ProfileInfo) GetString(k string) string {
	value, _ := pi.Other[k]
	s, _ := value.(string)
	return s
}

func (pi ProfileInfo) ToMap() map[string]string {
	m := make(map[string]string)
	m["type"] = pi.Type
	for k, v := range pi.Other {
		if s, ok := v.(string); ok {
			m[k] = s
		}
	}
	return m
}

type Profile struct {
	Info         ProfileInfo `mapstructure:"info"`
	TokenPayload TokenPayload      `mapstructure:"token_payload"`
}

type Credentials struct {
	viper *viper.Viper
	Profiles map[string]Profile `mapstructure:"profiles"`
}

func (c *Credentials) Write() error {
	c.viper.Set("profiles", c.Profiles)
	return c.viper.WriteConfig()
}

func (c *Credentials) UpdateProfileTokenPayload(tokenType, accessToken, refreshToken string) error {
	tokenPayload := c.Profiles[RunConfig.GetProfileName()].TokenPayload
	tokenPayload.AccessToken = accessToken
	tokenPayload.TokenType = tokenType
	if refreshToken != "" {
		tokenPayload.RefreshToken = refreshToken
	}
	profile := c.Profiles[RunConfig.GetProfileName()]
	profile.TokenPayload = tokenPayload
	c.Profiles[RunConfig.GetProfileName()] = profile
	return c.Write()
}

// Creds represents a configuration file storing credential-related information.
// Use this only after `InitCredentials` has been called.
var Creds *Credentials

// GetActiveProfile returns the Profile for the currently configured profile.
func GetActiveProfile() Profile {
	return Creds.Profiles[RunConfig.GetProfileName()]
}

// InitCredentials sets up the creds file and `profile` global parameter.
func InitCredentials() {
	// Setup a credentials file, kept separate from configuration which might
	// get checked into source control.
	credConfig := viper.New()

	credConfig.SetConfigName("credentials")
	credConfig.AddConfigPath("$HOME/." + viper.GetString("app-name") + "/")
	Creds = &Credentials{}
	credConfig.Unmarshal(Creds)
	Creds.viper = credConfig
	if Creds.Profiles == nil {
		Creds.Profiles = make(map[string]Profile)
	}

	// Register a new `--profile` flag.
	AddGlobalFlag("profile", "", "Credentials profile to use for authentication", "default")
}
