package cli

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)


func TestConfig(t *testing.T) {
	clientConfiguration := ClientConfiguration{
		ProfileName:  "john",
		Secrets:      Secrets{
			Credentials: map[string]Credentials{
				"A": {
					TokenPayload: TokenPayload{
						RefreshToken: "refresh1",
						AccessToken:  "access1",
						TokenType:    "Bearer",
					},
				},
			},
		},
		Settings:     Settings{
			DefaultProfileName: "default",
			Profiles:           map[string]Profile{
				"tester": {
					ApiURL: "https://api.dev.qcs.rigetti.com",
				},
			},
			AuthServers:        nil,
		},
		secretsPath:  "",
		settingsPath: "",
	}


	t.Run("test getValueOfTagField", func(t *testing.T) {
		t.Parallel()

		val, err := getValueOfTagField(clientConfiguration.Settings, "profiles")
		assert.Nil(t, err)

		profiles, ok := val.(map[string]Profile)
		if !ok {
			t.Fatal("profiles is not a map[string]Profile")
		}
		assert.Equal(t, 1, len(profiles))
	})

	t.Run("test getValueFromPath success", func(t *testing.T) {
		t.Parallel()

		value, err := getValueFromPath(clientConfiguration.Settings, "profiles.tester.api_url")
		assert.Nil(t, err)
		assert.Equal(t, clientConfiguration.Settings.Profiles["tester"].ApiURL, value.String())

		value, err = getValueFromPath(clientConfiguration.Secrets, "credentials.A.token_payload.access_token")
		assert.Nil(t, err)
		assert.Equal(t, clientConfiguration.Secrets.Credentials["A"].TokenPayload.AccessToken, value.String())
	})

	t.Run("test getValueFromPath map key doesnt exist", func(t *testing.T) {
		t.Parallel()

		_, err := getValueFromPath(clientConfiguration.Settings, "profiles.doesntexist.api_url")
		assert.Equal(t, errTagNotFound, err)

		_, err = getValueFromPath(clientConfiguration.Secrets, "credentials.doesntexist.access_token")
		assert.Equal(t, errTagNotFound, err)
	})

	t.Run("test getValueFromPath struct field doesnt exist", func(t *testing.T) {
		t.Parallel()

		_, err := getValueFromPath(clientConfiguration.Settings, "profiles.doesntexist.doesntexist")
		assert.Equal(t, errTagNotFound, err)

		_, err = getValueFromPath(clientConfiguration.Secrets, "credentials.A.token_payload.doesntexist")
		assert.NotNil(t, errTagNotFound, err)
	})

	t.Run("test getValueFromPath struct field doesnt deep", func(t *testing.T) {
		t.Parallel()

		value, err := getValueFromPath(clientConfiguration.Settings, "profiles.doesntexist.applications.cli.verbosity")
		assert.Equal(t, errTagNotFound, err)
		assert.Equal(t, reflect.ValueOf(nil), value)
	})

	t.Run("test getTypeFromPath struct field doesnt deep", func(t *testing.T) {
		t.Parallel()

		reflectType, err := getTypeFromPath(reflect.TypeOf(clientConfiguration.Settings), "profiles.doesntexist.applications.cli.verbosity")
		assert.Nil(t, err)
		assert.Equal(t, reflect.TypeOf(VerbosityTypeDebug), reflectType)
	})

	t.Run("test parseNewValue success", func(t *testing.T) {
		t.Parallel()

		value, err := getValueFromPath(clientConfiguration.Settings, "profiles.tester.api_url")
		assert.Nil(t, err)
		assert.Equal(t, clientConfiguration.Settings.Profiles["tester"].ApiURL, value.String())

		reflectType, err := getTypeFromPath(reflect.TypeOf(clientConfiguration.Settings), "profiles.P.api_url")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, reflect.String, reflectType.Kind())

		newValue, err := parseNewValue("https://www.update.com", reflectType)
		assert.Nil(t, err)
		val, ok := newValue.(string)
		assert.True(t, ok)
		assert.Equal(t, "https://www.update.com", val)
	})
}
