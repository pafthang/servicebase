package models_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	settingsmodels "github.com/pafthang/servicebase/services/settings/models"
	"github.com/pafthang/servicebase/tools/auth"
	"github.com/pafthang/servicebase/tools/mailer"
	"github.com/pafthang/servicebase/tools/types"
)

func TestSettingsValidate(t *testing.T) {
	s := settingsmodels.NewSettings()

	// set invalid settings data
	s.Meta.AppName = ""
	s.Logs.MaxDays = -10
	s.Smtp.Enabled = true
	s.Smtp.Host = ""
	s.S3.Enabled = true
	s.S3.Endpoint = "invalid"
	s.RecordAuthToken.Duration = -10
	s.RecordPasswordResetToken.Duration = -10
	s.RecordEmailChangeToken.Duration = -10
	s.RecordVerificationToken.Duration = -10
	s.RecordFileToken.Duration = -10
	s.GithubAuth.Enabled = true
	s.GithubAuth.ClientId = ""
	s.VKAuth.Enabled = true
	s.VKAuth.ClientId = ""
	s.YandexAuth.Enabled = true
	s.YandexAuth.ClientId = ""

	// check if Validate() is triggering the members validate methods.
	err := s.Validate()
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	expectations := []string{
		`"meta":{`,
		`"logs":{`,
		`"smtp":{`,
		`"s3":{`,
		`"recordAuthToken":{`,
		`"recordPasswordResetToken":{`,
		`"recordEmailChangeToken":{`,
		`"recordVerificationToken":{`,
		`"recordFileToken":{`,
		`"githubAuth":{`,
		`"vkAuth":{`,
		`"yandexAuth":{`,
	}

	errBytes, _ := json.Marshal(err)
	jsonErr := string(errBytes)
	for _, expected := range expectations {
		if !strings.Contains(jsonErr, expected) {
			t.Errorf("Expected error key %s in %v", expected, jsonErr)
		}
	}
}

func TestSettingsMerge(t *testing.T) {
	s1 := settingsmodels.NewSettings()
	s1.Meta.AppUrl = "old_app_url"

	s2 := settingsmodels.NewSettings()
	s2.Meta.AppName = "test"
	s2.Logs.MaxDays = 123
	s2.Smtp.Host = "test"
	s2.Smtp.Enabled = true
	s2.S3.Enabled = true
	s2.S3.Endpoint = "test"
	s2.Backups.Cron = "* * * * *"
	s2.RecordAuthToken.Duration = 3
	s2.RecordPasswordResetToken.Duration = 4
	s2.RecordEmailChangeToken.Duration = 5
	s2.RecordVerificationToken.Duration = 6
	s2.RecordFileToken.Duration = 7
	s2.GithubAuth.Enabled = true
	s2.GithubAuth.ClientId = "github_test"
	s2.VKAuth.Enabled = true
	s2.VKAuth.ClientId = "vk_test"
	s2.YandexAuth.Enabled = true
	s2.YandexAuth.ClientId = "yandex_test"

	if err := s1.Merge(s2); err != nil {
		t.Fatal(err)
	}

	s1Encoded, err := json.Marshal(s1)
	if err != nil {
		t.Fatal(err)
	}

	s2Encoded, err := json.Marshal(s2)
	if err != nil {
		t.Fatal(err)
	}

	if string(s1Encoded) != string(s2Encoded) {
		t.Fatalf("Expected the same serialization, got %v VS %v", string(s1Encoded), string(s2Encoded))
	}
}

func TestSettingsClone(t *testing.T) {
	s1 := settingsmodels.NewSettings()

	s2, err := s1.Clone()
	if err != nil {
		t.Fatal(err)
	}

	s1Bytes, err := json.Marshal(s1)
	if err != nil {
		t.Fatal(err)
	}

	s2Bytes, err := json.Marshal(s2)
	if err != nil {
		t.Fatal(err)
	}

	if string(s1Bytes) != string(s2Bytes) {
		t.Fatalf("Expected equivalent serialization, got %v VS %v", string(s1Bytes), string(s2Bytes))
	}

	// verify that it is a deep copy
	s1.Meta.AppName = "new"
	if s1.Meta.AppName == s2.Meta.AppName {
		t.Fatalf("Expected s1 and s2 to have different Meta.AppName, got %s", s1.Meta.AppName)
	}
}

