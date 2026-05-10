package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/pafthang/servicebase/tools/auth"
	"github.com/pafthang/servicebase/tools/cron"
	"github.com/pafthang/servicebase/tools/mailer"
	"github.com/pafthang/servicebase/tools/rest"
	"github.com/pafthang/servicebase/tools/security"
)

// SecretMask is the default settings secrets replacement value
// (see Settings.RedactClone()).
const SecretMask string = "******"

// Settings defines common app configuration options.
type Settings struct {
	mux sync.RWMutex

	Meta     MetaConfig     `form:"meta" json:"meta"`
	Logs     LogsConfig     `form:"logs" json:"logs"`
	Smtp     SmtpConfig     `form:"smtp" json:"smtp"`
	S3       S3Config       `form:"s3" json:"s3"`
	Backups  BackupsConfig  `form:"backups" json:"backups"`
	AI       AIConfig       `form:"ai" json:"ai"`
	Calendar CalendarConfig `form:"calendar" json:"calendar"`
	Mail     MailConfig     `form:"mail" json:"mail"`
	Weather  WeatherConfig  `form:"weather" json:"weather"`

	RecordAuthToken          TokenConfig `form:"recordAuthToken" json:"recordAuthToken"`
	RecordPasswordResetToken TokenConfig `form:"recordPasswordResetToken" json:"recordPasswordResetToken"`
	RecordEmailChangeToken   TokenConfig `form:"recordEmailChangeToken" json:"recordEmailChangeToken"`
	RecordVerificationToken  TokenConfig `form:"recordVerificationToken" json:"recordVerificationToken"`
	RecordFileToken          TokenConfig `form:"recordFileToken" json:"recordFileToken"`

	EmailAuth EmailAuthConfig `form:"emailAuth" json:"emailAuth"`

	GithubAuth AuthProviderConfig `form:"githubAuth" json:"githubAuth"`

	VKAuth     AuthProviderConfig `form:"vkAuth" json:"vkAuth"`
	YandexAuth AuthProviderConfig `form:"yandexAuth" json:"yandexAuth"`
}

// New creates and returns a new default Settings instance.
func NewSettings() *Settings {
	return &Settings{
		Meta: MetaConfig{
			AppName:                    "Acme",
			AppUrl:                     "http://localhost:8090",
			HideControls:               false,
			SenderName:                 "Support",
			SenderAddress:              "support@example.com",
			VerificationTemplate:       defaultVerificationTemplate,
			ResetPasswordTemplate:      defaultResetPasswordTemplate,
			ConfirmEmailChangeTemplate: defaultConfirmEmailChangeTemplate,
		},
		Logs: LogsConfig{
			MaxDays: 5,
			LogIp:   true,
		},
		Smtp: SmtpConfig{
			Enabled:  false,
			Host:     "smtp.example.com",
			Port:     587,
			Username: "",
			Password: "",
			Tls:      false,
		},
		Backups: BackupsConfig{
			CronMaxKeep: 3,
		},
		AI: AIConfig{
			Service: AIServiceNone,
		},
		Calendar: CalendarConfig{},
		Mail:     MailConfig{},
		Weather: WeatherConfig{
			Units: "metric",
			Lang:  "en",
		},
		RecordAuthToken: TokenConfig{
			Secret:   security.RandomString(50),
			Duration: 1209600, // 14 days
		},
		RecordPasswordResetToken: TokenConfig{
			Secret:   security.RandomString(50),
			Duration: 1800, // 30 minutes
		},
		RecordVerificationToken: TokenConfig{
			Secret:   security.RandomString(50),
			Duration: 604800, // 7 days
		},
		RecordFileToken: TokenConfig{
			Secret:   security.RandomString(50),
			Duration: 120, // 2 minutes
		},
		RecordEmailChangeToken: TokenConfig{
			Secret:   security.RandomString(50),
			Duration: 1800, // 30 minutes
		},
		GithubAuth: AuthProviderConfig{
			Enabled: false,
		},
		VKAuth: AuthProviderConfig{
			Enabled: false,
		},
		YandexAuth: AuthProviderConfig{
			Enabled: false,
		},
	}
}

