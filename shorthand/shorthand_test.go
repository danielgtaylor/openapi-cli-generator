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

func TestParseIP(t *testing.T) {
	assert.JSONEq(t, `{"foo": "1.2.3.4"}`, parsed(`foo: 1.2.3.4`))
}

func TestParseTrailingSpace(t *testing.T) {
	assert.JSONEq(t, `{"foo": {"a": 1}}`, parsed(`foo{a: 1 }`))
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

func TestGetShorthandSimple(t *testing.T) {
	result := Get(map[string]interface{}{
		"foo": "bar",
	})
	assert.Equal(t, "foo: bar", result)
}

func TestGetShorthandMultiKey(t *testing.T) {
	result := Get(map[string]interface{}{
		"foo":   "bar",
		"hello": "world",
		"num":   1,
		"empty": nil,
		"bool":  false,
	})
	assert.Equal(t, "bool: false, empty: null, foo: bar, hello: world, num: 1", result)
}

func TestGetShorthandNestedSimple(t *testing.T) {
	result := Get(map[string]interface{}{
		"foo": map[string]interface{}{
			"bar": 1,
		},
	})
	assert.Equal(t, "foo.bar: 1", result)
}

func TestGetShorthandNestedMultiKey(t *testing.T) {
	result := Get(map[string]interface{}{
		"foo": map[string]interface{}{
			"bar": 1,
			"baz": 2,
		},
	})
	assert.Equal(t, "foo{bar: 1, baz: 2}", result)
}

func TestGetShorthandListSimple(t *testing.T) {
	result := Get(map[string]interface{}{
		"foo": []interface{}{1, 2, 3},
	})
	assert.Equal(t, "foo: 1, 2, 3", result)
}

func TestGetShorthandListOfListScalar(t *testing.T) {
	result := Get(map[string]interface{}{
		"foo": []interface{}{
			[]interface{}{1, 2, 3},
		},
	})
	assert.Equal(t, "foo[]: 1, 2, 3", result)
}

func TestGetShorthandListOfObjects(t *testing.T) {
	result := Get(map[string]interface{}{
		"tags": []interface{}{
			map[string]interface{}{
				"id": "tag1",
				"count": map[string]interface{}{
					"clicks": 15,
					"sales":  3,
				},
			},
			map[string]interface{}{
				"id": "tag2",
				"count": map[string]interface{}{
					"clicks": 7,
					"sales":  4,
				},
			},
		},
	})
	assert.Equal(t, "tags[]{count{clicks: 15, sales: 3}, id: tag1}, []{count{clicks: 7, sales: 4}, id: tag2}", result)
}

func TestGetShorthandCoerced(t *testing.T) {
	result := Get(map[string]interface{}{
		"null": "null",
		"bool": "true",
		"num":  "1234",
		"str":  "hello",
	})
	assert.Equal(t, "bool:~ true, null:~ null, num:~ 1234, str: hello", result)
}

func TestGetShorthandFromFile(t *testing.T) {
	result := Get(map[string]interface{}{
		"multi": "I am\na multiline\n value.",
		"long":  "I am a really long line of text that should probably get loaded from a file",
	})
	assert.Equal(t, "long: @file, multi: @file", result)
}
