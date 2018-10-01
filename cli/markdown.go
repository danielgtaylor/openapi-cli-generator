package cli

import (
	"strings"

	"github.com/alecthomas/chroma/quick"
)

// Markdown renders terminal-friendly Markdown content.
func Markdown(content string) string {
	if !tty {
		return content
	}

	// For now, just use syntax highlighting. Later, we might replace this
	// with something to render nice tables and lists, replace formatting
	// characters, etc.
	builder := strings.Builder{}
	if err := quick.Highlight(&builder, content, "markdown", "terminal256", "cli-dark"); err != nil {
		return content
	}

	return builder.String()
}