// Validate makes Settings validatable by implementing [validation.Validatable] interface.
func (s *Settings) Validate() error {
	s.mux.Lock()
	defer s.mux.Unlock()

	return validation.ValidateStruct(s,
		validation.Field(&s.Meta),
		validation.Field(&s.Logs),
		validation.Field(&s.RecordAuthToken),
		validation.Field(&s.RecordPasswordResetToken),
		validation.Field(&s.RecordEmailChangeToken),
		validation.Field(&s.RecordVerificationToken),
		validation.Field(&s.RecordFileToken),
		validation.Field(&s.Smtp),
		validation.Field(&s.S3),
		validation.Field(&s.Backups),
		validation.Field(&s.AI),
		validation.Field(&s.Calendar),
		validation.Field(&s.Mail),
		validation.Field(&s.Weather),
		validation.Field(&s.GithubAuth),
		validation.Field(&s.VKAuth),
		validation.Field(&s.YandexAuth),
	)
}

// Merge merges `other` settings into the current one.
func (s *Settings) Merge(other *Settings) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	bytes, err := json.Marshal(other)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, s)
}

// Clone creates a new deep copy of the current settings.
func (s *Settings) Clone() (*Settings, error) {
	clone := &Settings{}
	if err := clone.Merge(s); err != nil {
		return nil, err
	}
	return clone, nil
}

// RedactClone creates a new deep copy of the current settings,
// while replacing the secret values with `******`.
func (s *Settings) RedactClone() (*Settings, error) {
	clone, err := s.Clone()
	if err != nil {
		return nil, err
	}

	sensitiveFields := []*string{
		&clone.Smtp.Password,
		&clone.S3.Secret,
		&clone.Backups.S3.Secret,
		&clone.AI.APIKey,
		&clone.AI.AnthropicKey,
		&clone.Calendar.ClientSecret,
		&clone.Mail.ClientSecret,
		&clone.Weather.APIKey,
		&clone.RecordAuthToken.Secret,
		&clone.RecordPasswordResetToken.Secret,
		&clone.RecordEmailChangeToken.Secret,
		&clone.RecordVerificationToken.Secret,
		&clone.RecordFileToken.Secret,
		&clone.GithubAuth.ClientSecret,
		&clone.VKAuth.ClientSecret,
		&clone.YandexAuth.ClientSecret,
	}

	// mask all sensitive fields
	for _, v := range sensitiveFields {
		if v != nil && *v != "" {
			*v = SecretMask
		}
	}

	return clone, nil
}

// NamedAuthProviderConfigs returns a map with all registered OAuth2
// provider configurations (indexed by their name identifier).
func (s *Settings) NamedAuthProviderConfigs() map[string]AuthProviderConfig {
	s.mux.RLock()
	defer s.mux.RUnlock()

	return map[string]AuthProviderConfig{
		auth.NameGithub: s.GithubAuth,
		auth.NameVK:     s.VKAuth,
		auth.NameYandex: s.YandexAuth,
	}
}

// -------------------------------------------------------------------

type TokenConfig struct {
	Secret   string `form:"secret" json:"secret"`
	Duration int64  `form:"duration" json:"duration"`
}

// Validate makes TokenConfig validatable by implementing [validation.Validatable] interface.
func (c TokenConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Secret, validation.Required, validation.Length(30, 300)),
		validation.Field(&c.Duration, validation.Required, validation.Min(5), validation.Max(63072000)),
	)
}

// -------------------------------------------------------------------

type SmtpConfig struct {
	Enabled  bool   `form:"enabled" json:"enabled"`
	Host     string `form:"host" json:"host"`
	Port     int    `form:"port" json:"port"`
	Username string `form:"username" json:"username"`
	Password string `form:"password" json:"password"`

	// SMTP AUTH - PLAIN (default) or LOGIN
	AuthMethod string `form:"authMethod" json:"authMethod"`

	// Whether to enforce TLS encryption for the mail server connection.
	//
	// When set to false StartTLS command is send, leaving the server
	// to decide whether to upgrade the connection or not.
	Tls bool `form:"tls" json:"tls"`

	// LocalName is optional domain name or IP address used for the
	// EHLO/HELO exchange (if not explicitly set, defaults to "localhost").
	//
	// This is required only by some SMTP servers, such as Gmail SMTP-relay.
	LocalName string `form:"localName" json:"localName"`
}

