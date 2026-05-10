package forms

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pafthang/servicebase/core"
	"github.com/pafthang/servicebase/daos"
	baseforms "github.com/pafthang/servicebase/services/base/forms"
	collectionmodels "github.com/pafthang/servicebase/services/collection/models"
	recordforms "github.com/pafthang/servicebase/services/record/forms"
	recordmodels "github.com/pafthang/servicebase/services/record/models"
	usermodels "github.com/pafthang/servicebase/services/user/models"
	"github.com/pafthang/servicebase/tools/auth"
	"github.com/pafthang/servicebase/tools/security"
	"github.com/pocketbase/dbx"
	"golang.org/x/oauth2"
)

var usernameRegex = regexp.MustCompile(`^[\w][\w\.\-]*$`)

type RecordOAuth2LoginData struct {
	ExternalAuth   *usermodels.ExternalAuth
	Record         *recordmodels.Record
	OAuth2User     *auth.AuthUser
	ProviderClient auth.Provider
}

type BeforeOAuth2RecordCreateFunc func(createForm *recordforms.RecordUpsert, userRecord *recordmodels.Record, authUser *auth.AuthUser) error

type RecordOAuth2Login struct {
	app        core.App
	dao        *daos.Dao
	collection *collectionmodels.Collection

	beforeOAuth2RecordCreateFunc BeforeOAuth2RecordCreateFunc
	loggedUserRecord             *recordmodels.Record

	Provider     string         `form:"provider" json:"provider"`
	Code         string         `form:"code" json:"code"`
	CodeVerifier string         `form:"codeVerifier" json:"codeVerifier"`
	RedirectUrl  string         `form:"redirectUrl" json:"redirectUrl"`
	CreateData   map[string]any `form:"createData" json:"createData"`
}

func NewRecordOAuth2Login(app core.App, collection *collectionmodels.Collection, optAuthRecord *recordmodels.Record) *RecordOAuth2Login {
	return &RecordOAuth2Login{
		app:              app,
		dao:              app.Dao(),
		collection:       collection,
		loggedUserRecord: optAuthRecord,
	}
}

func (form *RecordOAuth2Login) SetDao(dao *daos.Dao) {
	form.dao = dao
}

func (form *RecordOAuth2Login) SetBeforeNewRecordCreateFunc(f BeforeOAuth2RecordCreateFunc) {
	form.beforeOAuth2RecordCreateFunc = f
}

func (form *RecordOAuth2Login) Validate() error {
	return validation.ValidateStruct(form,
		validation.Field(&form.Provider, validation.Required, validation.By(form.checkProviderName)),
		validation.Field(&form.Code, validation.Required),
		validation.Field(&form.RedirectUrl, validation.Required),
	)
}

func (form *RecordOAuth2Login) checkProviderName(value any) error {
	name, _ := value.(string)

	config, ok := form.app.Settings().NamedAuthProviderConfigs()[name]
	if !ok || !config.Enabled {
		return validation.NewError("validation_invalid_provider", fmt.Sprintf("%q is missing or is not enabled.", name))
	}

	return nil
}

