package cli

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/olekukonko/tablewriter"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"os"
	"reflect"
	"strings"
	"time"
)

type TokenPayload struct {
	ExpiresIn    int    `mapstructure:"expires_in"`
	RefreshToken string `mapstructure:"refresh_token"`
	AccessToken  string `mapstructure:"access_token"`
	IDToken      string `mapstructure:"id_token"`
	Scope        string `mapstructure:"scope"`
	TokenType    string `mapstructure:"token_type"`
}

func (tp TokenPayload) ToMap() map[string]interface{} {
	m := make(map[string]interface{})
	m["expires_in"] = tp.ExpiresIn
	m["refresh_token"] = tp.RefreshToken
	m["access_token"] = tp.AccessToken
	m["expires_in"] = tp.ExpiresIn
	m["id_token"] = tp.IDToken
	m["scope"] = tp.Scope
	m["token_type"] = tp.TokenType
	return m
}

func (tp TokenPayload) ExpiresAt() time.Time {
	token, _, _ := new(jwt.Parser).ParseUnverified(tp.AccessToken, jwt.MapClaims{})
	if token == nil {
		return time.Time{}
	}
	claims, _ := token.Claims.(jwt.MapClaims)
	iat, _ := claims["iat"].(int)
	return time.Unix(int64(iat), 0)
}

func (tp TokenPayload) Issuer() string {
	token, _, _ := new(jwt.Parser).ParseUnverified(tp.AccessToken, jwt.MapClaims{})
	claims, _ := token.Claims.(jwt.MapClaims)
	iss := claims["iss"].(string)
	return iss
}

func (tp TokenPayload) ClientID() string {
	token, _, _ := new(jwt.Parser).ParseUnverified(tp.AccessToken, jwt.MapClaims{})
	claims, _ := token.Claims.(jwt.MapClaims)
	cid := claims["cid"].(string)
	return cid
}

type Credentials struct {
	TokenPayload TokenPayload `mapstructure:"token_payload"`
}

type Secrets struct {
	Credentials map[string]Credentials `mapstructure:"credentials"`
}

type VerbosityType string

const VerbosityTypePanic VerbosityType = "panic"
const VerbosityTypeFatal VerbosityType = "fatal"
const VerbosityTypeError VerbosityType = "error"
const VerbosityTypeWarn VerbosityType = "warn"
const VerbosityTypeInfo VerbosityType = "info"
const VerbosityTypeDebug VerbosityType = "debug"

type CLI struct {
	Verbosity    VerbosityType `mapstructure:"verbosity"`
	OutputFormat string        `mapstructure:"output_format"`
	Query        string        `mapstructure:"query"`
	Raw          bool          `mapstructure:"raw"`
}

func (c CLI) ZeroLogLevel() zerolog.Level {
	switch c.Verbosity {
	case VerbosityTypePanic:
		return zerolog.PanicLevel
	case VerbosityTypeFatal:
		return zerolog.FatalLevel
	case VerbosityTypeError:
		return zerolog.ErrorLevel
	case VerbosityTypeWarn:
		return zerolog.WarnLevel
	case VerbosityTypeInfo:
		return zerolog.InfoLevel
	case VerbosityTypeDebug:
		return zerolog.DebugLevel
	}
	return zerolog.GlobalLevel()
}

type AuthServer struct {
	ClientID string   `mapstructure:"client_id"`
	Issuer   string   `mapstructure:"issuer"`
	Keys     []string `mapstructure:"keys"`
	ListKeys []string `mapstructure:"list_keys"`
	Scopes []string `mapstructure:"scopes"`
}

type Applications struct {
	CLI CLI `mapstructure:"cli"`
}

type Profile struct {
	ApiURL          string `mapstructure:"api_url"`
	AuthServerName  string `mapstructure:"auth_server_name"`
	CredentialsName string `mapstructure:"credentials_name"`
	Applications `mapstructure:"applications"`
}

type Settings struct {
	DefaultProfileName string                `mapstructure:"default_profile_name"`
	Profiles           map[string]Profile    `mapstructure:"profiles"`
	AuthServers        map[string]AuthServer `mapstructure:"auth_servers"`
	viper              *viper.Viper
}