func TestSettingsRedactClone(t *testing.T) {
	testSecret := "test_secret"

	s1 := settingsmodels.NewSettings()

	// control fields
	s1.Meta.AppName = "test123"

	// secrets
	s1.Smtp.Password = testSecret
	s1.S3.Secret = testSecret
	s1.Backups.S3.Secret = testSecret
	s1.RecordAuthToken.Secret = testSecret
	s1.RecordPasswordResetToken.Secret = testSecret
	s1.RecordEmailChangeToken.Secret = testSecret
	s1.RecordVerificationToken.Secret = testSecret
	s1.RecordFileToken.Secret = testSecret
	s1.Calendar.ClientSecret = testSecret
	s1.Mail.ClientSecret = testSecret
	s1.Weather.APIKey = testSecret
	s1.GithubAuth.ClientSecret = testSecret
	s1.VKAuth.ClientSecret = testSecret
	s1.YandexAuth.ClientSecret = testSecret

	s1Bytes, err := json.Marshal(s1)
	if err != nil {
		t.Fatal(err)
	}

	s2, err := s1.RedactClone()
	if err != nil {
		t.Fatal(err)
	}

	s2Bytes, err := json.Marshal(s2)
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(s1Bytes, s2Bytes) {
		t.Fatalf("Expected the 2 settings to differ, got \n%s", s2Bytes)
	}

	if strings.Contains(string(s2Bytes), testSecret) {
		t.Fatalf("Expected %q secret to be replaced with mask, got \n%s", testSecret, s2Bytes)
	}

	if !strings.Contains(string(s2Bytes), settingsmodels.SecretMask) {
		t.Fatalf("Expected the secrets to be replaced with the secret mask, got \n%s", s2Bytes)
	}

	if !strings.Contains(string(s2Bytes), `"appName":"test123"`) {
		t.Fatalf("Missing control field in \n%s", s2Bytes)
	}
}

func TestNamedAuthProviderConfigs(t *testing.T) {
	s := settingsmodels.NewSettings()

	s.GithubAuth.ClientId = "github_test"
	s.VKAuth.ClientId = "vk_test"
	s.YandexAuth.ClientId = "yandex_test"

	result := s.NamedAuthProviderConfigs()

	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	encodedStr := string(encoded)

	expectedParts := []string{
		`"github":{"enabled":false,"clientId":"github_test"`,
		`"vk":{"enabled":false,"clientId":"vk_test"`,
		`"yandex":{"enabled":false,"clientId":"yandex_test"`,
	}
	for _, p := range expectedParts {
		if !strings.Contains(encodedStr, p) {
			t.Fatalf("Expected \n%s \nin \n%s", p, encodedStr)
		}
	}
}

func TestTokenConfigValidate(t *testing.T) {
	scenarios := []struct {
		config      settingsmodels.TokenConfig
		expectError bool
	}{
		// zero values
		{
			settingsmodels.TokenConfig{},
			true,
		},
		// invalid data
		{
			settingsmodels.TokenConfig{
				Secret:   strings.Repeat("a", 5),
				Duration: 4,
			},
			true,
		},
		// valid secret but invalid duration
		{
			settingsmodels.TokenConfig{
				Secret:   strings.Repeat("a", 30),
				Duration: 63072000 + 1,
			},
			true,
		},
		// valid data
		{
			settingsmodels.TokenConfig{
				Secret:   strings.Repeat("a", 30),
				Duration: 100,
			},
			false,
		},
	}

	for i, scenario := range scenarios {
		result := scenario.config.Validate()

		if result != nil && !scenario.expectError {
			t.Errorf("(%d) Didn't expect error, got %v", i, result)
		}

		if result == nil && scenario.expectError {
			t.Errorf("(%d) Expected error, got nil", i)
		}
	}
}

