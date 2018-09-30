package auth0

import (
	"fmt"
	"path"
	"time"

	"github.com/danielgtaylor/openapi-cli-generator/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	gentleman "gopkg.in/h2non/gentleman.v2"
	"gopkg.in/h2non/gentleman.v2/context"
)

// Creds represent a configuration file storing credential-related information.
var Creds *viper.Viper

type auth0Token struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

func tokenKey() string {
	return "profiles." + viper.GetString("profile") + ".token"
}

func expiresKey() string {
	return "profiles." + viper.GetString("profile") + ".expires"
}

// getToken tries to get the token from the cache first. If it's not in the
// cache or the token has expired, we need to refresh it.
func getToken(issuer string) (token string, err error) {
	cli.Log.Debug("Using auth profile: " + viper.GetString("profile"))

	if expires := cli.Cache.GetTime(expiresKey()); !expires.IsZero() {
		token = cli.Cache.GetString(tokenKey())

		if token == "" || time.Now().After(expires) {
			cli.Log.Debug("Token exipred")
			if token, err = refreshToken(issuer); err != nil {
				return
			}
		}
	} else {
		if token, err = refreshToken(issuer); err != nil {
			return
		}
	}

	return
}

// refreshToken fetches a new token from Auth0.
func refreshToken(issuer string) (string, error) {
	cli.Log.Debug("Refreshing Auth0 token")

	profile := GetProfile()

	if len(profile) == 0 {
		return "", fmt.Errorf("Cannot find profile: %s", viper.GetString("profile"))
	}

	data := fmt.Sprintf(`{
		"client_id": "%s",
		"client_secret": "%s",
		"audience": "%s",
		"grant_type": "client_credentials"}`, profile["client_id"], profile["client_secret"], profile["audience"])

	resp, err := gentleman.New().Post().URL(issuer+"oauth/token").
		AddHeader("Content-Type", "application/json").
		BodyString(data).Do()
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf(string(resp.Bytes()))
	}

	var token auth0Token
	if err := resp.JSON(&token); err != nil {
		return "", err
	}

	// Set the token info into the cache for the current profile.
	cli.Cache.Set(tokenKey(), token.AccessToken)

	expires := time.Duration(token.ExpiresIn) * time.Second
	cli.Cache.Set(expiresKey(), time.Now().Add(expires))

	// Save the cache to disk.
	if err := cli.Cache.WriteConfig(); err != nil {
		panic(err)
	}

	return token.AccessToken, nil
}

// Init sets up the Auth0 client authentication. Must be called *after* you
// have called `cli.Init()`. Pass in profile-related extra variables to store
// them alongside the default profile information.
func Init(issuer string, extra ...string) {
	// Setup a credentials file, kept separate from configuration which might
	// get checked into source control.
	Creds = viper.New()
	Creds.SetConfigName("credentials")
	Creds.AddConfigPath("$HOME/." + viper.GetString("app-name") + "/")
	Creds.ReadInConfig()

	// Register a new `--profile` flag.
	cli.AddFlag("profile", "", "Credentials profile to use for auth", "default")

	// Register auth management commands to create and list profiles.
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication settings",
	}
	cli.Root.AddCommand(cmd)

	use := "add-profile [flags] <name> <client-id> <client-secret> <audience>"
	for _, name := range extra {
		use += " <" + name + ">"
	}

	cmd.AddCommand(&cobra.Command{
		Use:     use,
		Aliases: []string{"add"},
		Short:   "Add a new named Auth0 profile",
		Args:    cobra.ExactArgs(4 + len(extra)),
		Run: func(cmd *cobra.Command, args []string) {
			Creds.Set("profiles."+args[0]+".client_id", args[1])
			Creds.Set("profiles."+args[0]+".client_secret", args[2])
			Creds.Set("profiles."+args[0]+".audience", args[3])

			for i, name := range extra {
				Creds.Set("profiles."+args[0]+"."+name, args[3+i])
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
		Short:   "List available configured auth profiles",
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			profiles := Creds.GetStringMap("profiles")
			if profiles != nil {
				fmt.Println("Profile (Client ID)")
				for name, profile := range profiles {
					fmt.Printf("%s (%s)\n", name, profile.(map[string]interface{})["client_id"].(string))
				}
			} else {
				fmt.Printf("No profiles configured. Use `%s auth add-profile` to add one.\n", viper.GetString("executable"))
			}
		},
	})

	cli.Client.UseRequest(func(ctx *context.Context, h context.Handler) {
		if ctx.Request.Header.Get("Authorization") == "" {
			// No auth is set, so let's get the token either from a cache
			// or generate a new one from the issuing server.
			token, err := getToken(issuer)
			if err != nil {
				h.Error(ctx, err)
				return
			}

			ctx.Request.Header.Add("Authorization", fmt.Sprintf("bearer %s", token))
		}

		h.Next(ctx)
	})

	// TODO: retry on 401
	// cli.Client.UseResponse(func(ctx *context.Context, h context.Handler) {
	// 	h.Next(ctx)
	// })
}

// GetProfile returns the current profile's configuration.
func GetProfile() map[string]string {
	return Creds.GetStringMapString("profiles." + viper.GetString("profile"))
}
