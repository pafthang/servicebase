package mails

import (
	"github.com/pafthang/servicebase/core"
	baseconfig "github.com/pafthang/servicebase/services/base/config"
	settingsmodels "github.com/pafthang/servicebase/services/settings/models"
)

const (
	mailEnvClientID     = "GCP_MAIL_CLIENT_ID"
	mailEnvClientSecret = "GCP_MAIL_CLIENT_SECRET"
	mailEnvRedirectURL  = "GCP_MAIL_REDIRECT_URL"
)

type MailServiceConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func DefaultConfig() MailServiceConfig {
	return MailServiceConfig{}
}

func ConfigFromEnv() MailServiceConfig {
	return MailServiceConfig{
		ClientID:     baseconfig.String("", mailEnvClientID, ""),
		ClientSecret: baseconfig.String("", mailEnvClientSecret, ""),
		RedirectURL:  baseconfig.String("", mailEnvRedirectURL, ""),
	}
}

func ConfigFromSettings(app core.App) MailServiceConfig {
	if app == nil || app.Settings() == nil {
		return DefaultConfig()
	}

	return ConfigFromSettingsModel(app.Settings().Mail)
}

func ConfigFromSettingsModel(settings settingsmodels.MailConfig) MailServiceConfig {
	return MailServiceConfig{
		ClientID:     settings.ClientID,
		ClientSecret: settings.ClientSecret,
		RedirectURL:  settings.RedirectURL,
	}
}

func ResolveConfig(app core.App) MailServiceConfig {
	config := ConfigFromSettings(app)
	envConfig := ConfigFromEnv()

	config.ClientID = baseconfig.String(config.ClientID, mailEnvClientID, envConfig.ClientID)
	config.ClientSecret = baseconfig.String(config.ClientSecret, mailEnvClientSecret, envConfig.ClientSecret)
	config.RedirectURL = baseconfig.String(config.RedirectURL, mailEnvRedirectURL, envConfig.RedirectURL)

	return config
}
