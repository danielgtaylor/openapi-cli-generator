package main

import (
	"encoding/json"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sort"
	"strings"
	"text/template"

	"github.com/danielgtaylor/openapi-cli-generator/shorthand"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

//go:generate go-bindata ./templates/...

// OpenAPI Extensions
const (
	ExtAliases     = "x-cli-aliases"
	ExtDescription = "x-cli-description"
	ExtIgnore      = "x-cli-ignore"
	ExtName        = "x-cli-name"
)

// Param describes an OpenAPI parameter (path, query, header, etc)
type Param struct {
	Name        string
	CLIName     string
	GoName      string
	Description string
	In          string
	Required    bool
	Type        string
	TypeNil     string
}

// Operation describes an OpenAPI operation (GET/POST/PUT/PATCH/DELETE)
type Operation struct {
	Use            string
	Aliases        []string
	Short          string
	Long           string
	Method         string
	CanHaveBody    bool
	Path           string
	AllParams      []*Param
	RequiredParams []*Param
	OptionalParams []*Param
	MediaType      string
	Examples       []string
}

// Server describes an OpenAPI server endpoint
type Server struct {
	Description string
	URL         string
	// TODO: handle server parameters
}

// Imports describe optional imports based on features in use.
type Imports struct {
	Fmt     bool
	Strings bool
}

// OpenAPI describes an API
type OpenAPI struct {
	Imports     Imports
	Name        string
	GoName      string
	Title       string
	Description string
	Servers     []*Server
	Operations  []*Operation
}

// ProcessAPI returns the API description to be used with the commands template
// for a loaded and dereferenced OpenAPI 3 document.
func ProcessAPI(shortName string, api *openapi3.Swagger) *OpenAPI {
	apiName := shortName
	if api.Info.Extensions[ExtName] != nil {
		apiName = extStr(api.Info.Extensions[ExtName])
	}

	apiDescription := api.Info.Description
	if api.Info.Extensions[ExtDescription] != nil {
		apiDescription = extStr(api.Info.Extensions[ExtDescription])
	}

	result := &OpenAPI{
		Name:        apiName,
		GoName:      toGoName(shortName, false),
		Title:       api.Info.Title,
		Description: escapeString(apiDescription),
	}

	for _, s := range api.Servers {
		result.Servers = append(result.Servers, &Server{
			Description: s.Description,
			URL:         s.URL,
		})
	}

	var keys []string
	for path := range api.Paths {
		keys = append(keys, path)
	}
	sort.Strings(keys)

	for _, path := range keys {
		item := api.Paths[path]

		if item.Extensions[ExtIgnore] != nil {
			// Ignore this path.
			continue
		}

		for method, operation := range item.Operations() {
			if operation.Extensions[ExtIgnore] != nil {
				// Ignore this operation.
				continue
			}

			name := operation.OperationID
			if operation.Extensions[ExtName] != nil {
				name = extStr(operation.Extensions[ExtName])
			}

			var aliases []string
			if operation.Extensions[ExtAliases] != nil {
				// We need to decode the raw extension value into our string slice.
				json.Unmarshal(operation.Extensions[ExtAliases].(json.RawMessage), &aliases)
			}

			params := getParams(item, method)
			requiredParams := getRequiredParams(params)
			optionalParams := getOptionalParams(params)
			short := operation.Summary
			if short == "" {
				short = name
			}

			use := usage(name, requiredParams)

			description := operation.Description
			if operation.Extensions[ExtDescription] != nil {
				description = extStr(operation.Extensions[ExtDescription])
			}

			reqMt, reqSchema, reqExamples := getRequestInfo(operation)

			var examples []string
			if len(reqExamples) > 0 {
				wroteHeader := false
				for _, ex := range reqExamples {
					if _, ok := ex.(string); !ok {
						// Not a string, so it's structured data. Let's marshal it to the
						// shorthand syntax if we can.
						if m, ok := ex.(map[string]interface{}); ok {
							ex = shorthand.Get(m)
							examples = append(examples, ex.(string))
							continue
						}

						b, _ := json.Marshal(ex)

						if !wroteHeader {
							description += "\n## Input Example\n\n"
							wroteHeader = true
						}

						description += "\n" + string(b) + "\n"
						continue
					}

					if !wroteHeader {
						description += "\n## Input Example\n\n"
						wroteHeader = true
					}

					description += "\n" + ex.(string) + "\n"
				}
			}

			if reqSchema != "" {
				description += "\n## Request Schema (" + reqMt + ")\n\n" + reqSchema
			}

			method := strings.Title(strings.ToLower(method))

			o := &Operation{
				Use:            use,
				Aliases:        aliases,
				Short:          short,
				Long:           escapeString(description),
				Method:         method,
				CanHaveBody:    method == "Post" || method == "Put" || method == "Patch",
				Path:           path,
				AllParams:      params,
				RequiredParams: requiredParams,
				OptionalParams: optionalParams,
				MediaType:      reqMt,
				Examples:       examples,
			}

			result.Operations = append(result.Operations, o)

			for _, p := range params {
				if p.In == "path" {
					result.Imports.Strings = true
				}
			}

			for _, p := range optionalParams {
				if p.In == "query" || p.In == "header" {
					result.Imports.Fmt = true
				}
			}
		}
	}

	return result
}

// extStr returns the string value of an OpenAPI extension stored as a JSON
// raw message.
func extStr(i interface{}) (decoded string) {
	if err := json.Unmarshal(i.(json.RawMessage), &decoded); err != nil {
		panic(err)
	}

	return
}

func toGoName(input string, public bool) string {
	transformed := strings.Replace(input, "-", " ", -1)
	transformed = strings.Replace(transformed, "_", " ", -1)
	transformed = strings.Title(transformed)
	transformed = strings.Replace(transformed, " ", "", -1)

	if !public {
		transformed = strings.ToLower(string(transformed[0])) + transformed[1:]
	}

	return transformed
}

func escapeString(value string) string {
	transformed := strings.Replace(value, "\n", "\\n", -1)
	transformed = strings.Replace(transformed, "\"", "\\\"", -1)
	return transformed
}

func slug(operationID string) string {
	transformed := strings.ToLower(operationID)
	transformed = strings.Replace(transformed, "_", "-", -1)
	transformed = strings.Replace(transformed, " ", "-", -1)
	return transformed
}

func usage(name string, requiredParams []*Param) string {
	usage := slug(name)

	for _, p := range requiredParams {
		usage += " " + slug(p.Name)
	}

	return usage
}

func getParams(path *openapi3.PathItem, httpMethod string) []*Param {
	operation := path.Operations()[httpMethod]
	allParams := make([]*Param, 0, len(path.Parameters))

	var total openapi3.Parameters
	total = append(total, path.Parameters...)
	total = append(total, operation.Parameters...)

	for _, p := range total {
		if p.Value != nil && p.Value.Extensions["x-cli-ignore"] == nil {
			t := "string"
			tn := "\"\""
			if p.Value.Schema != nil && p.Value.Schema.Value != nil && p.Value.Schema.Value.Type != "" {
				switch p.Value.Schema.Value.Type {
				case "boolean":
					t = "bool"
					tn = "false"
				case "integer":
					t = "int64"
					tn = "0"
				case "number":
					t = "float64"
					tn = "0.0"
				}
			}

			cliName := slug(p.Value.Name)
			if p.Value.Extensions[ExtName] != nil {
				cliName = extStr(p.Value.Extensions[ExtName])
			}

			description := p.Value.Description
			if p.Value.Extensions[ExtDescription] != nil {
				description = extStr(p.Value.Extensions[ExtDescription])
			}

			allParams = append(allParams, &Param{
				Name:        p.Value.Name,
				CLIName:     cliName,
				GoName:      toGoName("param "+cliName, false),
				Description: description,
				In:          p.Value.In,
				Required:    p.Value.Required,
				Type:        t,
				TypeNil:     tn,
			})
		}
	}

	return allParams
}

func getRequiredParams(allParams []*Param) []*Param {
	required := make([]*Param, 0)

	for _, param := range allParams {
		if param.Required || param.In == "path" {
			required = append(required, param)
		}
	}

	return required
}

func getOptionalParams(allParams []*Param) []*Param {
	optional := make([]*Param, 0)

	for _, param := range allParams {
		if !param.Required && param.In != "path" {
			optional = append(optional, param)
		}
	}

	return optional
}

func getRequestInfo(op *openapi3.Operation) (string, string, []interface{}) {
	mts := make(map[string][]interface{})

	if op.RequestBody != nil && op.RequestBody.Value != nil {
		for mt, item := range op.RequestBody.Value.Content {
			var schema string
			var examples []interface{}

			if item.Schema != nil && item.Schema.Value != nil {
				// Let's make this a bit more concise. Since it has special JSON
				// marshalling functions, we do a dance to get it into plain JSON before
				// converting to YAML.
				data, err := json.Marshal(item.Schema.Value)
				if err != nil {
					continue
				}

				var unmarshalled interface{}
				json.Unmarshal(data, &unmarshalled)

				data, err = yaml.Marshal(unmarshalled)
				if err == nil {
					schema = string(data)
				}
			}

			if item.Example != nil {
				examples = append(examples, item.Example)
			} else {
				for _, ex := range item.Examples {
					if ex.Value != nil {
						examples = append(examples, ex.Value.Value)
						break
					}
				}
			}

			mts[mt] = []interface{}{schema, examples}
		}
	}

	// Prefer JSON.
	for mt, item := range mts {
		if strings.Contains(mt, "json") {
			return mt, item[0].(string), item[1].([]interface{})
		}
	}

	// Fall back to YAML next.
	for mt, item := range mts {
		if strings.Contains(mt, "yaml") {
			return mt, item[0].(string), item[1].([]interface{})
		}
	}

	// Last resort: return the first we find!
	for mt, item := range mts {
		return mt, item[0].(string), item[1].([]interface{})
	}

	return "", "", nil
}

func writeFormattedFile(filename string, data []byte) {
	formatted, errFormat := format.Source(data)
	if errFormat != nil {
		formatted = data
	}

	err := ioutil.WriteFile(filename, formatted, 0600)
	if errFormat != nil {
		panic(errFormat)
	} else if err != nil {
		panic(err)
	}
}

func initCmd(cmd *cobra.Command, args []string) {
	if _, err := os.Stat("main.go"); err == nil {
		fmt.Println("Refusing to overwrite existing main.go")
		return
	}

	data, _ := Asset("templates/main.tmpl")
	tmpl, err := template.New("cli").Parse(string(data))
	if err != nil {
		panic(err)
	}

	templateData := map[string]string{
		"Name":    args[0],
		"NameEnv": strings.Replace(strings.ToUpper(args[0]), "-", "_", -1),
	}

	var sb strings.Builder
	err = tmpl.Execute(&sb, templateData)
	if err != nil {
		panic(err)
	}

	writeFormattedFile("main.go", []byte(sb.String()))
}

func generate(cmd *cobra.Command, args []string) {
	data, err := ioutil.ReadFile(args[0])
	if err != nil {
		log.Fatal(err)
	}

	// Load the OpenAPI document.
	loader := openapi3.NewSwaggerLoader()
	var swagger *openapi3.Swagger
	if strings.HasSuffix(args[0], ".yaml") || strings.HasSuffix(args[0], ".yml") {
		swagger, err = loader.LoadSwaggerFromYAMLData(data)
	} else {
		swagger, err = loader.LoadSwaggerFromData(data)
	}
	if err != nil {
		log.Fatal(err)
	}

	funcs := template.FuncMap{
		"escapeStr": escapeString,
		"slug":      slug,
	}

	data, _ = Asset("templates/commands.tmpl")
	tmpl, err := template.New("cli").Funcs(funcs).Parse(string(data))
	if err != nil {
		panic(err)
	}

	shortName := strings.TrimSuffix(path.Base(args[0]), ".yaml")

	templateData := ProcessAPI(shortName, swagger)

	var sb strings.Builder
	err = tmpl.Execute(&sb, templateData)
	if err != nil {
		panic(err)
	}

	writeFormattedFile(shortName+".go", []byte(sb.String()))
}

func main() {
	root := &cobra.Command{}

	root.AddCommand(&cobra.Command{
		Use:   "init <app-name>",
		Short: "Initialize and generate a `main.go` file for your project",
		Args:  cobra.ExactArgs(1),
		Run:   initCmd,
	})

	root.AddCommand(&cobra.Command{
		Use:   "generate <api-spec>",
		Short: "Generate a `commands.go` file from an OpenAPI spec",
		Args:  cobra.ExactArgs(1),
		Run:   generate,
	})

	root.Execute()
}
