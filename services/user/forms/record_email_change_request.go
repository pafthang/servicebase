package forms

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/pafthang/servicebase/core"
	"github.com/pafthang/servicebase/daos"
	baseforms "github.com/pafthang/servicebase/services/base/forms"
	collectionmodels "github.com/pafthang/servicebase/services/collection/models"
	"github.com/pafthang/servicebase/services/mails"
	recordmodels "github.com/pafthang/servicebase/services/record/models"
)

type RecordEmailChangeRequest struct {
	app    core.App
	dao    *daos.Dao
	record *recordmodels.Record

	NewEmail string `form:"newEmail" json:"newEmail"`
}

func NewRecordEmailChangeRequest(app core.App, record *recordmodels.Record) *RecordEmailChangeRequest {
	return &RecordEmailChangeRequest{
		app:    app,
		dao:    app.Dao(),
		record: record,
	}
}

func (form *RecordEmailChangeRequest) SetDao(dao *daos.Dao) {
	form.dao = dao
}

func (form *RecordEmailChangeRequest) Validate() error {
	return validation.ValidateStruct(form,
		validation.Field(&form.NewEmail, validation.Required, validation.Length(1, 255), is.EmailFormat, validation.By(form.checkUniqueEmail)),
	)
}

func (form *RecordEmailChangeRequest) checkUniqueEmail(value any) error {
	v, _ := value.(string)
	if !form.dao.IsRecordValueUnique(form.record.Collection().Id, collectionmodels.FieldNameEmail, v) {
		return validation.NewError("validation_record_email_invalid", "User email already exists or it is invalid.")
	}
	return nil
}

func (form *RecordEmailChangeRequest) Submit(interceptors ...baseforms.InterceptorFunc[*recordmodels.Record]) error {
	if err := form.Validate(); err != nil {
		return err
	}

	return baseforms.RunInterceptors(form.record, func(m *recordmodels.Record) error {
		return mails.SendRecordChangeEmail(form.app, m, form.NewEmail)
	}, interceptors...)
}