func TestSmtpConfigValidate(t *testing.T) {
	scenarios := []struct {
		config      settingsmodels.SmtpConfig
		expectError bool
	}{
		// zero values (disabled)
		{
			settingsmodels.SmtpConfig{},
			false,
		},
		// zero values (enabled)
		{
			settingsmodels.SmtpConfig{Enabled: true},
			true,
		},
		// invalid data
		{
			settingsmodels.SmtpConfig{
				Enabled: true,
				Host:    "test:test:test",
				Port:    -10,
			},
			true,
		},
		// invalid auth method
		{
			settingsmodels.SmtpConfig{
				Enabled:    true,
				Host:       "example.com",
				Port:       100,
				AuthMethod: "example",
			},
			true,
		},
		// valid data (no explicit auth method)
		{
			settingsmodels.SmtpConfig{
				Enabled: true,
				Host:    "example.com",
				Port:    100,
				Tls:     true,
			},
			false,
		},
		// valid data (explicit auth method - login)
		{
			settingsmodels.SmtpConfig{
				Enabled:    true,
				Host:       "example.com",
				Port:       100,
				AuthMethod: mailer.SmtpAuthLogin,
			},
			false,
		},
		// invalid ehlo/helo name
		{
			settingsmodels.SmtpConfig{
				Enabled:   true,
				Host:      "example.com",
				Port:      100,
				LocalName: "invalid!",
			},
			true,
		},
		// valid ehlo/helo name
		{
			settingsmodels.SmtpConfig{
				Enabled:   true,
				Host:      "example.com",
				Port:      100,
				LocalName: "example.com",
			},
			false,
		},
	}

	for i, scenario := range scenarios {
		result := scenario.config.Validate()

		if result != nil && !scenario.expectError {
			t.Errorf("(%d) Didn't expect error, got %v", i, result)
		}

		if result == nil && scenario.expectError {
			t.Errorf("(%d) Expected error, got nil", i)
		}
	}
}

func TestS3ConfigValidate(t *testing.T) {
	scenarios := []struct {
		config      settingsmodels.S3Config
		expectError bool
	}{
		// zero values (disabled)
		{
			settingsmodels.S3Config{},
			false,
		},
		// zero values (enabled)
		{
			settingsmodels.S3Config{Enabled: true},
			true,
		},
		// invalid data
		{
			settingsmodels.S3Config{
				Enabled:  true,
				Endpoint: "test:test:test",
			},
			true,
		},
		// valid data (url endpoint)
		{
			settingsmodels.S3Config{
				Enabled:   true,
				Endpoint:  "https://localhost:8090",
				Bucket:    "test",
				Region:    "test",
				AccessKey: "test",
				Secret:    "test",
			},
			false,
		},
		// valid data (hostname endpoint)
		{
			settingsmodels.S3Config{
				Enabled:   true,
				Endpoint:  "example.com",
				Bucket:    "test",
				Region:    "test",
				AccessKey: "test",
				Secret:    "test",
			},
			false,
		},
	}

	for i, scenario := range scenarios {
		result := scenario.config.Validate()

		if result != nil && !scenario.expectError {
			t.Errorf("(%d) Didn't expect error, got %v", i, result)
		}

		if result == nil && scenario.expectError {
			t.Errorf("(%d) Expected error, got nil", i)
		}
	}
}

