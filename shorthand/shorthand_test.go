package shorthand

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func parsed(input string) string {
	result, err := ParseAndBuild("stdin", input)
	if err != nil {
		panic(err)
	}

	j, _ := json.Marshal(result)
	return string(j)
}

func TestParseCoerce(t *testing.T) {
	result := parsed(`n: null, b: true, i: 1, f: 1.0, s: hello`)
	assert.JSONEq(t, `{"n": null, "b": true, "i": 1, "f": 1.0, "s": "hello"}`, result)
}

func TestParseWhitespace(t *testing.T) {
	assert.JSONEq(t, `{"foo": "hello", "bar": "world"}`, parsed(`foo :    hello   ,    bar:world  `))
}

func TestParseForceString(t *testing.T) {
	assert.JSONEq(t, `{"foo": "1"}`, parsed(`foo:~ 1`))
}

func TestParseMultipleProperties(t *testing.T) {
	assert.JSONEq(t, `{"foo": {"bar": {"baz": 1}}}`, parsed(`foo.bar.baz: 1`))
}

func TestParseContext(t *testing.T) {
	result := parsed(`foo.bar: 1, .baz: 2, qux: 3`)
	assert.JSONEq(t, `{"foo": {"bar": 1, "baz": 2}, "qux": 3}`, result)
}

func TestParsePropertyGrouping(t *testing.T) {
	assert.JSONEq(t, `{"foo": {"bar": 1, "baz": 2}}`, parsed(`foo{bar: 1, baz: 2}`))
}

func TestParserShortList(t *testing.T) {
	assert.JSONEq(t, `{"foo": [1, 2, 3]}`, parsed(`foo: 1, 2, 3`))
}

func TestParserShortStringList(t *testing.T) {
	assert.JSONEq(t, `{"foo": ["1", "2", "3"]}`, parsed(`foo:~ 1, 2, 3`))
}

func TestParserListOfList(t *testing.T) {
	assert.JSONEq(t, `{"foo": [[null, [1]]]}`, parsed(`foo[][1][]: 1`))
}

func TestParserAppendList(t *testing.T) {
	assert.JSONEq(t, `{"foo": [1, 2, 3]}`, parsed(`foo[]: 1, []: 2, []: 3`))
}

func TestParserListIndex(t *testing.T) {
	assert.JSONEq(t, `{"foo": [true, null, null, "three", null, "five"]}`,
		parsed(`foo[3]: three, foo[5]: five, foo[0]: true`))
}

func TestParserListIndexObject(t *testing.T) {
	result := parsed(`foo[0].bar: 1, foo[0].baz: 2`)
	assert.JSONEq(t, `{"foo": [{"bar": 1, "baz": 2}]}`, result)
}

func TestParserAppendBackRef(t *testing.T) {
	assert.JSONEq(t, `{"foo": [1, 3], "bar": 2}`, parsed(`foo[]: 1, bar: 2, []: 3`))
}

func TestParserListObjects(t *testing.T) {
	result := parsed(`foo[].id: 1, .count: 1, [].id: 2, .count: 2`)
	assert.JSONEq(t, `{"foo": [{"id": 1, "count": 1}, {"id": 2, "count": 2}]}`, result)
}

func TestParserNonFile(t *testing.T) {
	assert.JSONEq(t, `{"foo": "@user"}`, parsed(`foo:~ @user`))
}

func TestParserFileStructured(t *testing.T) {
	result := parsed(`foo: @testdata/hello.json`)
	assert.JSONEq(t, `{"foo": {"hello": "world"}}`, result)
}

func TestParserFileForceString(t *testing.T) {
	result := parsed(`foo: @~testdata/hello.json`)
	assert.JSONEq(t, `{"foo": "{\n  \"hello\": \"world\"\n}\n"}`, result)
}
