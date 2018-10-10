package cli

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestCustomFlags(t *testing.T) {
	root := &cobra.Command{
		Use: "main",
	}

	cmd := &cobra.Command{
		Use: "test",
	}

	root.AddCommand(cmd)

	AddFlag("test", "bool", "b", "description", false)
	AddFlag("test", "int", "i", "description", 0)
	AddFlag("test", "float", "f", "description", 0.0)
	AddFlag("test", "string", "s", "description", "")

	SetCustomFlags(cmd)

	assert.NotNil(t, cmd.Flags().Lookup("bool"))
	assert.NotNil(t, cmd.Flags().Lookup("int"))
	assert.NotNil(t, cmd.Flags().Lookup("float"))
	assert.NotNil(t, cmd.Flags().Lookup("string"))
}
