package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gopkg.in/h2non/gentleman.v2"
	"gopkg.in/h2non/gentleman.v2/context"
)

// HTTP Client Errors
var (
	ErrCannotUnmarshal = errors.New("Unable to unmarshal response")
)

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

// LogMiddleware adds verbose log info to HTTP requests.
func LogMiddleware() {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	Client.UseHandler("before dial", func(ctx *context.Context, h context.Handler) {
		ctx.Set("start", time.Now())

		l := Log.With(
			zap.String("request-id", fmt.Sprintf("%x", rnd.Uint64())),
		)

		if viper.GetBool("verbose") {
			headers := ""
			for key, val := range ctx.Request.Header {
				headers += key + ": " + val[0] + "\n"
			}

			body, newReader, err := getBody(ctx.Request.Body)
			if err != nil {
				h.Error(ctx, err)
				return
			}
			ctx.Request.Body = newReader

			l.Debug(fmt.Sprintf("Making request:\n%s %s %s\n%s%s", ctx.Request.Method, ctx.Request.URL, ctx.Request.Proto, headers, body))
		}

		ctx.Set("log", l)

		h.Next(ctx)
	})

	Client.UseResponse(func(ctx *context.Context, h context.Handler) {
		l := ctx.Get("log").(*zap.Logger)

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

			l.Debug(fmt.Sprintf("Got response in %s:\n%s %s\n%s%s", time.Since(ctx.Get("start").(time.Time)), ctx.Response.Proto, ctx.Response.Status, headers, body))
		}

		h.Next(ctx)
	})
}

// Unmarshal a response into a given structure `s`. Supports both JSON and
// YAML depending on the response's content-type header.
func Unmarshal(resp *gentleman.Response, s interface{}) error {
	data := resp.Bytes()

	if len(data) == 0 {
		return nil
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d:\n%s", resp.StatusCode, string(data))
	}

	ct := resp.Header.Get("content-type")
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
