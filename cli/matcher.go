package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	jmespath "github.com/danielgtaylor/go-jmespath-plus"
	"github.com/rs/zerolog/log"
	"gopkg.in/h2non/gentleman.v2/context"
)

// The following equality functions are from stretchr/testify/assert.
// objectsAreEqual determines if two objects are considered equal.
func objectsAreEqual(expected, actual interface{}) bool {
	if expected == nil || actual == nil {
		return expected == actual
	}

	exp, ok := expected.([]byte)
	if !ok {
		return reflect.DeepEqual(expected, actual)
	}

	act, ok := actual.([]byte)
	if !ok {
		return false
	}
	if exp == nil || act == nil {
		return exp == nil && act == nil
	}
	return bytes.Equal(exp, act)
}

// objectsAreEqualValues gets whether two objects are equal, or if their
// values are equal.
func objectsAreEqualValues(expected, actual interface{}) bool {
	if objectsAreEqual(expected, actual) {
		return true
	}

	actualType := reflect.TypeOf(actual)
	if actualType == nil {
		return false
	}
	expectedValue := reflect.ValueOf(expected)
	if expectedValue.IsValid() && expectedValue.Type().ConvertibleTo(actualType) {
		// Attempt comparison after type conversion
		return reflect.DeepEqual(expectedValue.Convert(actualType).Interface(), actual)
	}

	return false
}

// GetMatchValue returns a value for the given selector query.
func GetMatchValue(ctx *context.Context, selector string, reqParams map[string]interface{}, decoded interface{}) (interface{}, error) {
	parts := strings.Split(selector, "#")
	base := parts[0]
	args := strings.Join(parts[1:], "#")

	l := log.Debug().Str("selector", base)

	if args != "" {
		l = l.Str("query", args)
	}

	var actual interface{}
	switch base {
	case "request.param":
		actual = reqParams[args]
	case "request.body":
		var decoded interface{}
		if err := UnmarshalRequest(ctx, &decoded); err != nil {
			return nil, err
		}

		result, err := jmespath.Search(args, decoded)
		if err != nil {
			return nil, err
		}

		actual = result
	case "response.status":
		actual = ctx.Response.StatusCode
	case "response.header":
		actual = ctx.Response.Header.Get(args)
	case "response.body":
		// Perform a JMESPath query to match the expected value.
		result, err := jmespath.Search(args, decoded)

		if err != nil {
			return nil, err
		}

		actual = result
	default:
		return false, fmt.Errorf("Cannot match selector: %s", selector)
	}

	l.Interface("actual", actual).Msg("Found matcher value")

	return actual, nil
}

// Match returns `true` if the expected value of the match type is found in the
// given response data.
func Match(test string, expected json.RawMessage, actual interface{}) (bool, error) {
	var exp interface{}
	if err := json.Unmarshal(expected, &exp); err != nil {
		return false, err
	}

	var matches bool
	switch test {
	case "equal":
		if objectsAreEqualValues(exp, actual) {
			matches = true
		}
	case "any":
		if list, ok := actual.([]interface{}); ok {
			// We have a list of items, so at least one must match.
			matches = false
			for _, item := range list {
				if objectsAreEqualValues(exp, item) {
					matches = true
					break
				}
			}
		} else {
			return false, fmt.Errorf("Expected a list but got: %v", actual)
		}
	case "all":
		if list, ok := actual.([]interface{}); ok {
			// We have a list of items, so each one must match the expected value.
			matches = true
			for _, item := range list {
				if !objectsAreEqualValues(exp, item) {
					matches = false
					break
				}
			}
		} else {
			return false, fmt.Errorf("Expected a list but got: %v", actual)
		}
	default:
		return false, fmt.Errorf("Unknown test: %s", test)
	}

	l := log.Debug().Str("test", test).Interface("expected", exp).Interface("actual", actual)

	if matches {
		l.Msg("Found match")
	} else {
		l.Msg("No match")
	}

	return matches, nil
}