// Validate makes SmtpConfig validatable by implementing [validation.Validatable] interface.
func (c SmtpConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(
			&c.Host,
			validation.When(c.Enabled, validation.Required),
			is.Host,
		),
		validation.Field(
			&c.Port,
			validation.When(c.Enabled, validation.Required),
			validation.Min(0),
		),
		validation.Field(
			&c.AuthMethod,
			// don't require it for backward compatibility
			// (fallback internally to PLAIN)
			// validation.When(c.Enabled, validation.Required),
			validation.In(mailer.SmtpAuthLogin, mailer.SmtpAuthPlain),
		),
		validation.Field(&c.LocalName, is.Host),
	)
}

// -------------------------------------------------------------------

const (
	AIServiceNone   string = "none"
	AIServiceOpenAI string = "openai"
	AIServiceClaude string = "claude"
)

type AIConfig struct {
	Service      string `form:"service" json:"service"`
	Model        string `form:"model" json:"model"`
	APIKey       string `form:"apiKey" json:"apiKey"`
	AnthropicKey string `form:"anthropicKey" json:"anthropicKey"`
}

// Validate makes AIConfig validatable by implementing [validation.Validatable] interface.
func (c AIConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Service, validation.Required, validation.In(AIServiceNone, AIServiceOpenAI, AIServiceClaude)),
		validation.Field(&c.APIKey, validation.When(c.Service == AIServiceOpenAI, validation.Required)),
		validation.Field(&c.AnthropicKey, validation.When(c.Service == AIServiceClaude, validation.Required)),
	)
}

// -------------------------------------------------------------------

type CalendarConfig struct {
	ClientID     string `form:"clientId" json:"clientId"`
	ClientSecret string `form:"clientSecret" json:"clientSecret"`
	RedirectURL  string `form:"redirectUrl" json:"redirectUrl"`
}

// Validate makes CalendarConfig validatable by implementing [validation.Validatable] interface.
func (c CalendarConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(
			&c.RedirectURL,
			validation.When(c.RedirectURL != "", is.URL),
		),
	)
}

// -------------------------------------------------------------------

type MailConfig struct {
	ClientID     string `form:"clientId" json:"clientId"`
	ClientSecret string `form:"clientSecret" json:"clientSecret"`
	RedirectURL  string `form:"redirectUrl" json:"redirectUrl"`
}

// Validate makes MailConfig validatable by implementing [validation.Validatable] interface.
func (c MailConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(
			&c.RedirectURL,
			validation.When(c.RedirectURL != "", is.URL),
		),
	)
}

// -------------------------------------------------------------------

type WeatherConfig struct {
	APIKey string `form:"apiKey" json:"apiKey"`
	Units  string `form:"units" json:"units"`
	Lang   string `form:"lang" json:"lang"`
}

// Validate makes WeatherConfig validatable by implementing [validation.Validatable] interface.
func (c WeatherConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Units, validation.In("", "standard", "metric", "imperial")),
		validation.Field(&c.Lang, validation.Length(0, 10)),
	)
}

// -------------------------------------------------------------------

type S3Config struct {
	Enabled        bool   `form:"enabled" json:"enabled"`
	Bucket         string `form:"bucket" json:"bucket"`
	Region         string `form:"region" json:"region"`
	Endpoint       string `form:"endpoint" json:"endpoint"`
	AccessKey      string `form:"accessKey" json:"accessKey"`
	Secret         string `form:"secret" json:"secret"`
	ForcePathStyle bool   `form:"forcePathStyle" json:"forcePathStyle"`
}

// Validate makes S3Config validatable by implementing [validation.Validatable] interface.
func (c S3Config) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Endpoint, is.URL, validation.When(c.Enabled, validation.Required)),
		validation.Field(&c.Bucket, validation.When(c.Enabled, validation.Required)),
		validation.Field(&c.Region, validation.When(c.Enabled, validation.Required)),
		validation.Field(&c.AccessKey, validation.When(c.Enabled, validation.Required)),
		validation.Field(&c.Secret, validation.When(c.Enabled, validation.Required)),
	)
}