func TestMetaConfigValidate(t *testing.T) {
	invalidTemplate := settingsmodels.EmailTemplate{
		Subject:   "test",
		ActionUrl: "test",
		Body:      "test",
	}

	noPlaceholdersTemplate := settingsmodels.EmailTemplate{
		Subject:   "test",
		ActionUrl: "http://example.com",
		Body:      "test",
	}

	withPlaceholdersTemplate := settingsmodels.EmailTemplate{
		Subject:   "test",
		ActionUrl: "http://example.com" + settingsmodels.EmailPlaceholderToken,
		Body:      "test" + settingsmodels.EmailPlaceholderActionUrl,
	}

	scenarios := []struct {
		config      settingsmodels.MetaConfig
		expectError bool
	}{
		// zero values
		{
			settingsmodels.MetaConfig{},
			true,
		},
		// invalid data
		{
			settingsmodels.MetaConfig{
				AppName:                    strings.Repeat("a", 300),
				AppUrl:                     "test",
				SenderName:                 strings.Repeat("a", 300),
				SenderAddress:              "invalid_email",
				VerificationTemplate:       invalidTemplate,
				ResetPasswordTemplate:      invalidTemplate,
				ConfirmEmailChangeTemplate: invalidTemplate,
			},
			true,
		},
		// invalid data (missing required placeholders)
		{
			settingsmodels.MetaConfig{
				AppName:                    "test",
				AppUrl:                     "https://example.com",
				SenderName:                 "test",
				SenderAddress:              "test@example.com",
				VerificationTemplate:       noPlaceholdersTemplate,
				ResetPasswordTemplate:      noPlaceholdersTemplate,
				ConfirmEmailChangeTemplate: noPlaceholdersTemplate,
			},
			true,
		},
		// valid data
		{
			settingsmodels.MetaConfig{
				AppName:                    "test",
				AppUrl:                     "https://example.com",
				SenderName:                 "test",
				SenderAddress:              "test@example.com",
				VerificationTemplate:       withPlaceholdersTemplate,
				ResetPasswordTemplate:      withPlaceholdersTemplate,
				ConfirmEmailChangeTemplate: withPlaceholdersTemplate,
			},
			false,
		},
	}

	for i, scenario := range scenarios {
		result := scenario.config.Validate()

		if result != nil && !scenario.expectError {
			t.Errorf("(%d) Didn't expect error, got %v", i, result)
		}

		if result == nil && scenario.expectError {
			t.Errorf("(%d) Expected error, got nil", i)
		}
	}
}

func TestBackupsConfigValidate(t *testing.T) {
	scenarios := []struct {
		name           string
		config         settingsmodels.BackupsConfig
		expectedErrors []string
	}{
		{
			"zero value",
			settingsmodels.BackupsConfig{},
			[]string{},
		},
		{
			"invalid cron",
			settingsmodels.BackupsConfig{
				Cron:        "invalid",
				CronMaxKeep: 0,
			},
			[]string{"cron", "cronMaxKeep"},
		},
		{
			"invalid enabled S3",
			settingsmodels.BackupsConfig{
				S3: settingsmodels.S3Config{
					Enabled: true,
				},
			},
			[]string{"s3"},
		},
		{
			"valid data",
			settingsmodels.BackupsConfig{
				S3: settingsmodels.S3Config{
					Enabled:   true,
					Endpoint:  "example.com",
					Bucket:    "test",
					Region:    "test",
					AccessKey: "test",
					Secret:    "test",
				},
				Cron:        "*/10 * * * *",
				CronMaxKeep: 1,
			},
			[]string{},
		},
	}

	for _, s := range scenarios {
		result := s.config.Validate()

		// parse errors
		errs, ok := result.(validation.Errors)
		if !ok && result != nil {
			t.Errorf("[%s] Failed to parse errors %v", s.name, result)
			continue
		}

		// check errors
		if len(errs) > len(s.expectedErrors) {
			t.Errorf("[%s] Expected error keys %v, got %v", s.name, s.expectedErrors, errs)
		}
		for _, k := range s.expectedErrors {
			if _, ok := errs[k]; !ok {
				t.Errorf("[%s] Missing expected error key %q in %v", s.name, k, errs)
			}
		}
	}
}

