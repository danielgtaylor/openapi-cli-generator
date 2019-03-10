package cli

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	gentleman "gopkg.in/h2non/gentleman.v2"
)

func TestMatchValueInvalidSelector(t *testing.T) {
	_, err := GetMatchValue("invalid", nil, nil, nil, nil)
	assert.Error(t, err)
}

func TestMatchValueReqParam(t *testing.T) {
	params := map[string]interface{}{
		"id": "foo",
	}

	value, err := GetMatchValue("request.param#id", nil, params, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, "foo", value)
}

func TestMatchValueReqBody(t *testing.T) {
	// Matching a request body assumes we have a content type set and that
	// the body content has been saved in the request context by the log
	// middleware.
	req := gentleman.New().Post()
	req.Context.Request.Header.Add("Content-Type", "application/json")

	// Bad body
	req.Context.Set("request-body", `{"foo": {"bar": 2`)
	value, err := GetMatchValue("request.body#foo..bar", req, nil, nil, nil)
	assert.Error(t, err)

	// Fix the body
	req.Context.Set("request-body", `{"foo": {"bar": 2}}`)

	// Bad JMESPath value
	value, err = GetMatchValue("request.body#foo..bar", req, nil, nil, nil)
	assert.Error(t, err)

	// Correct query
	value, err = GetMatchValue("request.body#foo.bar", req, nil, nil, nil)
	assert.NoError(t, err)
	assert.EqualValues(t, 2, value)
}

func TestMatchValueResStatus(t *testing.T) {
	response := &gentleman.Response{
		StatusCode: 200,
	}

	value, err := GetMatchValue("response.status", nil, nil, response, nil)
	assert.NoError(t, err)
	assert.EqualValues(t, 200, value)
}

func TestMatchValueResHeader(t *testing.T) {
	headers := make(http.Header)

	headers.Add("x-ready", "true")

	response := &gentleman.Response{
		Header: headers,
	}

	value, err := GetMatchValue("response.header#x-ready", nil, nil, response, nil)
	assert.NoError(t, err)
	assert.EqualValues(t, "true", value)
}

func TestMatchValueResBody(t *testing.T) {
	var decoded interface{}
	err := json.Unmarshal([]byte(`{"foo": {"bar": 2}}`), &decoded)
	assert.NoError(t, err)

	// Invalid query
	value, err := GetMatchValue("response.body#foo..bar", nil, nil, nil, decoded)
	assert.Error(t, err)

	// Valid query
	value, err = GetMatchValue("response.body#foo.bar", nil, nil, nil, decoded)
	assert.NoError(t, err)
	assert.EqualValues(t, 2, value)
}

func TestMatchBadJSON(t *testing.T) {
	expected := json.RawMessage([]byte("{"))
	_, err := Match("equal", expected, 2)
	assert.Error(t, err)
}

func TestMatchBadTestString(t *testing.T) {
	expected := json.RawMessage([]byte("null"))
	_, err := Match("invalid", expected, nil)
	assert.Error(t, err)
}

func TestMatchEqual(t *testing.T) {
	expected := json.RawMessage([]byte("2"))
	m, err := Match("equal", expected, 2)
	assert.NoError(t, err)
	assert.Equal(t, true, m)
}

func TestMatchNoMatch(t *testing.T) {
	expected := json.RawMessage([]byte("2"))
	m, err := Match("equal", expected, 3)
	assert.NoError(t, err)
	assert.Equal(t, false, m)
}

func TestMatchEqualComplex(t *testing.T) {
	expected := json.RawMessage([]byte(`{"foo": [1.0, 2.0]}`))
	var actual interface{}
	json.Unmarshal(expected, &actual)
	m, err := Match("equal", expected, actual)
	assert.NoError(t, err)
	assert.Equal(t, true, m)
}

func TestMatchNotListAll(t *testing.T) {
	expected := json.RawMessage([]byte(`"ready"`))
	_, err := Match("all", expected, "invalid")
	assert.Error(t, err)
}

func TestMatchListAll(t *testing.T) {
	expected := json.RawMessage([]byte(`"ready"`))
	m, err := Match("all", expected, []interface{}{"ready", "ready"})
	assert.NoError(t, err)
	assert.Equal(t, true, m)
}

func TestMatchListAllFail(t *testing.T) {
	expected := json.RawMessage([]byte(`"ready"`))
	m, err := Match("all", expected, []interface{}{"ready", "off", "ready"})
	assert.NoError(t, err)
	assert.Equal(t, false, m)
}

func TestMatchNotListAny(t *testing.T) {
	expected := json.RawMessage([]byte(`"ready"`))
	_, err := Match("any", expected, "invalid")
	assert.Error(t, err)
}

func TestMatchListAny(t *testing.T) {
	expected := json.RawMessage([]byte(`"ready"`))
	m, err := Match("any", expected, []interface{}{"off", "ready"})
	assert.NoError(t, err)
	assert.Equal(t, true, m)
}

func TestMatchListAnyFail(t *testing.T) {
	expected := json.RawMessage([]byte(`"nope"`))
	m, err := Match("any", expected, []interface{}{"ready", "off", "ready"})
	assert.NoError(t, err)
	assert.Equal(t, false, m)
}
