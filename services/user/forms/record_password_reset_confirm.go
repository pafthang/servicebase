package forms

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pafthang/servicebase/core"
	"github.com/pafthang/servicebase/daos"
	baseforms "github.com/pafthang/servicebase/services/base/forms"
	collectionmodels "github.com/pafthang/servicebase/services/collection/models"
	recordmodels "github.com/pafthang/servicebase/services/record/models"
	"github.com/pafthang/servicebase/tools/security"
	"github.com/pafthang/servicebase/tools/validators"
	"github.com/spf13/cast"
)

type RecordPasswordResetConfirm struct {
	app        core.App
	collection *collectionmodels.Collection
	dao        *daos.Dao

	Token           string `form:"token" json:"token"`
	Password        string `form:"password" json:"password"`
	PasswordConfirm string `form:"passwordConfirm" json:"passwordConfirm"`
}

func NewRecordPasswordResetConfirm(app core.App, collection *collectionmodels.Collection) *RecordPasswordResetConfirm {
	return &RecordPasswordResetConfirm{
		app:        app,
		dao:        app.Dao(),
		collection: collection,
	}
}

func (form *RecordPasswordResetConfirm) SetDao(dao *daos.Dao) {
	form.dao = dao
}

func (form *RecordPasswordResetConfirm) Validate() error {
	minPasswordLength := form.collection.UserOptions().MinPasswordLength

	return validation.ValidateStruct(form,
		validation.Field(&form.Token, validation.Required, validation.By(form.checkToken)),
		validation.Field(&form.Password, validation.Required, validation.Length(minPasswordLength, 100)),
		validation.Field(&form.PasswordConfirm, validation.Required, validation.By(validators.Compare(form.Password))),
	)
}

func (form *RecordPasswordResetConfirm) checkToken(value any) error {
	v, _ := value.(string)
	if v == "" {
		return nil
	}

	record, err := form.dao.FindUserRecordByToken(v, form.app.Settings().RecordPasswordResetToken.Secret)
	if err != nil || record == nil {
		return validation.NewError("validation_invalid_token", "Invalid or expired token.")
	}

	if record.Collection().Id != form.collection.Id {
		return validation.NewError("validation_token_collection_mismatch", "The provided token is for a different users collection.")
	}

	return nil
}

func (form *RecordPasswordResetConfirm) Submit(interceptors ...baseforms.InterceptorFunc[*recordmodels.Record]) (*recordmodels.Record, error) {
	if err := form.Validate(); err != nil {
		return nil, err
	}

	authRecord, err := form.dao.FindUserRecordByToken(form.Token, form.app.Settings().RecordPasswordResetToken.Secret)
	if err != nil {
		return nil, err
	}

	if err := authRecord.SetPassword(form.Password); err != nil {
		return nil, err
	}

	if !authRecord.Verified() {
		payload, err := security.ParseUnverifiedJWT(form.Token)
		if err != nil {
			return nil, err
		}
		if authRecord.Email() == cast.ToString(payload["email"]) {
			authRecord.SetVerified(true)
		}
	}

	interceptorsErr := baseforms.RunInterceptors(authRecord, func(m *recordmodels.Record) error {
		authRecord = m
		return form.dao.SaveRecord(authRecord)
	}, interceptors...)
	if interceptorsErr != nil {
		return nil, interceptorsErr
	}

	return authRecord, nil
}
