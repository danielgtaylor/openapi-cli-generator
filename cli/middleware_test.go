package cli

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	gentleman "gopkg.in/h2non/gentleman.v2"
)

func TestCommandMiddleware(t *testing.T) {
	root := &cobra.Command{
		Use: "main",
	}

	foo := &cobra.Command{
		Use: "foo arg1 arg2",
	}

	bar := &cobra.Command{
		Use: "bar arg3",
		Run: func(cmd *cobra.Command, args []string) {
			HandleBefore(cmd, nil, nil)
			HandleAfter(cmd, nil, nil, nil)
		},
	}

	root.AddCommand(foo)
	foo.AddCommand(bar)

	before := false
	after := false

	RegisterBefore("foo bar", func(cmd *cobra.Command, params *viper.Viper, r *gentleman.Request) {
		before = true
	})

	RegisterAfter("foo bar", func(cmd *cobra.Command, param *viper.Viper, resp *gentleman.Response, data interface{}) interface{} {
		after = true
		return data
	})

	bar.Run(bar, []string{})

	assert.Equal(t, true, before)
	assert.Equal(t, true, after)
}
