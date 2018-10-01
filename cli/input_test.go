package cli_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/danielgtaylor/openapi-cli-generator/cli"
)

func deepAssign(target, source string) string {
	var targetMap map[string]interface{}
	if err := json.Unmarshal([]byte(target), &targetMap); err != nil {
		panic(err)
	}

	var sourceMap map[string]interface{}
	if err := json.Unmarshal([]byte(source), &sourceMap); err != nil {
		panic(err)
	}

	cli.DeepAssign(targetMap, sourceMap)

	marshalled, err := json.MarshalIndent(targetMap, "", "  ")
	if err != nil {
		panic(err)
	}

	return string(marshalled)
}

func TestDeepAssignMerge(t *testing.T) {
	target := `{
		"foo": {
			"bar": {
				"baz": 1
			}
		}
	}`

	source := `{
		"foo": {
			"bar": {
				"blarg": true
			}
		}
	}`

	expected := `{
		"foo": {
			"bar": {
				"baz": 1,
				"blarg": true
			}
		}
	}`

	result := deepAssign(target, source)

	assert.JSONEq(t, expected, result)
}

func TestDeepAssignOverwrite(t *testing.T) {
	target := `{
		"foo": {
			"bar": {
				"baz": 1
			}
		}
	}`

	source := `{
		"foo": [1, 2, 3]
	}`

	expected := `{
		"foo": [1, 2, 3]
	}`

	result := deepAssign(target, source)

	assert.JSONEq(t, expected, result)
}
