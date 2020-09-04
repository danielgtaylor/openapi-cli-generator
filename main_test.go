// Note: do not run with `go test`. Instead use the `./test.sh` script!
package main_test

import (
	"context"
	"encoding/json"
	"fmt"
	main "github.com/danielgtaylor/openapi-cli-generator"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	// Set up a test server that implements the API described in
	// `./example-cli/openapi.yaml` and start it before running the tests.
	server := &http.Server{Addr: ":8005", Handler: http.DefaultServeMux}

	http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}

		ct := r.Header.Get("Content-Type")
		if ct != "" {
			w.Header().Add("Content-Type", ct)
		}

		// For the test, add the param values to the echoed response.
		var decoded map[string]interface{}
		json.Unmarshal(body, &decoded)

		q := r.URL.Query().Get("q")
		if q != "" {
			decoded["q"] = q
		}

		rid := r.Header.Get("X-Request-ID")
		if rid != "" {
			decoded["request-id"] = rid
		}

		marshalled, _ := json.Marshal(decoded)
		w.Write(marshalled)
	})

	go func() {
		server.ListenAndServe()
	}()
	defer server.Shutdown(context.Background())

	os.Exit(m.Run())
}

func TestEchoSuccess(t *testing.T) {
	// Call the precompiled executable CLI to hit our test server.
	out, err := exec.Command("sh", "-c", "example-cli echo hello: world --echo-query=foo --x-request-id bar").CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		panic(err)
	}

	assert.JSONEq(t, "{\"hello\": \"world\", \"q\": \"foo\", \"request-id\": \"bar\"}", string(out))
}

func Test_slug(t *testing.T) {
	type args struct {
		operationID string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "lowercase with spaces", args: args{operationID: "get all articles"}, want: "get-all-articles"},
		{name: "Mixed sentence case", args: args{operationID: "Get all articles"}, want: "get-all-articles"},
		{name: "lower snake case", args: args{operationID: "get_all_articles"}, want: "get-all-articles"},
		{name: "upper snake case", args: args{operationID: "GET_ALL_ARTICLES"}, want: "get-all-articles"},
		{name: "kebab case", args: args{operationID: "get-all-articles"}, want: "get-all-articles"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := main.Slug(tt.args.operationID); got != tt.want {
				t.Errorf("Slug() = %v, want %v", got, tt.want)
			}
		})
	}
}