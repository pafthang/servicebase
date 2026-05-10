package user

import (
	"github.com/pafthang/servicebase/core"
	servicebase "github.com/pafthang/servicebase/services/base"
)

type ProviderInfo struct {
	Name                string `json:"name"`
	DisplayName         string `json:"displayName"`
	State               string `json:"state"`
	AuthURL             string `json:"authUrl"`
	CodeVerifier        string `json:"codeVerifier"`
	CodeChallenge       string `json:"codeChallenge"`
	CodeChallengeMethod string `json:"codeChallengeMethod"`
}

var Descriptor = servicebase.Descriptor{
	Name:    "user",
	Purpose: "Provides user authentication workflows, auth-provider discovery, and external auth relationship helpers.",
	Dependencies: []string{
		"core.App",
		"services/base/forms",
		"tools/auth",
	},
	Operations: []string{
		"AuthMethods",
		"NewOAuth2LoginForm",
		"SubmitOAuth2Login",
		"NewPasswordLoginForm",
		"SubmitPasswordLogin",
		"NewPasswordResetRequestForm",
		"SubmitPasswordResetRequest",
		"NewPasswordResetConfirmForm",
		"SubmitPasswordResetConfirm",
		"NewVerificationRequestForm",
		"SubmitVerificationRequest",
		"NewVerificationConfirmForm",
		"SubmitVerificationConfirm",
		"NewEmailChangeRequestForm",
		"SubmitEmailChangeRequest",
		"NewEmailChangeConfirmForm",
		"SubmitEmailChangeConfirm",
		"FindRecordByID",
		"FindFirstExternalAuthByRecord",
		"FindAllExternalAuthsByRecord",
		"DeleteExternalAuth",
	},
}

type Service struct {
	servicebase.Service
}

func New(app core.App) *Service {
	return &Service{
		Service: servicebase.NewWithApp(app),
	}
}