func TestEmailTemplateValidate(t *testing.T) {
	scenarios := []struct {
		emailTemplate  settingsmodels.EmailTemplate
		expectedErrors []string
	}{
		// require values
		{
			settingsmodels.EmailTemplate{},
			[]string{"subject", "actionUrl", "body"},
		},
		// missing placeholders
		{
			settingsmodels.EmailTemplate{
				Subject:   "test",
				ActionUrl: "test",
				Body:      "test",
			},
			[]string{"actionUrl", "body"},
		},
		// valid data
		{
			settingsmodels.EmailTemplate{
				Subject:   "test",
				ActionUrl: "test" + settingsmodels.EmailPlaceholderToken,
				Body:      "test" + settingsmodels.EmailPlaceholderActionUrl,
			},
			[]string{},
		},
	}

	for i, s := range scenarios {
		result := s.emailTemplate.Validate()

		// parse errors
		errs, ok := result.(validation.Errors)
		if !ok && result != nil {
			t.Errorf("(%d) Failed to parse errors %v", i, result)
			continue
		}

		// check errors
		if len(errs) > len(s.expectedErrors) {
			t.Errorf("(%d) Expected error keys %v, got %v", i, s.expectedErrors, errs)
		}
		for _, k := range s.expectedErrors {
			if _, ok := errs[k]; !ok {
				t.Errorf("(%d) Missing expected error key %q in %v", i, k, errs)
			}
		}
	}
}

func TestEmailTemplateResolve(t *testing.T) {
	allPlaceholders := settingsmodels.EmailPlaceholderActionUrl + settingsmodels.EmailPlaceholderToken + settingsmodels.EmailPlaceholderAppName + settingsmodels.EmailPlaceholderAppUrl

	scenarios := []struct {
		emailTemplate     settingsmodels.EmailTemplate
		expectedSubject   string
		expectedBody      string
		expectedActionUrl string
	}{
		// no placeholders
		{
			emailTemplate: settingsmodels.EmailTemplate{
				Subject:   "subject:",
				Body:      "body:",
				ActionUrl: "/actionUrl////",
			},
			expectedSubject:   "subject:",
			expectedActionUrl: "/actionUrl/",
			expectedBody:      "body:",
		},
		// with placeholders
		{
			emailTemplate: settingsmodels.EmailTemplate{
				ActionUrl: "/actionUrl////" + allPlaceholders,
				Subject:   "subject:" + allPlaceholders,
				Body:      "body:" + allPlaceholders,
			},
			expectedActionUrl: fmt.Sprintf(
				"/actionUrl/%%7BACTION_URL%%7D%s%s%s",
				"token_test",
				"name_test",
				"url_test",
			),
			expectedSubject: fmt.Sprintf(
				"subject:%s%s%s%s",
				settingsmodels.EmailPlaceholderActionUrl,
				settingsmodels.EmailPlaceholderToken,
				"name_test",
				"url_test",
			),
			expectedBody: fmt.Sprintf(
				"body:%s%s%s%s",
				fmt.Sprintf(
					"/actionUrl/%%7BACTION_URL%%7D%s%s%s",
					"token_test",
					"name_test",
					"url_test",
				),
				"token_test",
				"name_test",
				"url_test",
			),
		},
	}

	for i, s := range scenarios {
		subject, body, actionUrl := s.emailTemplate.Resolve("name_test", "url_test", "token_test")

		if s.expectedSubject != subject {
			t.Errorf("(%d) Expected subject %q got %q", i, s.expectedSubject, subject)
		}

		if s.expectedBody != body {
			t.Errorf("(%d) Expected body \n%v got \n%v", i, s.expectedBody, body)
		}

		if s.expectedActionUrl != actionUrl {
			t.Errorf("(%d) Expected actionUrl \n%v got \n%v", i, s.expectedActionUrl, actionUrl)
		}
	}
}

func TestLogsConfigValidate(t *testing.T) {
	scenarios := []struct {
		config      settingsmodels.LogsConfig
		expectError bool
	}{
		// zero values
		{
			settingsmodels.LogsConfig{},
			false,
		},
		// invalid data
		{
			settingsmodels.LogsConfig{MaxDays: -10},
			true,
		},
		// valid data
		{
			settingsmodels.LogsConfig{MaxDays: 1},
			false,
		},
	}

	for i, scenario := range scenarios {
		result := scenario.config.Validate()

		if result != nil && !scenario.expectError {
			t.Errorf("(%d) Didn't expect error, got %v", i, result)
		}

		if result == nil && scenario.expectError {
			t.Errorf("(%d) Expected error, got nil", i)
		}
	}
}

