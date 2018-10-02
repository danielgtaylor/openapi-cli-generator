package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/quick"
	"github.com/alecthomas/chroma/styles"
	jmespath "github.com/jmespath/go-jmespath"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

func init() {
	// Simple 256-color theme for JSON/YAML output in a terminal.
	styles.Register(chroma.MustNewStyle("cli-dark", chroma.StyleEntries{
		// Used for JSON/YAML
		chroma.Text:        "#b2b2b2",
		chroma.Comment:     "#9e9e9e",
		chroma.Keyword:     "#ff5f87",
		chroma.Punctuation: "#9e9e9e",
		chroma.NameTag:     "#5fafd7",
		chroma.Number:      "#d78700",
		chroma.String:      "#afd787",

		// Used for Markdown
		chroma.GenericHeading:    "#5fafd7",
		chroma.GenericSubheading: "#5fafd7",
		chroma.GenericEmph:       "italic #756ac1",
		chroma.GenericStrong:     "bold #f1ea83",
		chroma.GenericDeleted:    "#3e3e3e",
		chroma.NameAttribute:     "underline",
	}))
}

// ResponseFormatter will filter, prettify, and print out the results of a call.
type ResponseFormatter interface {
	Format(interface{}) error
}

// DefaultFormatter can apply JMESPath queries and can output prettyfied JSON
// and YAML output. If Stdout is a TTY, then colorized output is provided. The
// default formatter uses the `query` and `output-format` configuration
// values to perform JMESPath queries and set JSON (default) or YAML output.
type DefaultFormatter struct {
	tty bool
}

// NewDefaultFormatter creates a new formatted with autodetected TTY
// capabilities.
func NewDefaultFormatter(tty bool) *DefaultFormatter {
	return &DefaultFormatter{
		tty: tty,
	}
}

// Format will filter, prettify, colorize and output the data.
func (f *DefaultFormatter) Format(data interface{}) error {
	if data == nil {
		return nil
	}

	if viper.GetString("query") != "" {
		result, err := jmespath.Search(viper.GetString("query"), data)

		if err != nil {
			return err
		}

		data = result
	}

	// Encode to the requested output format using nice formatting.
	var encoded []byte
	var err error
	var lexer string

	if dStr, ok := data.(string); ok && viper.GetBool("raw") {
		encoded = []byte(dStr)
		lexer = ""

		if len(dStr) != 0 && (dStr[0] == '{' || dStr[0] == '[') {
			// Looks like JSON to me!
			lexer = "json"
		}
	} else {
		if viper.GetString("output-format") == "yaml" {
			encoded, err = yaml.Marshal(data)

			if err != nil {
				return err
			}

			lexer = "yaml"
		} else {
			encoded, err = json.MarshalIndent(data, "", "  ")

			if err != nil {
				return err
			}

			lexer = "json"
		}
	}

	// Make sure we end with a newline, otherwise things won't look right
	// in the terminal.
	if encoded[len(encoded)-1] != '\n' {
		encoded = append(encoded, '\n')
	}

	// Only colorize if we are a TTY.
	if f.tty {
		if err = quick.Highlight(os.Stdout, string(encoded), lexer, "terminal256", "cli-dark"); err != nil {
			return err
		}
	} else {
		fmt.Println(string(encoded))
	}

	return nil
}