type ClientConfiguration struct {
	ProfileName  string   `mapstructure:"profile_name"`
	Secrets      Secrets  `mapstructure:"secrets"`
	Settings     Settings `mapstructure:"settings"`
	secretsPath  string
	settingsPath string
	globalFlags []GlobalFlag
}

func (cc ClientConfiguration) GetProfile() Profile {
	return cc.Settings.Profiles[cc.ProfileName]
}

func (cc ClientConfiguration) GetAuthServer() AuthServer {
	return cc.Settings.AuthServers[cc.GetProfile().AuthServerName]
}

func (cc ClientConfiguration) GetCredentials() Credentials {
	return cc.Secrets.Credentials[cc.GetProfile().CredentialsName]
}

func (cc *ClientConfiguration) bindGlobalFlags() (err error) {
	for _, globalFlag := range cc.globalFlags {
		err = globalFlag.bindFlag(cc.Settings.viper, cc.Settings)
		if err != nil {
			return
		}
	}

	var settings Settings
	err = cc.Settings.viper.Unmarshal(&settings)
	if err != nil {
		return
	}
	cc.Settings = settings
	return
}

func loadSecrets(envPrefix, secretsFilePath string) (secrets Secrets, err error) {
	touchFile(secretsFilePath)

	v := viper.New()

	v.SetEnvPrefix(fmt.Sprintf("%s_SECRETS", envPrefix))

	v.SetConfigFile(secretsFilePath)
	err = v.ReadInConfig()
	if err != nil {
		return
	}

	v.AutomaticEnv()

	err = v.Unmarshal(&secrets)
	if err != nil {
		return
	}

	if secrets.Credentials == nil {
		secrets.Credentials = make(map[string]Credentials)
	}

	return secrets, nil
}

func loadSettings(envPrefix, settingsFilePath string) (settings Settings, err error) {
	touchFile(settingsFilePath)

	v := viper.New()

	v.SetEnvPrefix(fmt.Sprintf("%s_SETTINGS", envPrefix))

	v.SetConfigFile(settingsFilePath)
	err = v.ReadInConfig()
	if err != nil {
		return
	}

	v.AutomaticEnv()

	err = v.Unmarshal(&settings)
	if err != nil {
		return
	}

	if settings.AuthServers == nil {
		settings.AuthServers = make(map[string]AuthServer)
	}
	if settings.Profiles == nil {
		settings.Profiles = make(map[string]Profile)
	}
	settings.viper = v
	return settings, nil
}

func touchFile(fileName string) error {
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		file, err := os.Create(fileName)
		if err != nil {
			return err
		}
		defer file.Close()
	}
	return nil
}

func InitConfiguration(envPrefix, settingsFilePath, secretsFilePath string, globalFlags []GlobalFlag) (err error) {
	clientConfiguration, err := LoadConfiguration(envPrefix, settingsFilePath, secretsFilePath, globalFlags)
	if err != nil {
		return
	}
	zerolog.SetGlobalLevel(clientConfiguration.GetProfile().CLI.ZeroLogLevel())
	RunConfig = clientConfiguration
	return
}

// LoadConfiguration loads secret and settings files. It will additional override those persisted values
// with (1) environment variables and (2) flag values (in order of increasing precedence).
func LoadConfiguration(envPrefix, settingsFilePath, secretsFilePath string, globalFlags []GlobalFlag) (config ClientConfiguration, err error) {
	secrets, err := loadSecrets(envPrefix, secretsFilePath)
	if err != nil {
		return
	}

	settings, err := loadSettings(envPrefix, settingsFilePath)
	if err != nil {
		return
	}

	config = ClientConfiguration{
		Secrets:     secrets,
		secretsPath: secretsFilePath,
		Settings:    settings,
		settingsPath: settingsFilePath,
		globalFlags: globalFlags,
	}

	err = config.bindGlobalFlags()
	if err != nil {
		return
	}
	config.ProfileName = config.Settings.DefaultProfileName

	return
}