func TestAuthProviderConfigValidate(t *testing.T) {
	scenarios := []struct {
		config      settingsmodels.AuthProviderConfig
		expectError bool
	}{
		// zero values (disabled)
		{
			settingsmodels.AuthProviderConfig{},
			false,
		},
		// zero values (enabled)
		{
			settingsmodels.AuthProviderConfig{Enabled: true},
			true,
		},
		// invalid data
		{
			settingsmodels.AuthProviderConfig{
				Enabled:      true,
				ClientId:     "",
				ClientSecret: "",
				AuthUrl:      "test",
				TokenUrl:     "test",
				UserApiUrl:   "test",
			},
			true,
		},
		// valid data (only the required)
		{
			settingsmodels.AuthProviderConfig{
				Enabled:      true,
				ClientId:     "test",
				ClientSecret: "test",
			},
			false,
		},
		// valid data (fill all fields)
		{
			settingsmodels.AuthProviderConfig{
				Enabled:      true,
				ClientId:     "test",
				ClientSecret: "test",
				DisplayName:  "test",
				PKCE:         types.Pointer(true),
				AuthUrl:      "https://example.com",
				TokenUrl:     "https://example.com",
				UserApiUrl:   "https://example.com",
			},
			false,
		},
	}

	for i, scenario := range scenarios {
		result := scenario.config.Validate()

		if result != nil && !scenario.expectError {
			t.Errorf("(%d) Didn't expect error, got %v", i, result)
		}

		if result == nil && scenario.expectError {
			t.Errorf("(%d) Expected error, got nil", i)
		}
	}
}

func TestAuthProviderConfigSetupProvider(t *testing.T) {
	provider := auth.NewGithubProvider()

	// disabled config
	c1 := settingsmodels.AuthProviderConfig{Enabled: false}
	if err := c1.SetupProvider(provider); err == nil {
		t.Errorf("Expected error, got nil")
	}

	c2 := settingsmodels.AuthProviderConfig{
		Enabled:      true,
		ClientId:     "test_ClientId",
		ClientSecret: "test_ClientSecret",
		AuthUrl:      "test_AuthUrl",
		UserApiUrl:   "test_UserApiUrl",
		TokenUrl:     "test_TokenUrl",
		DisplayName:  "test_DisplayName",
		PKCE:         types.Pointer(true),
	}
	if err := c2.SetupProvider(provider); err != nil {
		t.Error(err)
	}

	if provider.ClientId() != c2.ClientId {
		t.Fatalf("Expected ClientId %s, got %s", c2.ClientId, provider.ClientId())
	}

	if provider.ClientSecret() != c2.ClientSecret {
		t.Fatalf("Expected ClientSecret %s, got %s", c2.ClientSecret, provider.ClientSecret())
	}

	if provider.AuthUrl() != c2.AuthUrl {
		t.Fatalf("Expected AuthUrl %s, got %s", c2.AuthUrl, provider.AuthUrl())
	}

	if provider.UserApiUrl() != c2.UserApiUrl {
		t.Fatalf("Expected UserApiUrl %s, got %s", c2.UserApiUrl, provider.UserApiUrl())
	}

	if provider.TokenUrl() != c2.TokenUrl {
		t.Fatalf("Expected TokenUrl %s, got %s", c2.TokenUrl, provider.TokenUrl())
	}

	if provider.DisplayName() != c2.DisplayName {
		t.Fatalf("Expected DisplayName %s, got %s", c2.DisplayName, provider.DisplayName())
	}

	if provider.PKCE() != *c2.PKCE {
		t.Fatalf("Expected PKCE %v, got %v", *c2.PKCE, provider.PKCE())
	}
}
