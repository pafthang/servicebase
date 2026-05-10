package user

import (
	"log/slog"
	"sort"

	collectionmodels "github.com/pafthang/servicebase/services/collection/models"
	recordmodels "github.com/pafthang/servicebase/services/record/models"
	userforms "github.com/pafthang/servicebase/services/user/forms"
	usermodels "github.com/pafthang/servicebase/services/user/models"
	userqueries "github.com/pafthang/servicebase/services/user/queries"
	"github.com/pafthang/servicebase/tools/auth"
	"github.com/pafthang/servicebase/tools/security"
	"golang.org/x/oauth2"
)

func (s *Service) AuthMethods(collection *collectionmodels.Collection) ([]ProviderInfo, error) {
	userOptions := collection.UserOptions()
	if !userOptions.AllowOAuth2Auth {
		return []ProviderInfo{}, nil
	}

	result := []ProviderInfo{}
	nameConfigMap := s.App().Settings().NamedAuthProviderConfigs()
	for name, config := range nameConfigMap {
		if !config.Enabled {
			continue
		}

		provider, err := auth.NewProviderByName(name)
		if err != nil {
			s.App().Logger().Debug("Missing or invalid provider name", slog.String("name", name))
			continue
		}

		if err := config.SetupProvider(provider); err != nil {
			s.App().Logger().Debug(
				"Failed to setup provider",
				slog.String("name", name),
				slog.String("error", err.Error()),
			)
			continue
		}

		info := ProviderInfo{
			Name:        name,
			DisplayName: provider.DisplayName(),
			State:       security.RandomString(30),
		}

		if info.DisplayName == "" {
			info.DisplayName = name
		}

		urlOpts := []oauth2.AuthCodeOption{}
		if provider.PKCE() {
			info.CodeVerifier = security.RandomString(43)
			info.CodeChallenge = security.S256Challenge(info.CodeVerifier)
			info.CodeChallengeMethod = "S256"
			urlOpts = append(urlOpts,
				oauth2.SetAuthURLParam("code_challenge", info.CodeChallenge),
				oauth2.SetAuthURLParam("code_challenge_method", info.CodeChallengeMethod),
			)
		}

		info.AuthURL = provider.BuildAuthUrl(info.State, urlOpts...) + "&redirect_uri="
		result = append(result, info)
	}

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}

func (s *Service) NewOAuth2LoginForm(
	collection *collectionmodels.Collection,
	loggedAuthRecord *recordmodels.Record,
) *userforms.OAuth2Login {
	return userforms.NewOAuth2Login(s.App(), collection, loggedAuthRecord)
}

func (s *Service) SubmitOAuth2Login(
	form *userforms.OAuth2Login,
	interceptors ...userforms.OAuth2LoginInterceptorFunc,
) (*recordmodels.Record, *auth.AuthUser, error) {
	return form.Submit(interceptors...)
}

func (s *Service) NewPasswordLoginForm(collection *collectionmodels.Collection) *userforms.PasswordLogin {
	return userforms.NewPasswordLogin(s.App(), collection)
}

func (s *Service) SubmitPasswordLogin(
	form *userforms.PasswordLogin,
	interceptors ...userforms.RecordInterceptorFunc,
) (*recordmodels.Record, error) {
	return form.Submit(interceptors...)
}

func (s *Service) NewPasswordResetRequestForm(collection *collectionmodels.Collection) *userforms.PasswordResetRequest {
	return userforms.NewPasswordResetRequest(s.App(), collection)
}

func (s *Service) SubmitPasswordResetRequest(
	form *userforms.PasswordResetRequest,
	interceptors ...userforms.RecordInterceptorFunc,
) error {
	return form.Submit(interceptors...)
}

func (s *Service) NewPasswordResetConfirmForm(collection *collectionmodels.Collection) *userforms.PasswordResetConfirm {
	return userforms.NewPasswordResetConfirm(s.App(), collection)
}

func (s *Service) SubmitPasswordResetConfirm(
	form *userforms.PasswordResetConfirm,
	interceptors ...userforms.RecordInterceptorFunc,
) (*recordmodels.Record, error) {
	return form.Submit(interceptors...)
}

func (s *Service) NewVerificationRequestForm(collection *collectionmodels.Collection) *userforms.VerificationRequest {
	return userforms.NewVerificationRequest(s.App(), collection)
}

func (s *Service) SubmitVerificationRequest(
	form *userforms.VerificationRequest,
	interceptors ...userforms.RecordInterceptorFunc,
) error {
	return form.Submit(interceptors...)
}

func (s *Service) NewVerificationConfirmForm(collection *collectionmodels.Collection) *userforms.VerificationConfirm {
	return userforms.NewVerificationConfirm(s.App(), collection)
}

func (s *Service) SubmitVerificationConfirm(
	form *userforms.VerificationConfirm,
	interceptors ...userforms.RecordInterceptorFunc,
) (*recordmodels.Record, error) {
	return form.Submit(interceptors...)
}

func (s *Service) NewEmailChangeRequestForm(authRecord *recordmodels.Record) *userforms.EmailChangeRequest {
	return userforms.NewEmailChangeRequest(s.App(), authRecord)
}

func (s *Service) SubmitEmailChangeRequest(
	form *userforms.EmailChangeRequest,
	interceptors ...userforms.RecordInterceptorFunc,
) error {
	return form.Submit(interceptors...)
}

func (s *Service) NewEmailChangeConfirmForm(collection *collectionmodels.Collection) *userforms.EmailChangeConfirm {
	return userforms.NewEmailChangeConfirm(s.App(), collection)
}

func (s *Service) SubmitEmailChangeConfirm(
	form *userforms.EmailChangeConfirm,
	interceptors ...userforms.RecordInterceptorFunc,
) (*recordmodels.Record, error) {
	return form.Submit(interceptors...)
}

func (s *Service) FindRecordByID(collectionID, id string) (*recordmodels.Record, error) {
	return userqueries.FindRecordByID(s.Dao(), collectionID, id)
}

func (s *Service) FindFirstExternalAuthByRecord(
	collectionID string,
	recordID string,
	provider string,
) (*usermodels.ExternalAuth, error) {
	return userqueries.FindFirstExternalAuthByRecord(s.Dao(), collectionID, recordID, provider)
}

func (s *Service) FindAllExternalAuthsByRecord(collectionID, recordID string) ([]*usermodels.ExternalAuth, error) {
	return userqueries.FindAllExternalAuthsByRecord(s.Dao(), collectionID, recordID)
}

func (s *Service) DeleteExternalAuth(rel *usermodels.ExternalAuth) error {
	return userqueries.DeleteExternalAuth(s.Dao(), rel)
}
