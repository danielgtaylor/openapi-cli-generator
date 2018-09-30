package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/danielgtaylor/openapi-cli-generator/shorthand"
	yaml "gopkg.in/yaml.v2"
)

// GetBody returns the request body if one was passed either as shorthand
// arguments or via stdin.
func GetBody(mediaType string, args []string) (string, error) {
	var body string
	if len(args) > 0 {
		bodyInput := strings.Join(args, " ")
		result, err := shorthand.ParseAndBuild("stdin", bodyInput)
		if err != nil {
			return "", err
		}

		if strings.Contains(mediaType, "json") {
			marshalled, err := json.Marshal(result)
			if err != nil {
				return "", err
			}
			body = string(marshalled)
		} else if strings.Contains(mediaType, "yaml") {
			marshalled, err := yaml.Marshal(result)
			if err != nil {
				return "", err
			}
			body = string(marshalled)
		} else {
			return "", fmt.Errorf("Not sure how to marshal %s", mediaType)
		}
	} else {
		info, err := os.Stdin.Stat()
		if err != nil {
			return "", err
		}
		if info.Size() > 0 {
			input, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				return "", err
			}

			body = string(input)
		}
	}

	return body, nil
}