func (cc *ClientConfiguration) UpdateCredentialsToken(credentialsName string, token *oauth2.Token) error {
	tokenPayload := cc.Secrets.Credentials[credentialsName].TokenPayload
	tokenPayload.AccessToken = token.AccessToken
	tokenPayload.TokenType = token.TokenType
	if token.RefreshToken != "" {
		tokenPayload.RefreshToken = token.RefreshToken
	}
	credentials := cc.Secrets.Credentials[credentialsName]
	credentials.TokenPayload = tokenPayload
	cc.Secrets.Credentials[credentialsName] = credentials

	updates := make(map[string]interface{})
	updates[fmt.Sprintf("credentials.%s.token_payload.access_token", credentialsName)] = tokenPayload.AccessToken
	updates[fmt.Sprintf("credentials.%s.token_payload.refresh_token", credentialsName)] = tokenPayload.RefreshToken
	updates[fmt.Sprintf("credentials.%s.token_payload.id_token", credentialsName)] = tokenPayload.IDToken
	updates[fmt.Sprintf("credentials.%s.token_payload.token_type", credentialsName)] = tokenPayload.TokenType

	return cc.writeSecrets(updates)
}

func (cc ClientConfiguration) writeSettings(updates map[string]interface{}) (err error) {
	return cc.write(cc.settingsPath, updates)
}

func (cc ClientConfiguration) writeSecrets(updates map[string]interface{}) (err error) {
	return cc.write(cc.secretsPath, updates)
}

func (cc ClientConfiguration) write(filePath string, updates map[string]interface{}) (err error) {
	v := viper.New()

	v.SetConfigFile(filePath)
	err = v.ReadInConfig()
	if err != nil {
		return
	}

	for path, value := range updates {
		v.Set(path, value)
	}

	err = v.WriteConfig()
	return
}

func BuildSettingsCommands() (configCommand *cobra.Command) {
	configCommand = &cobra.Command{
		Use:   "settings",
		Short: "Interact with your settings.toml file",
	}
	configCommand.AddCommand(
		buildSettingsAddAuthServerCommand(),
		buildSettingsListAuthServersCommand(),
		buildSettingsGetCommand(), 
		buildSettingsSetCommand())

	return
}

func buildSettingsAddAuthServerCommand() (cmd *cobra.Command) {
	var clientID string
	var issuer string
	cmd = &cobra.Command{
		Use:   "add-auth-server",
		Short: "Add a new authentication server",
		Args:  cobra.ExactArgs(1),
		Run:  func(cmd *cobra.Command, args []string) {
			logger := log.With().Str("profile", RunConfig.ProfileName).Logger()

			authServerName := strings.Replace(args[0], ".", "-", -1)
			_, exists := RunConfig.Settings.AuthServers[authServerName]
			if exists {
				logger.Fatal().Msgf("credential %q already exists", authServerName)
			}

			updates := make(map[string]interface{})
			updates[fmt.Sprintf("auth_servers.%s.issuer", authServerName)] = issuer
			updates[fmt.Sprintf("auth_servers.%s.client_id", authServerName)] = clientID
			err := RunConfig.write(RunConfig.settingsPath, updates)
			if err != nil {
				logger.Fatal().Err(err).Msg("Failed to write updated settings")
			}
		},
	}
	cmd.Flags().StringVar(&clientID, "client-id", "", "")
	cmd.Flags().StringVar(&issuer, "issuer", "", "")

	return
}

func buildSettingsListAuthServersCommand() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "list-auth-servers",
		Short:   "List available authentication servers",
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			authServers := RunConfig.Settings.AuthServers
			if authServers != nil {
				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"Name", "Client ID", "Issuer"})

				// For each type name, draw a table with the relevant profileName keys
				for authServerName, authServer := range authServers {
					table.Append([]string{authServerName, authServer.ClientID, authServer.Issuer})
				}
				table.Render()
			} else {
				fmt.Printf("No authentication servers configured. Use `%s auth addServer` to add one.\n", Root.CommandPath())
			}
		},
	}
	return
}

// WARNING: This does not support array indices in its current implementation.
func runConfig(filePath string, topLevel interface{}, args []string) {
	logger := log.With().Logger()

	path := args[0]
	if len(args) == 1 {
		currentValue, err := getValueFromPath(topLevel, path)
		if err != nil {
			logger.Fatal().Err(err).Msgf("could not find value at path %q", path)
			return
		}
		fmt.Printf("%v\n", currentValue.Interface())
		return
	}

	reflectType, err := getTypeFromPath(reflect.TypeOf(topLevel), path)
	if err != nil {
		logger.Fatal().Err(err).Msgf("%q is not a valid path", path)
	}
	valueString := args[1]
	value, err := parseNewValue(valueString, reflectType)
	if err != nil {
		logger.Fatal().Err(err).Msgf("an error occurred parsing value %q", valueString)
	}

	updates := make(map[string]interface{})
	updates[path] = value
	err = RunConfig.write(filePath, updates)
	if err != nil {
		logger.Fatal().Err(err).Msgf("an error occurred writing updates to %q", filePath)
	}
}

