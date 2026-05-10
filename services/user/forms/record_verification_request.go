package forms

import (
	"errors"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/pafthang/servicebase/core"
	"github.com/pafthang/servicebase/daos"
	baseforms "github.com/pafthang/servicebase/services/base/forms"
	collectionmodels "github.com/pafthang/servicebase/services/collection/models"
	"github.com/pafthang/servicebase/services/mails"
	recordmodels "github.com/pafthang/servicebase/services/record/models"
	"github.com/pafthang/servicebase/tools/types"
)

type RecordVerificationRequest struct {
	app             core.App
	collection      *collectionmodels.Collection
	dao             *daos.Dao
	resendThreshold float64

	Email string `form:"email" json:"email"`
}

func NewRecordVerificationRequest(app core.App, collection *collectionmodels.Collection) *RecordVerificationRequest {
	return &RecordVerificationRequest{
		app:             app,
		dao:             app.Dao(),
		collection:      collection,
		resendThreshold: 120,
	}
}

func (form *RecordVerificationRequest) SetDao(dao *daos.Dao) {
	form.dao = dao
}

func (form *RecordVerificationRequest) Validate() error {
	return validation.ValidateStruct(form,
		validation.Field(&form.Email, validation.Required, validation.Length(1, 255), is.EmailFormat),
	)
}

func (form *RecordVerificationRequest) Submit(interceptors ...baseforms.InterceptorFunc[*recordmodels.Record]) error {
	if err := form.Validate(); err != nil {
		return err
	}

	record, err := form.dao.FindFirstRecordByData(form.collection.Id, collectionmodels.FieldNameEmail, form.Email)
	if err != nil {
		return err
	}

	if !record.Verified() {
		now := time.Now().UTC()
		lastVerificationSentAt := record.LastVerificationSentAt().Time()
		if now.Sub(lastVerificationSentAt).Seconds() < form.resendThreshold {
			return errors.New("A verification email was already sent.")
		}
	}

	return baseforms.RunInterceptors(record, func(m *recordmodels.Record) error {
		if m.Verified() {
			return nil
		}

		if err := mails.SendRecordVerification(form.app, m); err != nil {
			return err
		}

		m.SetLastVerificationSentAt(types.NowDateTime())
		return form.dao.SaveRecord(m)
	}, interceptors...)
}
