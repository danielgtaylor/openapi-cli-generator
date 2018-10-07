// Note: do not run with `go test`. Instead use the `./test.sh` script!
package main_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"testing"

	"github.com/alecthomas/assert"
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

		w.Write(body)
	})

	go func() {
		server.ListenAndServe()
	}()
	defer server.Shutdown(context.Background())

	os.Exit(m.Run())
}

func TestEchoSuccess(t *testing.T) {
	// Call the precompiled executable CLI to hit our test server.
	out, err := exec.Command("sh", "-c", "example-cli echo hello: world").CombinedOutput()
	if err != nil {
		panic(err)
	}

	assert.Equal(t, "{\n  \"hello\": \"world\"\n}\n\n", string(out))
}
