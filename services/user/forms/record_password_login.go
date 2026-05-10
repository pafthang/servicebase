package forms

import (
	"database/sql"
	"errors"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/pafthang/servicebase/core"
	"github.com/pafthang/servicebase/daos"
	baseforms "github.com/pafthang/servicebase/services/base/forms"
	collectionmodels "github.com/pafthang/servicebase/services/collection/models"
	recordmodels "github.com/pafthang/servicebase/services/record/models"
)

type RecordPasswordLogin struct {
	app        core.App
	dao        *daos.Dao
	collection *collectionmodels.Collection

	Identity string `form:"identity" json:"identity"`
	Password string `form:"password" json:"password"`
}

func NewRecordPasswordLogin(app core.App, collection *collectionmodels.Collection) *RecordPasswordLogin {
	return &RecordPasswordLogin{
		app:        app,
		dao:        app.Dao(),
		collection: collection,
	}
}

func (form *RecordPasswordLogin) SetDao(dao *daos.Dao) {
	form.dao = dao
}

func (form *RecordPasswordLogin) Validate() error {
	return validation.ValidateStruct(form,
		validation.Field(&form.Identity, validation.Required, validation.Length(1, 255)),
		validation.Field(&form.Password, validation.Required, validation.Length(1, 255)),
	)
}

func (form *RecordPasswordLogin) Submit(interceptors ...baseforms.InterceptorFunc[*recordmodels.Record]) (*recordmodels.Record, error) {
	if err := form.Validate(); err != nil {
		return nil, err
	}

	userOptions := form.collection.UserOptions()

	var userRecord *recordmodels.Record
	var fetchErr error

	isEmail := is.EmailFormat.Validate(form.Identity) == nil

	if isEmail {
		if userOptions.AllowEmailAuth {
			userRecord, fetchErr = form.dao.FindUserRecordByEmail(form.collection.Id, form.Identity)
		}
	} else if userOptions.AllowUsernameAuth {
		userRecord, fetchErr = form.dao.FindUserRecordByUsername(form.collection.Id, form.Identity)
	}

	if fetchErr != nil && !errors.Is(fetchErr, sql.ErrNoRows) {
		return nil, fetchErr
	}

	interceptorsErr := baseforms.RunInterceptors(userRecord, func(m *recordmodels.Record) error {
		userRecord = m

		if userRecord == nil || !userRecord.ValidatePassword(form.Password) {
			return errors.New("Invalid login credentials.")
		}

		return nil
	}, interceptors...)
	if interceptorsErr != nil {
		return nil, interceptorsErr
	}

	return userRecord, nil
}
