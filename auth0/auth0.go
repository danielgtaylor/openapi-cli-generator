package auth0

import (
	"fmt"
	"time"

	"github.com/danielgtaylor/openapi-cli-generator/cli"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	gentleman "gopkg.in/h2non/gentleman.v2"
	"gopkg.in/h2non/gentleman.v2/context"
)

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
func getToken(log *zerolog.Logger, issuer string) (token string, err error) {
	log.Debug().Msgf("Using auth profile: %s", viper.GetString("profile"))

	if expires := cli.Cache.GetTime(expiresKey()); !expires.IsZero() {
		token = cli.Cache.GetString(tokenKey())

		if token == "" || time.Now().After(expires) {
			log.Debug().Msg("Token exipred")
			if token, err = refreshToken(log, issuer); err != nil {
				return
			}
		}
	} else {
		if token, err = refreshToken(log, issuer); err != nil {
			return
		}
	}

	return
}

// refreshToken fetches a new token from Auth0.
func refreshToken(log *zerolog.Logger, issuer string) (string, error) {
	log.Debug().Msg("Refreshing Auth0 token")

	profile := cli.GetProfile()

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
	standard := []string{"client-id", "client-secret", "audience"}

	cli.InitCredentials(
		cli.ProfileKeys(append(standard, extra...)...),
		cli.ProfileListKeys("client-id"))

	cli.Client.UseRequest(func(ctx *context.Context, h context.Handler) {
		if ctx.Request.Header.Get("Authorization") == "" {
			// No auth is set, so let's get the token either from a cache
			// or generate a new one from the issuing server.
			log := ctx.Get("log").(*zerolog.Logger)
			token, err := getToken(log, issuer)
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
