package forms

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pafthang/servicebase/core"
	"github.com/pafthang/servicebase/daos"
	baseforms "github.com/pafthang/servicebase/services/base/forms"
	collectionmodels "github.com/pafthang/servicebase/services/collection/models"
	recordmodels "github.com/pafthang/servicebase/services/record/models"
	"github.com/pafthang/servicebase/tools/security"
	"github.com/spf13/cast"
)

type RecordVerificationConfirm struct {
	app        core.App
	collection *collectionmodels.Collection
	dao        *daos.Dao

	Token string `form:"token" json:"token"`
}

func NewRecordVerificationConfirm(app core.App, collection *collectionmodels.Collection) *RecordVerificationConfirm {
	return &RecordVerificationConfirm{
		app:        app,
		dao:        app.Dao(),
		collection: collection,
	}
}

func (form *RecordVerificationConfirm) SetDao(dao *daos.Dao) {
	form.dao = dao
}

func (form *RecordVerificationConfirm) Validate() error {
	return validation.ValidateStruct(form,
		validation.Field(&form.Token, validation.Required, validation.By(form.checkToken)),
	)
}

func (form *RecordVerificationConfirm) checkToken(value any) error {
	v, _ := value.(string)
	if v == "" {
		return nil
	}

	claims, _ := security.ParseUnverifiedJWT(v)
	email := cast.ToString(claims["email"])
	if email == "" {
		return validation.NewError("validation_invalid_token_claims", "Missing email token claim.")
	}

	record, err := form.dao.FindUserRecordByToken(v, form.app.Settings().RecordVerificationToken.Secret)
	if err != nil || record == nil {
		return validation.NewError("validation_invalid_token", "Invalid or expired token.")
	}

	if record.Collection().Id != form.collection.Id {
		return validation.NewError("validation_token_collection_mismatch", "The provided token is for a different users collection.")
	}

	if record.Email() != email {
		return validation.NewError("validation_token_email_mismatch", "The record email doesn't match with the requested token claims.")
	}

	return nil
}

func (form *RecordVerificationConfirm) Submit(interceptors ...baseforms.InterceptorFunc[*recordmodels.Record]) (*recordmodels.Record, error) {
	if err := form.Validate(); err != nil {
		return nil, err
	}

	record, err := form.dao.FindUserRecordByToken(form.Token, form.app.Settings().RecordVerificationToken.Secret)
	if err != nil {
		return nil, err
	}

	wasVerified := record.Verified()
	if !wasVerified {
		record.SetVerified(true)
	}

	interceptorsErr := baseforms.RunInterceptors(record, func(m *recordmodels.Record) error {
		record = m
		if wasVerified {
			return nil
		}
		return form.dao.SaveRecord(m)
	}, interceptors...)
	if interceptorsErr != nil {
		return nil, interceptorsErr
	}

	return record, nil
}