func buildSettingsGetCommand() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "get",
		Short: "Get a value from settings.toml",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runConfig(RunConfig.settingsPath, RunConfig.Settings, args)
		},
	}
	return
}
func buildSettingsSetCommand() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "set",
		Short: "Set a value in settings.toml",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			runConfig(RunConfig.settingsPath, RunConfig.Settings, args)
		},
	}
	return
}

var errTagNotFound = errors.New("tag not found")
var errUnsupportedTag = errors.New("unsupported tag")

func getValueOfTagField(t interface{}, tag string) (parsed interface{}, err error) {
	parsed = nil
	if t == nil {
		return
	}
	interfaceType := reflect.TypeOf(t)
	interfaceValue := reflect.ValueOf(t)
	if interfaceType.Kind() == reflect.Ptr {
		interfaceType = interfaceType.Elem()
		interfaceValue = interfaceValue.Elem()
	}
	if interfaceType.Kind() == reflect.Map {
		if interfaceValue.IsZero() {
			err = errTagNotFound
			return
		}
		elemValue := interfaceValue.MapIndex(reflect.ValueOf(tag))
		if !elemValue.IsValid() {
			err = errTagNotFound
			return
		}
		parsed = elemValue.Interface()
		return
	}
	if interfaceValue.Kind() != reflect.Struct {
		err = fmt.Errorf("unsupported interface kind %s", interfaceValue.Kind())
		return
	}
	for i := 0; i < interfaceType.NumField(); i++ {
		field := interfaceType.Field(i)
		val, ok := field.Tag.Lookup("mapstructure")
		if !ok {
			continue
		}
		firstTagValue := strings.Split(val, ",")[0]
		if firstTagValue == toSnakeCase(tag) {
			fieldValue := interfaceValue.Field(i)
			parsed = fieldValue.Interface()
			return
		}
	}
	err = errTagNotFound
	return
}

func getTypeOfTagField(interfaceType reflect.Type, tag string) (reflectType reflect.Type, err error) {
	if interfaceType.Kind() == reflect.Ptr {
		interfaceType = interfaceType.Elem()
	}
	if interfaceType.Kind() == reflect.Map {
		reflectType = interfaceType.Elem()
		return
	}
	if interfaceType.Kind() != reflect.Struct {
		err = fmt.Errorf("unsupported interface kind %s", interfaceType.Kind())
		return
	}
	for i := 0; i < interfaceType.NumField(); i++ {
		field := interfaceType.Field(i)
		val, ok := field.Tag.Lookup("mapstructure")
		if !ok {
			continue
		}
		firstTagValue := strings.Split(val, ",")[0]
		if firstTagValue == toSnakeCase(tag) {
			reflectType = field.Type
			return
		}
	}
	return nil, errTagNotFound
}


func getValueFromPath(value interface{}, path string) (reflectValue reflect.Value, err error) {
	parts := strings.Split(path, ".")

	cursor := value
	for _, part := range parts {
		cursor, err = getValueOfTagField(cursor, part)
		if err != nil {
			reflectValue = reflect.ValueOf(cursor)
			return
		}
	}
	reflectValue = reflect.ValueOf(cursor)
	return
}

func getTypeFromPath(parent reflect.Type, path string) (reflectType reflect.Type, err error) {
	parts := strings.Split(path, ".")

	reflectType = parent
	for _, part := range parts {
		reflectType, err = getTypeOfTagField(reflectType, part)
		if err != nil {
			return
		}
	}
	return
}

func parseNewValue(newValue interface{}, reflectType reflect.Type) (parsed interface{}, err error) {
	switch reflectType.Kind() {
	case reflect.Int:
		return cast.ToIntE(newValue)
	case reflect.Float64:
		return cast.ToFloat64E(newValue)
	case reflect.String:
		return cast.ToString(newValue), nil
	case reflect.Bool:
		return cast.ToBoolE(newValue)
	default:
		return nil, errUnsupportedTag
	}
}
