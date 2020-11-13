package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// execute a command against the configured CLI
func execute(cmd string) string {
	out := new(bytes.Buffer)
	Root.SetArgs(strings.Split(cmd, " "))
	Root.SetOutput(out)
	Stdout = out
	Stderr = out
	Root.Execute()
	return out.String()
}

func TestInit(t *testing.T) {
	Client = nil
	Root = nil

	viper.Set("color", true)

	Init(&Config{
		AppName: "test",
	})

	assert.NotNil(t, Client)
	assert.NotNil(t, Root)
}

func TestHelpCommands(t *testing.T) {
	Init(&Config{
		AppName: "test",
	})

	out := execute("help-config")
	assert.Contains(t, out, "CLI Configuration")

	out = execute("help-input")
	assert.Contains(t, out, "CLI Request Input")
}

func TestPreRun(t *testing.T) {
	Init(&Config{
		AppName: "test",
	})

	ran := false
	PreRun = func(cmd *cobra.Command, args []string) error {
		ran = true
		return nil
	}

	Root.Run = func(cmd *cobra.Command, args []string) {
		// Do nothing, but also don't error.
	}

	execute("")

	assert.True(t, ran)
}
