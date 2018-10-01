package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"sort"
	"strings"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/spf13/cobra"
)

type Param struct {
	Name        string
	GoName      string
	Description string
	In          string
	Required    bool
	Type        string
	TypeNil     string
}

type Operation struct {
	Use            string
	Short          string
	Long           string
	Method         string
	CanHaveBody    bool
	Path           string
	AllParams      []*Param
	RequiredParams []*Param
	OptionalParams []*Param
	MediaType      string
}

type Server struct {
	Description string
	URL         string
	// TODO: handle server parameters
}

type OpenAPI struct {
	Name        string
	GoName      string
	Title       string
	Description string
	Servers     []*Server
	Operations  []*Operation
}

func ProcessAPI(shortName string, api *openapi3.Swagger) *OpenAPI {
	result := &OpenAPI{
		Name:        shortName,
		GoName:      toGoName(shortName, false),
		Title:       api.Info.Title,
		Description: escapeString(api.Info.Description),
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
		for method, operation := range item.Operations() {
			params := getParams(item, method)
			requiredParams := getRequiredParams(params)
			optionalParams := getOptionalParams(params)
			short := operation.Summary
			if short == "" {
				short = operation.OperationID
			}

			method := strings.Title(strings.ToLower(method))

			o := &Operation{
				Use:            usage(operation.OperationID, requiredParams),
				Short:          short,
				Long:           escapeString(operation.Description),
				Method:         method,
				CanHaveBody:    method == "Post" || method == "Put" || method == "Patch",
				Path:           path,
				AllParams:      params,
				RequiredParams: requiredParams,
				OptionalParams: optionalParams,
				MediaType:      getRequestMediaType(operation),
			}

			result.Operations = append(result.Operations, o)
		}
	}

	return result
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
			if p.Value.Schema.Value != nil && p.Value.Schema.Value.Type != "" {
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

			allParams = append(allParams, &Param{
				Name:        p.Value.Name,
				GoName:      toGoName("param "+p.Value.Name, false),
				Description: p.Value.Description,
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

func getRequestMediaType(op *openapi3.Operation) string {
	if op.RequestBody != nil && op.RequestBody.Value != nil {
		for m := range op.RequestBody.Value.Content {
			if strings.Contains(m, "json") {
				return m
			}
		}
	}

	return ""
}

func initCmd(cmd *cobra.Command, args []string) {
	tmplData, err := ioutil.ReadFile("templates/main.tmpl")
	if err != nil {
		panic(err)
	}

	tmpl, err := template.New("cli").Parse(string(tmplData))
	if err != nil {
		panic(err)
	}

	templateData := map[string]string{
		"Name":    args[0],
		"NameEnv": strings.Replace(strings.ToUpper(args[0]), "-", "_", -1),
	}

	err = tmpl.Execute(os.Stdout, templateData)
	if err != nil {
		panic(err)
	}
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

	tmplData, err := ioutil.ReadFile("templates/commands.tmpl")
	if err != nil {
		panic(err)
	}

	funcs := template.FuncMap{
		"escapeStr": escapeString,
		"slug":      slug,
	}

	tmpl, err := template.New("cli").Funcs(funcs).Parse(string(tmplData))
	if err != nil {
		panic(err)
	}

	shortName := strings.TrimSuffix(path.Base(args[0]), ".yaml")

	templateData := ProcessAPI(shortName, swagger)

	err = tmpl.Execute(os.Stdout, templateData)
	if err != nil {
		panic(err)
	}
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
