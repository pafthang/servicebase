package forms

import (
	"errors"
	"fmt"
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

type RecordPasswordResetRequest struct {
	app             core.App
	dao             *daos.Dao
	collection      *collectionmodels.Collection
	resendThreshold float64

	Email string `form:"email" json:"email"`
}

func NewRecordPasswordResetRequest(app core.App, collection *collectionmodels.Collection) *RecordPasswordResetRequest {
	return &RecordPasswordResetRequest{
		app:             app,
		dao:             app.Dao(),
		collection:      collection,
		resendThreshold: 120,
	}
}

func (form *RecordPasswordResetRequest) SetDao(dao *daos.Dao) {
	form.dao = dao
}

func (form *RecordPasswordResetRequest) Validate() error {
	return validation.ValidateStruct(form,
		validation.Field(&form.Email, validation.Required, validation.Length(1, 255), is.EmailFormat),
	)
}

func (form *RecordPasswordResetRequest) Submit(interceptors ...baseforms.InterceptorFunc[*recordmodels.Record]) error {
	if err := form.Validate(); err != nil {
		return err
	}

	authRecord, err := form.dao.FindUserRecordByEmail(form.collection.Id, form.Email)
	if err != nil {
		return fmt.Errorf("failed to fetch %s record with email %s: %w", form.collection.Id, form.Email, err)
	}

	now := time.Now().UTC()
	lastResetSentAt := authRecord.LastResetSentAt().Time()
	if now.Sub(lastResetSentAt).Seconds() < form.resendThreshold {
		return errors.New("You've already requested a password reset.")
	}

	return baseforms.RunInterceptors(authRecord, func(m *recordmodels.Record) error {
		if err := mails.SendRecordPasswordReset(form.app, m); err != nil {
			return err
		}

		m.Set(collectionmodels.FieldNameLastResetSentAt, types.NowDateTime())
		return form.dao.SaveRecord(m)
	}, interceptors...)
}
