// Note: do not run with `go test`. Instead use the `./test.sh` script!
package main_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
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
	// Note, `echo-query` has `x-cli-name` set in OAS definition.

	out, err := exec.Command("sh", "-c", "example-cli echo hello: world --apiUrl http://127.0.0.1:8005 --echo-query=foo --xRequestId bar").CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		panic(err)
	}

	assert.JSONEq(t, "{\"hello\": \"world\", \"q\": \"foo\", \"request-id\": \"bar\"}", string(out))
}

func TestConfiguration(t *testing.T) {
	t.Run("can set and read settings", func(t *testing.T) {
		var out []byte
		var err error
		out, err = exec.Command("sh", "-c", "example-cli config setting default_profile_name bogus").CombinedOutput()
		if err != nil {
			fmt.Println(string(out))
			panic(err)
		}

		out, err = exec.Command("sh", "-c", "example-cli config setting default_profile_name").CombinedOutput()
		if err != nil {
			fmt.Println(string(out))
			panic(err)
		}

		assert.Equal(t, "bogus", strings.TrimSpace(string(out)))
	})

}

func TestAuth(t *testing.T) {
	t.Run("can add and list auth servers", func(t *testing.T) {
		var out []byte
		var err error
		out, err = exec.Command("sh", "-c", "example-cli auth addServer auth1 --issuer https://auth.test.sh --clientId 01").CombinedOutput()
		if err != nil {
			fmt.Println(string(out))
			panic(err)
		}

		out, err = exec.Command("sh", "-c", "example-cli auth listServers").CombinedOutput()
		if err != nil {
			fmt.Println(string(out))
			panic(err)
		}
		assert.Contains(t,  string(out),"| auth1   |        01 | https://auth.test.sh |")
	})
}

func TestGlobalFlags(t *testing.T) {
	t.Run("", func(t *testing.T) {
		out, _ := exec.Command("sh", "-c", "example-cli echo --help").CombinedOutput()
		fmt.Println(string(out))
		assert.Contains(t, string(out), "--apiUrl string")
		assert.Contains(t, string(out), "--authServerName string")
		assert.Contains(t, string(out), "--credentialsName string")
		assert.Contains(t, string(out), "--outputFormat string")
		assert.Contains(t, string(out), "--profileName string")
		assert.Contains(t, string(out), "--raw")
	})

}


