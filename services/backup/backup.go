package backup

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/pafthang/servicebase/core"
	"github.com/pafthang/servicebase/daos"
	"github.com/pafthang/servicebase/tools/archive"
	"github.com/pafthang/servicebase/tools/filesystem"
	"github.com/pafthang/servicebase/tools/inflector"
	"github.com/pafthang/servicebase/tools/osutils"
	"github.com/pafthang/servicebase/tools/security"
)

func (s *Service) CreateBackup(ctx context.Context, name string) error {
	app := s.App()

	if app.Store().Has(StoreKeyActiveBackup) {
		return errors.New("try again later - another backup/restore operation has already been started")
	}

	if name == "" {
		name = s.generateBackupName("pb_backup_")
	}

	app.Store().Set(StoreKeyActiveBackup, name)
	defer app.Store().Remove(StoreKeyActiveBackup)

	exclude := []string{core.LocalBackupsDirName, core.LocalTempDirName}

	localTempDir := filepath.Join(app.DataDir(), core.LocalTempDirName)
	if err := os.MkdirAll(localTempDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create a temp dir: %w", err)
	}

	tempPath := filepath.Join(localTempDir, "pb_backup_"+security.PseudorandomString(4))
	createErr := app.Dao().RunInTransaction(func(dataTXDao *daos.Dao) error {
		return app.LogsDao().RunInTransaction(func(logsTXDao *daos.Dao) error {
			_, _ = dataTXDao.DB().NewQuery("PRAGMA wal_checkpoint(TRUNCATE)").Execute()
			_, _ = logsTXDao.DB().NewQuery("PRAGMA wal_checkpoint(TRUNCATE)").Execute()

			return archive.Create(app.DataDir(), tempPath, exclude...)
		})
	})
	if createErr != nil {
		return createErr
	}
	defer os.Remove(tempPath)

	fsys, err := app.NewBackupsFilesystem()
	if err != nil {
		return err
	}
	defer fsys.Close()

	fsys.SetContext(ctx)

	file, err := filesystem.NewFileFromPath(tempPath)
	if err != nil {
		return err
	}
	file.OriginalName = name
	file.Name = file.OriginalName

	return fsys.UploadFile(file, file.Name)
}

func (s *Service) RestoreBackup(ctx context.Context, name string) error {
	app := s.App()

	if runtime.GOOS == "windows" {
		return errors.New("restore is not supported on windows")
	}

	if app.Store().Has(StoreKeyActiveBackup) {
		return errors.New("try again later - another backup/restore operation has already been started")
	}

	app.Store().Set(StoreKeyActiveBackup, name)
	defer app.Store().Remove(StoreKeyActiveBackup)

	fsys, err := app.NewBackupsFilesystem()
	if err != nil {
		return err
	}
	defer fsys.Close()

	fsys.SetContext(ctx)

	br, err := fsys.GetFile(name)
	if err != nil {
		return err
	}
	defer br.Close()

	localTempDir := filepath.Join(app.DataDir(), core.LocalTempDirName)
	if err := os.MkdirAll(localTempDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create a temp dir: %w", err)
	}

	tempZip, err := os.CreateTemp(localTempDir, "pb_restore_zip")
	if err != nil {
		return err
	}
	defer os.Remove(tempZip.Name())

	if _, err := io.Copy(tempZip, br); err != nil {
		return err
	}

	extractedDataDir := filepath.Join(localTempDir, "pb_restore_"+security.PseudorandomString(4))
	defer os.RemoveAll(extractedDataDir)

	if err := archive.Extract(tempZip.Name(), extractedDataDir); err != nil {
		return err
	}

	extractedDB := filepath.Join(extractedDataDir, "data.db")
	if _, err := os.Stat(extractedDB); err != nil {
		return fmt.Errorf("data.db file is missing or invalid: %w", err)
	}

	if err := os.Remove(tempZip.Name()); err != nil {
		app.Logger().Debug(
			"[RestoreBackup] Failed to remove the temp zip backup file",
			slog.String("file", tempZip.Name()),
			slog.String("error", err.Error()),
		)
	}

	exclude := []string{core.LocalBackupsDirName, core.LocalTempDirName}

	oldTempDataDir := filepath.Join(localTempDir, "old_pb_data_"+security.PseudorandomString(4))
	if err := osutils.MoveDirContent(app.DataDir(), oldTempDataDir, exclude...); err != nil {
		return fmt.Errorf("failed to move the current pb_data content to a temp location: %w", err)
	}

	if err := osutils.MoveDirContent(extractedDataDir, app.DataDir(), exclude...); err != nil {
		return fmt.Errorf("failed to move the extracted archive content to pb_data: %w", err)
	}

	revertDataDirChanges := func() error {
		if err := osutils.MoveDirContent(app.DataDir(), extractedDataDir, exclude...); err != nil {
			return fmt.Errorf("failed to revert the extracted dir change: %w", err)
		}
		if err := osutils.MoveDirContent(oldTempDataDir, app.DataDir(), exclude...); err != nil {
			return fmt.Errorf("failed to revert old pb_data dir change: %w", err)
		}
		return nil
	}

	if err := app.Restart(); err != nil {
		if revertErr := revertDataDirChanges(); revertErr != nil {
			panic(revertErr)
		}
		return fmt.Errorf("failed to restart the app process: %w", err)
	}

	return nil
}

func (s *Service) generateBackupName(prefix string) string {
	appName := inflector.Snakecase(s.App().Settings().Meta.AppName)
	if len(appName) > 50 {
		appName = appName[:50]
	}

	return fmt.Sprintf(
		"%s%s_%s.zip",
		prefix,
		appName,
		time.Now().UTC().Format("20060102150405"),
	)
}
