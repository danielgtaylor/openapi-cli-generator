package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/alecthomas/chroma/quick"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gopkg.in/h2non/gentleman.v2"
	"gopkg.in/h2non/gentleman.v2/context"
)

// HTTP Client Errors
var (
	ErrCannotUnmarshal = errors.New("Unable to unmarshal response")
)

func indent(value string) string {
	trimmed := strings.TrimSuffix(value, "\x1b[0m")
	trimmed = strings.TrimRight(trimmed, "\n")
	return "  ╎ " + strings.Replace(trimmed, "\n", "\n  ╎ ", -1) + "\n"
}

// getBody returns and wraps the request/response body in a new reader, which
// is useful for logging purposes.
func getBody(r io.ReadCloser) (string, io.ReadCloser, error) {
	newReader := r

	body := ""
	if r != nil {
		data, err := ioutil.ReadAll(r)
		if err != nil {
			return "", nil, err
		}

		if len(data) > 0 {
			body = "\n" + string(data) + "\n"
			newReader = ioutil.NopCloser(bytes.NewReader(data))
		}
	}

	return body, newReader, nil
}

// UserAgentMiddleware sets the user-agent header on requests.
func UserAgentMiddleware() {
	Client.UseRequest(func(ctx *context.Context, h context.Handler) {
		ctx.Request.Header.Set("User-Agent", viper.GetString("app-name")+"-cli-"+Root.Version)
		h.Next(ctx)
	})
}

// LogMiddleware adds verbose log info to HTTP requests.
func LogMiddleware(useColor bool) {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	Client.UseRequest(func(ctx *context.Context, h context.Handler) {
		l := log.With().Str("request-id", fmt.Sprintf("%x", rnd.Uint64())).Logger()
		ctx.Set("log", &l)

		h.Next(ctx)
	})

	Client.UseHandler("before dial", func(ctx *context.Context, h context.Handler) {
		ctx.Set("start", time.Now())

		log := ctx.Get("log").(*zerolog.Logger)

		// Make the request body available to downstream processors through the
		// request context as `request-body`.
		body, newReader, err := getBody(ctx.Request.Body)
		if err != nil {
			h.Error(ctx, err)
			return
		}
		ctx.Set("request-body", body)
		ctx.Request.Body = newReader

		if viper.GetBool("verbose") {
			headers := ""
			for key, val := range ctx.Request.Header {
				headers += key + ": " + val[0] + "\n"
			}

			if body != "" {
				body = "\n" + body
			}

			http := fmt.Sprintf("%s %s %s\n%s%s", ctx.Request.Method, ctx.Request.URL, ctx.Request.Proto, headers, body)

			if useColor {
				sb := strings.Builder{}
				if err := quick.Highlight(&sb, http, "http", "terminal256", "cli-dark"); err != nil {
					h.Error(ctx, err)
				}
				http = sb.String()
			}

			log.Debug().Msgf("Making request:\n%s", indent(http))
		}

		h.Next(ctx)
	})

	Client.UseResponse(func(ctx *context.Context, h context.Handler) {
		l := ctx.Get("log").(*zerolog.Logger)

		if viper.GetBool("verbose") {
			headers := ""
			for key, val := range ctx.Response.Header {
				headers += key + ": " + val[0] + "\n"
			}

			body, newReader, err := getBody(ctx.Response.Body)
			if err != nil {
				h.Error(ctx, err)
				return
			}
			ctx.Response.Body = newReader

			http := fmt.Sprintf("%s %s\n%s\n%s", ctx.Response.Proto, ctx.Response.Status, headers, body)

			if useColor {
				sb := strings.Builder{}
				if err := quick.Highlight(&sb, http, "http", "terminal256", "cli-dark"); err != nil {
					h.Error(ctx, err)
				}
				http = sb.String()
			}

			l.Debug().Msgf("Got response in %s:\n%s", time.Since(ctx.Get("start").(time.Time)), indent(http))
		}

		h.Next(ctx)
	})
}

// UnmarshalRequest body into a given structure `s`. Supports both JSON and
// YAML depending on the request's content-type header.
func UnmarshalRequest(ctx *context.Context, s interface{}) error {
	return unmarshalBody(ctx.Request.Header, []byte(ctx.GetString("request-body")), s)
}

// UnmarshalResponse into a given structure `s`. Supports both JSON and
// YAML depending on the response's content-type header.
func UnmarshalResponse(resp *gentleman.Response, s interface{}) error {
	data := resp.Bytes()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d:\n%s", resp.StatusCode, string(data))
	}

	return unmarshalBody(resp.Header, data, s)
}

func unmarshalBody(headers http.Header, data []byte, s interface{}) error {
	if len(data) == 0 {
		return nil
	}

	ct := headers.Get("content-type")
	if strings.Contains(ct, "json") || strings.Contains(ct, "javascript") {
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
	} else if strings.Contains(ct, "yaml") {
		if err := yaml.Unmarshal(data, &s); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("Not sure how to unmarshal %s", ct)
	}

	return nil
}