func (form *RecordOAuth2Login) Submit(interceptors ...baseforms.InterceptorFunc[*RecordOAuth2LoginData]) (*recordmodels.Record, *auth.AuthUser, error) {
	if err := form.Validate(); err != nil {
		return nil, nil, err
	}

	if !form.collection.UserOptions().AllowOAuth2Auth {
		return nil, nil, errors.New("OAuth2 authentication is not allowed for this users collection.")
	}

	provider, err := auth.NewProviderByName(form.Provider)
	if err != nil {
		return nil, nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	provider.SetContext(ctx)

	providerConfig := form.app.Settings().NamedAuthProviderConfigs()[form.Provider]
	if err := providerConfig.SetupProvider(provider); err != nil {
		return nil, nil, err
	}

	provider.SetRedirectUrl(form.RedirectUrl)

	var opts []oauth2.AuthCodeOption
	if provider.PKCE() {
		opts = append(opts, oauth2.SetAuthURLParam("code_verifier", form.CodeVerifier))
	}

	token, err := provider.FetchToken(form.Code, opts...)
	if err != nil {
		return nil, nil, err
	}

	authUser, err := provider.FetchAuthUser(token)
	if err != nil {
		return nil, nil, err
	}

	var userRecord *recordmodels.Record

	rel, _ := form.dao.FindFirstExternalAuthByExpr(dbx.HashExp{
		"provider":   form.Provider,
		"providerId": authUser.Id,
	})

	switch {
	case rel != nil:
		userRecord, err = form.dao.FindRecordById(form.collection.Id, rel.UserID)
		if err != nil {
			return nil, authUser, err
		}
	case form.loggedUserRecord != nil && form.loggedUserRecord.Collection().Id == form.collection.Id:
		userRecord = form.loggedUserRecord
	case authUser.Email != "":
		userRecord, _ = form.dao.FindUserRecordByEmail(form.collection.Id, authUser.Email)
	}

	interceptorData := &RecordOAuth2LoginData{
		ExternalAuth:   rel,
		Record:         userRecord,
		OAuth2User:     authUser,
		ProviderClient: provider,
	}

	interceptorsErr := baseforms.RunInterceptors(interceptorData, func(newData *RecordOAuth2LoginData) error {
		return form.submit(newData)
	}, interceptors...)
	if interceptorsErr != nil {
		return nil, nil, interceptorsErr
	}

	return interceptorData.Record, interceptorData.OAuth2User, nil
}

func (form *RecordOAuth2Login) submit(data *RecordOAuth2LoginData) error {
	return form.dao.RunInTransaction(func(txDao *daos.Dao) error {
		if data.Record == nil {
			data.Record = recordmodels.NewRecord(form.collection)
			data.Record.RefreshId()
			data.Record.MarkAsNew()

			createForm := recordforms.NewRecordUpsert(form.app, data.Record)
			createForm.SetFullManageAccess(true)
			createForm.SetDao(txDao)

			if data.OAuth2User.Username != "" &&
				len(data.OAuth2User.Username) >= 3 &&
				len(data.OAuth2User.Username) <= 150 &&
				usernameRegex.MatchString(data.OAuth2User.Username) {
				createForm.Username = form.dao.SuggestUniqueUserRecordUsername(form.collection.Id, data.OAuth2User.Username)
			}

			createForm.LoadData(form.CreateData)
			createForm.Email = data.OAuth2User.Email
			createForm.Verified = true

			if createForm.Password == "" {
				createForm.Password = security.RandomString(30)
				createForm.PasswordConfirm = createForm.Password
			}

			if form.beforeOAuth2RecordCreateFunc != nil {
				if err := form.beforeOAuth2RecordCreateFunc(createForm, data.Record, data.OAuth2User); err != nil {
					return err
				}
			}

			if err := createForm.Submit(); err != nil {
				return err
			}
		} else {
			isLoggedUserRecord := form.loggedUserRecord != nil &&
				form.loggedUserRecord.Id == data.Record.Id &&
				form.loggedUserRecord.Collection().Id == data.Record.Collection().Id

			if !isLoggedUserRecord && data.Record.Email() != "" && !data.Record.Verified() {
				data.Record.SetPassword(security.RandomString(30))
				if err := txDao.SaveRecord(data.Record); err != nil {
					return err
				}
			}

			if !data.Record.Verified() {
				externalAuths, err := txDao.FindAllExternalAuthsByRecord(data.Record)
				if err != nil {
					return err
				}
				for _, ea := range externalAuths {
					if err := txDao.DeleteExternalAuth(ea); err != nil {
						return err
					}
				}
				data.ExternalAuth = nil
			}

			if data.Record.Email() == "" && data.OAuth2User.Email != "" {
				data.Record.SetEmail(data.OAuth2User.Email)
				if err := txDao.SaveRecord(data.Record); err != nil {
					return err
				}
			}

			if !data.Record.Verified() && (data.Record.Email() == "" || data.Record.Email() == data.OAuth2User.Email) {
				data.Record.SetVerified(true)
				if err := txDao.SaveRecord(data.Record); err != nil {
					return err
				}
			}
		}

		if data.ExternalAuth == nil {
			data.ExternalAuth = &usermodels.ExternalAuth{
				UserCollectionID: data.Record.Collection().Id,
				UserID:           data.Record.Id,
				Provider:         form.Provider,
				ProviderID:       data.OAuth2User.Id,
			}
			if err := txDao.SaveExternalAuth(data.ExternalAuth); err != nil {
				return err
			}
		}

		return nil
	})
}
