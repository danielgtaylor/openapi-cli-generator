// Package apikey provides authentication profile support for APIs that require
// a pre-generated constant authenticationn key passed via a header, query
// parameter, or cookie value in each request.
package apikey

import (
	"net/http"

	"github.com/rigetti/openapi-cli-generator/cli"
	"github.com/rs/zerolog"
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

// Handler sets up the API key authentication flow.
type Handler struct {
	Name string
	In   Location
	Keys []string
}

// ProfileKeys returns the key names for fields to store in the profile.
func (h *Handler) ProfileKeys() []string {
	return append([]string{apiKey}, h.Keys...)
}

// OnRequest gets run before the request goes out on the wire.
func (h *Handler) OnRequest(log *zerolog.Logger, request *http.Request) error {
	profile := cli.GetProfile()

	switch h.In {
	case LocationHeader:
		if request.Header.Get(h.Name) == "" {
			request.Header.Add(h.Name, profile[apiKey])
		}
	case LocationQuery:
		if request.URL.Query().Get(h.Name) == "" {
			query := request.URL.Query()
			query.Set(h.Name, profile[apiKey])
			request.URL.RawQuery = query.Encode()
		}
	case LocationCookie:
		if c, err := request.Cookie(h.Name); err != nil || c == nil {
			request.AddCookie(&http.Cookie{
				Name:  h.Name,
				Value: profile[apiKey],
			})
		}
	}

	return nil
}

// Init sets up the API key client authentication. Must be called *after* you
// have called `cli.Init()`. Passing `extra` values will set additional custom
// keys to store for each profile.
func Init(name string, in Location, extra ...string) {
	cli.UseAuth("", &Handler{
		Name: name,
		In:   in,
		Keys: extra,
	})
}
