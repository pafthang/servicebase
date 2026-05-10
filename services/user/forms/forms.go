package forms

import (
	"github.com/pafthang/servicebase/core"
	baseforms "github.com/pafthang/servicebase/services/base/forms"
	collectionmodels "github.com/pafthang/servicebase/services/collection/models"
	recordmodels "github.com/pafthang/servicebase/services/record/models"
)

type OAuth2Login = RecordOAuth2Login
type OAuth2LoginData = RecordOAuth2LoginData
type PasswordLogin = RecordPasswordLogin
type PasswordResetRequest = RecordPasswordResetRequest
type PasswordResetConfirm = RecordPasswordResetConfirm
type VerificationRequest = RecordVerificationRequest
type VerificationConfirm = RecordVerificationConfirm
type EmailChangeRequest = RecordEmailChangeRequest
type EmailChangeConfirm = RecordEmailChangeConfirm

type RecordInterceptorFunc = baseforms.InterceptorFunc[*recordmodels.Record]
type OAuth2LoginInterceptorFunc = baseforms.InterceptorFunc[*OAuth2LoginData]

func NewOAuth2Login(
	app core.App,
	collection *collectionmodels.Collection,
	loggedAuthRecord *recordmodels.Record,
) *OAuth2Login {
	return NewRecordOAuth2Login(app, collection, loggedAuthRecord)
}

func NewPasswordLogin(app core.App, collection *collectionmodels.Collection) *PasswordLogin {
	return NewRecordPasswordLogin(app, collection)
}

func NewPasswordResetRequest(app core.App, collection *collectionmodels.Collection) *PasswordResetRequest {
	return NewRecordPasswordResetRequest(app, collection)
}

func NewPasswordResetConfirm(app core.App, collection *collectionmodels.Collection) *PasswordResetConfirm {
	return NewRecordPasswordResetConfirm(app, collection)
}

func NewVerificationRequest(app core.App, collection *collectionmodels.Collection) *VerificationRequest {
	return NewRecordVerificationRequest(app, collection)
}

func NewVerificationConfirm(app core.App, collection *collectionmodels.Collection) *VerificationConfirm {
	return NewRecordVerificationConfirm(app, collection)
}

func NewEmailChangeRequest(app core.App, authRecord *recordmodels.Record) *EmailChangeRequest {
	return NewRecordEmailChangeRequest(app, authRecord)
}

func NewEmailChangeConfirm(app core.App, collection *collectionmodels.Collection) *EmailChangeConfirm {
	return NewRecordEmailChangeConfirm(app, collection)
}