// -------------------------------------------------------------------

type BackupsConfig struct {
	// Cron is a cron expression to schedule auto backups, eg. "* * * * *".
	//
	// Leave it empty to disable the auto backups functionality.
	Cron string `form:"cron" json:"cron"`

	// CronMaxKeep is the max number of cron generated backups to
	// keep before removing older entries.
	//
	// This field works only when the cron config has valid cron expression.
	CronMaxKeep int `form:"cronMaxKeep" json:"cronMaxKeep"`

	// S3 is an optional S3 storage config specifying where to store the app backups.
	S3 S3Config `form:"s3" json:"s3"`
}

// Validate makes BackupsConfig validatable by implementing [validation.Validatable] interface.
func (c BackupsConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.S3),
		validation.Field(&c.Cron, validation.By(checkCronExpression)),
		validation.Field(
			&c.CronMaxKeep,
			validation.When(c.Cron != "", validation.Required),
			validation.Min(1),
		),
	)
}

func checkCronExpression(value any) error {
	v, _ := value.(string)
	if v == "" {
		return nil // nothing to check
	}

	_, err := cron.NewSchedule(v)
	if err != nil {
		return validation.NewError("validation_invalid_cron", err.Error())
	}

	return nil
}

// -------------------------------------------------------------------

type MetaConfig struct {
	AppName                    string        `form:"appName" json:"appName"`
	AppUrl                     string        `form:"appUrl" json:"appUrl"`
	HideControls               bool          `form:"hideControls" json:"hideControls"`
	SenderName                 string        `form:"senderName" json:"senderName"`
	SenderAddress              string        `form:"senderAddress" json:"senderAddress"`
	VerificationTemplate       EmailTemplate `form:"verificationTemplate" json:"verificationTemplate"`
	ResetPasswordTemplate      EmailTemplate `form:"resetPasswordTemplate" json:"resetPasswordTemplate"`
	ConfirmEmailChangeTemplate EmailTemplate `form:"confirmEmailChangeTemplate" json:"confirmEmailChangeTemplate"`
}

// Validate makes MetaConfig validatable by implementing [validation.Validatable] interface.
func (c MetaConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.AppName, validation.Required, validation.Length(1, 255)),
		validation.Field(&c.AppUrl, validation.Required, is.URL),
		validation.Field(&c.SenderName, validation.Required, validation.Length(1, 255)),
		validation.Field(&c.SenderAddress, is.EmailFormat, validation.Required),
		validation.Field(&c.VerificationTemplate, validation.Required),
		validation.Field(&c.ResetPasswordTemplate, validation.Required),
		validation.Field(&c.ConfirmEmailChangeTemplate, validation.Required),
	)
}

type EmailTemplate struct {
	Body      string `form:"body" json:"body"`
	Subject   string `form:"subject" json:"subject"`
	ActionUrl string `form:"actionUrl" json:"actionUrl"`
	Hidden    bool   `form:"hidden" json:"hidden"`
}

// Validate makes EmailTemplate validatable by implementing [validation.Validatable] interface.
func (t EmailTemplate) Validate() error {
	return validation.ValidateStruct(&t,
		validation.Field(&t.Subject, validation.Required),
		validation.Field(
			&t.Body,
			validation.Required,
			validation.By(checkPlaceholderParams(EmailPlaceholderActionUrl)),
		),
		validation.Field(
			&t.ActionUrl,
			validation.Required,
			validation.By(checkPlaceholderParams(EmailPlaceholderToken)),
		),
	)
}

func checkPlaceholderParams(params ...string) validation.RuleFunc {
	return func(value any) error {
		v, _ := value.(string)

		for _, param := range params {
			if !strings.Contains(v, param) {
				return validation.NewError(
					"validation_missing_required_param",
					fmt.Sprintf("Missing required parameter %q", param),
				)
			}
		}

		return nil
	}
}

