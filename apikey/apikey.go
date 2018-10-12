// Package apikey provides authentication profile support for APIs that require
// a pre-generated constant authenticationn key passed via a header, query
// parameter, or cookie value in each request.
package apikey

import (
	"net/http"

	"github.com/danielgtaylor/openapi-cli-generator/cli"
	"gopkg.in/h2non/gentleman.v2/context"
)

// Location defines how a parameter is sent.
type Location int

// API key parameter locations, which correspond to the OpenAPI `in` parameter
// values for the `apikey` security type.
const (
	LocationHeader Location = iota
	LocationQuery
	LocationCookie
)

const apiKey = "api_key"

// Init sets up the API key client authentication. Must be called *after* you
// have called `cli.Init()`. Passing `extra` values will set additional custom
// keys to store for each profile.
func Init(name string, in Location, extra ...string) {
	cli.InitCredentials(
		cli.ProfileKeys(append([]string{apiKey}, extra...)...),
		cli.ProfileListKeys(apiKey),
	)

	cli.Client.UseRequest(func(ctx *context.Context, h context.Handler) {
		profile := cli.GetProfile()

		switch in {
		case LocationHeader:
			if ctx.Request.Header.Get(name) == "" {
				ctx.Request.Header.Add(name, profile[apiKey])
			}
		case LocationQuery:
			if ctx.Request.URL.Query().Get(name) == "" {
				query := ctx.Request.URL.Query()
				query.Set(name, profile[apiKey])
				ctx.Request.URL.RawQuery = query.Encode()
			}
		case LocationCookie:
			if c, err := ctx.Request.Cookie(name); err != nil || c == nil {
				ctx.Request.AddCookie(&http.Cookie{
					Name:  name,
					Value: profile[apiKey],
				})
			}
		}

		h.Next(ctx)
	})
}
