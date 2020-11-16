package cli

import (
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

func TestInitCredentials(t *testing.T) {
	dir := path.Join(
		os.Getenv("PWD"),
		"__fixtures__")
	origCreds := Creds
	origProfile := viper.Get("profile_name")
	defer func() {
		Creds = origCreds
		viper.Set("profile_name", origProfile)
	}()
	credentials := initCredentialsFrom(dir, "credentials", "toml")
	Creds = &credentials

	viper.Set("profile_name", "default")
	assert.Equal(t, "", GetActiveProfile().Info.AuthServerName)

	viper.Set("profile_name", "cow")
	assert.Equal(t, "farm", GetActiveProfile().Info.AuthServerName)

	viper.Set("profile_name", "dog")
	assert.Equal(t, "home", GetActiveProfile().Info.AuthServerName)
}