// Resolve replaces the placeholder parameters in the current email
// template and returns its components as ready-to-use strings.
func (t EmailTemplate) Resolve(
	appName string,
	appUrl,
	token string,
) (subject, body, actionUrl string) {
	// replace action url placeholder params (if any)
	actionUrlParams := map[string]string{
		EmailPlaceholderAppName: appName,
		EmailPlaceholderAppUrl:  appUrl,
		EmailPlaceholderToken:   token,
	}
	actionUrl = t.ActionUrl
	for k, v := range actionUrlParams {
		actionUrl = strings.ReplaceAll(actionUrl, k, v)
	}
	actionUrl, _ = rest.NormalizeUrl(actionUrl)

	// replace body placeholder params (if any)
	bodyParams := map[string]string{
		EmailPlaceholderAppName:   appName,
		EmailPlaceholderAppUrl:    appUrl,
		EmailPlaceholderToken:     token,
		EmailPlaceholderActionUrl: actionUrl,
	}
	body = t.Body
	for k, v := range bodyParams {
		body = strings.ReplaceAll(body, k, v)
	}

	// replace subject placeholder params (if any)
	subjectParams := map[string]string{
		EmailPlaceholderAppName: appName,
		EmailPlaceholderAppUrl:  appUrl,
	}
	subject = t.Subject
	for k, v := range subjectParams {
		subject = strings.ReplaceAll(subject, k, v)
	}

	return subject, body, actionUrl
}

// -------------------------------------------------------------------

type LogsConfig struct {
	MaxDays  int  `form:"maxDays" json:"maxDays"`
	MinLevel int  `form:"minLevel" json:"minLevel"`
	LogIp    bool `form:"logIp" json:"logIp"`
}

// Validate makes LogsConfig validatable by implementing [validation.Validatable] interface.
func (c LogsConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.MaxDays, validation.Min(0)),
	)
}

// -------------------------------------------------------------------

type AuthProviderConfig struct {
	Enabled      bool   `form:"enabled" json:"enabled"`
	ClientId     string `form:"clientId" json:"clientId"`
	ClientSecret string `form:"clientSecret" json:"clientSecret"`
	AuthUrl      string `form:"authUrl" json:"authUrl"`
	TokenUrl     string `form:"tokenUrl" json:"tokenUrl"`
	UserApiUrl   string `form:"userApiUrl" json:"userApiUrl"`
	DisplayName  string `form:"displayName" json:"displayName"`
	PKCE         *bool  `form:"pkce" json:"pkce"`
}

// Validate makes `ProviderConfig` validatable by implementing [validation.Validatable] interface.
func (c AuthProviderConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.ClientId, validation.When(c.Enabled, validation.Required)),
		validation.Field(&c.ClientSecret, validation.When(c.Enabled, validation.Required)),
		validation.Field(&c.AuthUrl, is.URL),
		validation.Field(&c.TokenUrl, is.URL),
		validation.Field(&c.UserApiUrl, is.URL),
	)
}

// SetupProvider loads the current AuthProviderConfig into the specified provider.
func (c AuthProviderConfig) SetupProvider(provider auth.Provider) error {
	if !c.Enabled {
		return errors.New("the provider is not enabled")
	}

	if c.ClientId != "" {
		provider.SetClientId(c.ClientId)
	}

	if c.ClientSecret != "" {
		provider.SetClientSecret(c.ClientSecret)
	}

	if c.AuthUrl != "" {
		provider.SetAuthUrl(c.AuthUrl)
	}

	if c.UserApiUrl != "" {
		provider.SetUserApiUrl(c.UserApiUrl)
	}

	if c.TokenUrl != "" {
		provider.SetTokenUrl(c.TokenUrl)
	}

	if c.DisplayName != "" {
		provider.SetDisplayName(c.DisplayName)
	}

	if c.PKCE != nil {
		provider.SetPKCE(*c.PKCE)
	}

	return nil
}

// -------------------------------------------------------------------

type EmailAuthConfig struct {
	Enabled           bool     `form:"enabled" json:"enabled"`
	ExceptDomains     []string `form:"exceptDomains" json:"exceptDomains"`
	OnlyDomains       []string `form:"onlyDomains" json:"onlyDomains"`
	MinPasswordLength int      `form:"minPasswordLength" json:"minPasswordLength"`
}

func (c EmailAuthConfig) Validate() error {
	return nil
}
