package forms

import (
	"context"
	"fmt"
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pafthang/servicebase/core"
	"github.com/pafthang/servicebase/services/base/forms"
)

var backupNameRegex = regexp.MustCompile(`^[a-z0-9_-]+\.zip$`)

type CreateBackupFunc func(ctx context.Context, name string) error

type BackupCreate struct {
	app          core.App
	ctx          context.Context
	createBackup CreateBackupFunc

	Name string `form:"name" json:"name"`
}

func NewBackupCreate(app core.App, createBackup CreateBackupFunc) *BackupCreate {
	return &BackupCreate{
		app:          app,
		ctx:          context.Background(),
		createBackup: createBackup,
	}
}

func (form *BackupCreate) SetContext(ctx context.Context) {
	form.ctx = ctx
}

func (form *BackupCreate) Validate() error {
	return validation.ValidateStruct(form,
		validation.Field(
			&form.Name,
			validation.Length(1, 100),
			validation.Match(backupNameRegex),
			validation.By(form.checkUniqueName),
		),
	)
}

func (form *BackupCreate) checkUniqueName(value any) error {
	v, _ := value.(string)
	if v == "" {
		return nil
	}

	fsys, err := form.app.NewBackupsFilesystem()
	if err != nil {
		return err
	}
	defer fsys.Close()

	fsys.SetContext(form.ctx)

	if exists, err := fsys.Exists(v); err != nil || exists {
		return validation.NewError("validation_backup_name_exists", "The backup file name is invalid or already exists.")
	}

	return nil
}

func (form *BackupCreate) Submit(interceptors ...forms.InterceptorFunc[string]) error {
	if err := form.Validate(); err != nil {
		return err
	}
	if form.createBackup == nil {
		return fmt.Errorf("create backup handler is not configured")
	}

	return forms.RunInterceptors(form.Name, func(name string) error {
		return form.createBackup(form.ctx, name)
	}, interceptors...)
}
