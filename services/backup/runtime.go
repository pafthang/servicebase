package backup

import (
	"context"
	"errors"
	"time"

	"github.com/pafthang/servicebase/services/backup/forms"
	backupmodels "github.com/pafthang/servicebase/services/backup/models"
	baseforms "github.com/pafthang/servicebase/services/base/forms"
	fileservice "github.com/pafthang/servicebase/services/file"
	teamservice "github.com/pafthang/servicebase/services/team"
	"github.com/pafthang/servicebase/tools/filesystem"
	"github.com/pafthang/servicebase/tools/types"
	"github.com/spf13/cast"
)

const StoreKeyActiveBackup = "@activeBackup"

func (s *Service) HasActiveBackup() bool {
	return s.App().Store().Has(StoreKeyActiveBackup)
}

func (s *Service) List() ([]backupmodels.BackupFileInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fsys, err := s.App().NewBackupsFilesystem()
	if err != nil {
		return nil, err
	}
	defer fsys.Close()

	fsys.SetContext(ctx)

	backups, err := fsys.List("")
	if err != nil {
		return nil, err
	}

	result := make([]backupmodels.BackupFileInfo, len(backups))
	for i, obj := range backups {
		modified, _ := types.ParseDateTime(obj.ModTime)
		result[i] = backupmodels.BackupFileInfo{
			Key:      obj.Key,
			Size:     obj.Size,
			Modified: modified,
		}
	}

	return result, nil
}

func (s *Service) NewCreateForm() *forms.BackupCreate {
	return forms.NewBackupCreate(s.App(), s.CreateBackup)
}

func (s *Service) SubmitCreate(
	form *forms.BackupCreate,
	interceptors ...baseforms.InterceptorFunc[string],
) error {
	return form.Submit(interceptors...)
}

func (s *Service) NewUploadForm() *forms.BackupUpload {
	return forms.NewBackupUpload(s.App())
}

func (s *Service) SubmitUpload(
	form *forms.BackupUpload,
	interceptors ...baseforms.InterceptorFunc[*filesystem.File],
) error {
	return form.Submit(interceptors...)
}

func (s *Service) CanDelete(key string) bool {
	return key == "" || cast.ToString(s.App().Store().Get(StoreKeyActiveBackup)) != key
}

func (s *Service) CanDownload(fileToken string) error {
	record, err := fileservice.New(s.App()).FindAuthRecordByFileToken(fileToken)
	if err != nil {
		return err
	}

	if teamservice.New(s.App()).IsAdminTeamMember(record) {
		return nil
	}

	return errors.New("insufficient permissions")
}

func (s *Service) BackupExists(key string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fsys, err := s.App().NewBackupsFilesystem()
	if err != nil {
		return false, err
	}
	defer fsys.Close()

	fsys.SetContext(ctx)

	return fsys.Exists(key)
}

func (s *Service) RestoreAsync(key string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer cancel()

		time.Sleep(1 * time.Second)

		if err := s.RestoreBackup(ctx, key); err != nil {
			s.Logger().Error("Failed to restore backup", "key", key, "error", err.Error())
		}
	}()
}

func (s *Service) Delete(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fsys, err := s.App().NewBackupsFilesystem()
	if err != nil {
		return err
	}
	defer fsys.Close()

	fsys.SetContext(ctx)

	return fsys.Delete(key)
}
