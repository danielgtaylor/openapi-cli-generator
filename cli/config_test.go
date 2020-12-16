package cli

import (
	"github.com/stretchr/testify/assert"
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
				"default": {
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

		value, err := getValueFromPath(clientConfiguration.Settings, "profiles.default.api_url")
		assert.Nil(t, err)
		assert.Equal(t, clientConfiguration.Settings.Profiles["default"].ApiURL, value.String())

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

		_, err := getValueFromPath(clientConfiguration.Settings, "profiles.default.doesntexist")
		assert.Equal(t, errTagNotFound, err)

		_, err = getValueFromPath(clientConfiguration.Secrets, "credentials.A.token_payload.doesntexist")
		assert.NotNil(t, errTagNotFound, err)
	})

	t.Run("test parseNewValue success", func(t *testing.T) {
		t.Parallel()

		value, err := getValueFromPath(clientConfiguration.Settings, "profiles.default.api_url")
		assert.Nil(t, err)
		assert.Equal(t, clientConfiguration.Settings.Profiles["default"].ApiURL, value.String())

		newValue, err := parseNewValue("https://www.update.com", value)
		assert.Nil(t, err)
		val, ok := newValue.(string)
		assert.True(t, ok)
		assert.Equal(t, "https://www.update.com", val)
	})
}
