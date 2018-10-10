package cli

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	gentleman "gopkg.in/h2non/gentleman.v2"
)

// BeforeHandlerFunc is a function that runs before a command sends a request
// over the wire. It may modify the request.
type BeforeHandlerFunc func(*cobra.Command, *viper.Viper, *gentleman.Request)

// AfterHandlerFunc is a function that runs after a request has been sent and
// the response is unmarshalled. It may modify the response. It must return
// the response data regardless of whether it was modified.
type AfterHandlerFunc func(*cobra.Command, *viper.Viper, *gentleman.Response, interface{}) interface{}

var beforeRegistry = make(map[string][]BeforeHandlerFunc)
var afterRegistry = make(map[string][]AfterHandlerFunc)

// RegisterBefore registers a pre-request handler for the given command path.
// The handler may modify the request before it gets sent over the wire.
func RegisterBefore(path string, handler BeforeHandlerFunc) {
	if _, ok := beforeRegistry[path]; !ok {
		beforeRegistry[path] = make([]BeforeHandlerFunc, 0, 1)
	}

	beforeRegistry[path] = append(beforeRegistry[path], handler)
}

// RegisterAfter registers a post-request handler for the given command path.
// The handler may modify the unmarshalled response.
func RegisterAfter(path string, handler AfterHandlerFunc) {
	if _, ok := afterRegistry[path]; !ok {
		afterRegistry[path] = make([]AfterHandlerFunc, 0, 1)
	}

	afterRegistry[path] = append(afterRegistry[path], handler)
}

func commandPath(cmd *cobra.Command) string {
	parts := make([]string, 0)
	cur := cmd
	for cur != nil {
		name := strings.Split(cur.Use, " ")[0]
		parts = append([]string{name}, parts...)
		cur = cur.Parent()
	}

	return strings.Join(parts[1:], " ")
}

// HandleBefore runs any registered pre-request handlers for the given command.
func HandleBefore(cmd *cobra.Command, params *viper.Viper, r *gentleman.Request) {
	path := commandPath(cmd)

	if handlers, ok := beforeRegistry[path]; ok {
		for _, handler := range handlers {
			handler(cmd, params, r)
		}
	}
}

// HandleAfter runs any regeistered post-request handlers for the given command.
func HandleAfter(cmd *cobra.Command, params *viper.Viper, resp *gentleman.Response, data interface{}) interface{} {
	path := commandPath(cmd)

	tmp := data
	if handlers, ok := afterRegistry[path]; ok {
		for _, handler := range handlers {
			tmp = handler(cmd, params, resp, tmp)
		}
	}

	return tmp
}
