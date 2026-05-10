package forms

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pafthang/servicebase/core"
	"github.com/pafthang/servicebase/daos"
	baseforms "github.com/pafthang/servicebase/services/base/forms"
	collectionmodels "github.com/pafthang/servicebase/services/collection/models"
	recordmodels "github.com/pafthang/servicebase/services/record/models"
	"github.com/pafthang/servicebase/tools/security"
)

type RecordEmailChangeConfirm struct {
	app        core.App
	dao        *daos.Dao
	collection *collectionmodels.Collection

	Token    string `form:"token" json:"token"`
	Password string `form:"password" json:"password"`
}

func NewRecordEmailChangeConfirm(app core.App, collection *collectionmodels.Collection) *RecordEmailChangeConfirm {
	return &RecordEmailChangeConfirm{
		app:        app,
		dao:        app.Dao(),
		collection: collection,
	}
}

func (form *RecordEmailChangeConfirm) SetDao(dao *daos.Dao) {
	form.dao = dao
}

func (form *RecordEmailChangeConfirm) Validate() error {
	return validation.ValidateStruct(form,
		validation.Field(&form.Token, validation.Required, validation.By(form.checkToken)),
		validation.Field(&form.Password, validation.Required, validation.Length(1, 100), validation.By(form.checkPassword)),
	)
}

func (form *RecordEmailChangeConfirm) checkToken(value any) error {
	v, _ := value.(string)
	if v == "" {
		return nil
	}

	authRecord, _, err := form.parseToken(v)
	if err != nil {
		return err
	}

	if authRecord.Collection().Id != form.collection.Id {
		return validation.NewError("validation_token_collection_mismatch", "The provided token is for a different users collection.")
	}

	return nil
}

func (form *RecordEmailChangeConfirm) checkPassword(value any) error {
	v, _ := value.(string)
	if v == "" {
		return nil
	}

	authRecord, _, _ := form.parseToken(form.Token)
	if authRecord == nil || !authRecord.ValidatePassword(v) {
		return validation.NewError("validation_invalid_password", "Missing or invalid user record password.")
	}

	return nil
}

func (form *RecordEmailChangeConfirm) parseToken(token string) (*recordmodels.Record, string, error) {
	claims, _ := security.ParseUnverifiedJWT(token)
	newEmail, _ := claims["newEmail"].(string)
	if newEmail == "" {
		return nil, "", validation.NewError("validation_invalid_token_payload", "Invalid token payload - newEmail must be set.")
	}

	if !form.dao.IsRecordValueUnique(form.collection.Id, collectionmodels.FieldNameEmail, newEmail) {
		return nil, "", validation.NewError("validation_existing_token_email", "The new email address is already registered: "+newEmail)
	}

	authRecord, err := form.dao.FindUserRecordByToken(token, form.app.Settings().RecordEmailChangeToken.Secret)
	if err != nil || authRecord == nil {
		return nil, "", validation.NewError("validation_invalid_token", "Invalid or expired token.")
	}

	return authRecord, newEmail, nil
}

func (form *RecordEmailChangeConfirm) Submit(interceptors ...baseforms.InterceptorFunc[*recordmodels.Record]) (*recordmodels.Record, error) {
	if err := form.Validate(); err != nil {
		return nil, err
	}

	authRecord, newEmail, err := form.parseToken(form.Token)
	if err != nil {
		return nil, err
	}

	authRecord.SetEmail(newEmail)
	authRecord.SetVerified(true)
	authRecord.RefreshTokenKey()

	interceptorsErr := baseforms.RunInterceptors(authRecord, func(m *recordmodels.Record) error {
		authRecord = m
		return form.dao.SaveRecord(m)
	}, interceptors...)
	if interceptorsErr != nil {
		return nil, interceptorsErr
	}

	return authRecord, nil
}
