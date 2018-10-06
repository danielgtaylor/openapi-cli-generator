package shorthand

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"
	"strconv"
	"strings"
)

//go:generate pigeon -o generated.go shorthand.peg

const (
	modifierNone = iota
	modifierString
)

func toIfaceSlice(v interface{}) []interface{} {
	if v == nil {
		return nil
	}
	return v.([]interface{})
}

func repeatedWithIndex(v interface{}, index int, cb func(v interface{})) {
	for _, i := range v.([]interface{}) {
		cb(i.([]interface{})[index])
	}
}

// list can be appended in-place while building structured data.
type list []interface{}

func (l *list) Append(v interface{}) {
	*l = append(*l, v)
}

func (l *list) String() string {
	return fmt.Sprintf("%v", *l)
}

// AST contains all of the key-value pairs in the document.
type AST []*KeyValue

// KeyValue groups a Key with the key's associated value.
type KeyValue struct {
	PostProcess bool
	Key         *Key
	Value       interface{}
}

// Key contains parts and key-specific configuration.
type Key struct {
	ResetContext bool
	Parts        []*KeyPart
}

// KeyPart has a name and optional indices.
type KeyPart struct {
	Key   string
	Index []int
}

// ParseAndBuild takes a string and returns the structured data it represents.
func ParseAndBuild(filename, input string) (map[string]interface{}, error) {
	parsed, err := Parse(filename, []byte(input))
	if err != nil {
		return nil, err
	}

	return Build(parsed.(AST))
}

// Build an AST of key-value pairs into structured data.
func Build(ast AST) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	ctx := result
	var ctxSlice *list

	for _, kv := range ast {
		k := kv.Key
		v := kv.Value

		if subAST, ok := v.(AST); ok {
			// If the value is itself an AST, then recursively process it!
			parsed, err := Build(subAST)
			if err != nil {
				return result, err
			}
			v = parsed
		} else if vStr, ok := v.(string); ok && kv.PostProcess {
			// If the value is a string, then handle special cases here.
			if len(vStr) > 1 && strings.HasPrefix(vStr, "@") {
				filename := vStr[1:]

				forceString := false
				useBase64 := false
				if filename[0] == '~' {
					forceString = true
					filename = filename[1:]
				} else if filename[0] == '%' {
					forceString = true
					filename = filename[1:]
					useBase64 = true
				}
				data, err := ioutil.ReadFile(filename)
				if err != nil {
					return result, err
				}

				if !forceString && strings.HasSuffix(vStr, ".json") {
					// Try to load data from JSON file.
					var unmarshalled interface{}

					if err := json.Unmarshal(data, &unmarshalled); err != nil {
						return result, err
					}

					v = unmarshalled
				} else {
					if useBase64 {
						v = base64.StdEncoding.EncodeToString(data)
					} else {
						v = string(data)
					}
				}
			}
		}

		// Reset context to the root or keep going from where we left off.
		if k.ResetContext {
			ctx = result
		}

		for ki, kp := range k.Parts {
			// If there is a key, and the key is not in the current context, then it
			// must be created as either a list or map depending on whether there
			// are index items for one or more lists.
			if kp.Key != "" && (ki < len(k.Parts)-1 || len(kp.Index) > 0) {
				if ctx[kp.Key] == nil {
					if len(kp.Index) > 0 {
						ctx[kp.Key] = &list{}
						ctxSlice = ctx[kp.Key].(*list)
					} else {
						ctx[kp.Key] = make(map[string]interface{})
						ctx = ctx[kp.Key].(map[string]interface{})
					}
				} else {
					if len(kp.Index) > 0 {
						ctxSlice = ctx[kp.Key].(*list)
					} else {
						ctx = ctx[kp.Key].(map[string]interface{})
					}
				}
			}

			// For each index item, create the associated list item and update the
			// context.
			for i, index := range kp.Index {
				if index == -1 {
					if ctxSlice != nil {
						index = len(*ctxSlice)
					} else {
						index = 0
					}
				}

				for len(*ctxSlice) < index+1 {
					// Increase the size of the list to fit the new item if needed.
					ctxSlice.Append(nil)
				}

				if i < len(kp.Index)-1 {
					// Not the last index item, so create another list!
					(*ctxSlice)[index] = &list{}
					ctxSlice = (*ctxSlice)[index].(*list)
				} else {
					// This is the last index item. If it is also the last key part, then
					// set the value. Otherwise, create a map for the next key part to
					// use and update the context.
					if ki < len(k.Parts)-1 {
						if (*ctxSlice)[index] == nil {
							(*ctxSlice)[index] = make(map[string]interface{})
						}
						ctx = (*ctxSlice)[index].(map[string]interface{})
					} else {
						(*ctxSlice)[index] = v
					}
				}
			}

			// If this is the last key part and has no list indexes, then just set
			// the value on the current context.
			if ki == len(k.Parts)-1 && len(kp.Index) == 0 {
				ctx[kp.Key] = v

				if vSlice, ok := v.(*list); ok {
					ctxSlice = vSlice
				}
			}
		}
	}

	return result, nil
}

// Get the shorthand representation of an input map.
func Get(input map[string]interface{}) string {
	result := renderValue(true, input)
	return result[1 : len(result)-1]
}

func renderValue(start bool, value interface{}) string {
	// Go uses `<nil>` so here we hard-code `null` to match JSON/YAML.
	if value == nil {
		return ": null"
	}

	switch v := value.(type) {
	case map[string]interface{}:
		// Special case: foo.bar: 1
		if !start && len(v) == 1 {
			for k := range v {
				return "." + k + renderValue(false, v[k])
			}
		}

		// Normal case: foo{a: 1, b: 2}
		var keys []string

		for k := range v {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		var fields []string
		for _, k := range keys {
			fields = append(fields, k+renderValue(false, v[k]))
		}

		return "{" + strings.Join(fields, ", ") + "}"
	case []interface{}:
		var items []string

		// Special case: foo: 1, 2, 3
		scalars := true
		for _, item := range v {
			switch item.(type) {
			case map[string]interface{}:
				scalars = false
				break
			case []interface{}:
				scalars = false
				break
			}
		}

		if scalars {
			for _, item := range v {
				items = append(items, fmt.Sprintf("%v", item))
			}

			return ": " + strings.Join(items, ", ")
		}

		// Normal case: foo[]: 1, []{id: 1, count: 2}
		for _, item := range v {
			items = append(items, "[]"+renderValue(false, item))
		}

		return strings.Join(items, ", ")
	default:
		modifier := ""

		if s, ok := v.(string); ok {
			_, err := strconv.ParseFloat(s, 64)

			if err == nil || s == "null" || s == "true" || s == "false" {
				modifier = "~"
			}

			if len(s) > 50 || strings.Contains(s, "\n") {
				v = "@file"
			}
		}

		return fmt.Sprintf(":%s %v", modifier, v)
	}
}
