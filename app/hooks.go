package app

import (
	"fmt"
	"strings"

	"github.com/pafthang/servicebase/core"
	basemodels "github.com/pafthang/servicebase/services/base/models"
	filesvc "github.com/pafthang/servicebase/services/file"
	filemodels "github.com/pafthang/servicebase/services/file/models"
	logsvc "github.com/pafthang/servicebase/services/log"
	recordmodels "github.com/pafthang/servicebase/services/record/models"
	"github.com/pafthang/servicebase/tools/security"
	"github.com/pocketbase/dbx"
)

const realtimeClientUserIDKey = "userId"

type HookConfig struct {
	App            core.App
	EncryptionKey  string
	LoggingService *logsvc.Service
}

func RegisterHooks(cfg HookConfig) error {
	if cfg.App == nil {
		return fmt.Errorf("app is required")
	}

	registerRealtimeAuthHooks(cfg.App)
	registerFileHooks(cfg.App)
	registerFolderHooks(cfg.App)

	return nil
}

func registerRealtimeAuthHooks(app core.App) {
	app.OnRealtimeConnectRequest().Add(func(e *core.RealtimeConnectEvent) error {
		if authRecord, ok := e.Client.Get(ContextAuthRecordKey).(*recordmodels.Record); ok && authRecord != nil {
			e.Client.Set(realtimeClientUserIDKey, authRecord.Id)
			return nil
		}

		authHeader := strings.TrimSpace(e.HttpContext.Request().Header.Get("Authorization"))
		if authHeader == "" {
			return nil
		}

		token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if token == "" {
			return nil
		}

		userID, err := extractUserIDFromJWT(token)
		if err != nil || userID == "" {
			return nil
		}

		e.Client.Set(realtimeClientUserIDKey, userID)
		return nil
	})

	app.OnRealtimeBeforeSubscribeRequest().Add(func(e *core.RealtimeSubscribeEvent) error {
		userID, _ := e.Client.Get(realtimeClientUserIDKey).(string)
		if userID == "" {
			if authRecord, ok := e.Client.Get(ContextAuthRecordKey).(*recordmodels.Record); ok && authRecord != nil {
				userID = authRecord.Id
			}
		}

		if userID == "" {
			for _, subscription := range e.Subscriptions {
				if strings.Contains(subscription, ":") {
					return NewForbiddenError("authentication required for realtime subscriptions", nil)
				}
			}
			return nil
		}

		for _, subscription := range e.Subscriptions {
			if !isAuthorizedRealtimeTopic(userID, subscription) {
				return NewForbiddenError("not authorized to subscribe to this topic", nil)
			}
		}

		return nil
	})
}

func registerFileHooks(app core.App) {
	app.OnModelBeforeCreate("files").Add(func(e *core.ModelEvent) error {
		file, ok := e.Model.(*filemodels.File)
		if !ok {
			return nil
		}

		return validateFileCreate(app, file)
	})

	app.OnModelAfterCreate("files").Add(func(e *core.ModelEvent) error {
		file, ok := e.Model.(*filemodels.File)
		if !ok {
			return nil
		}

		return filesvc.UpdateQuotaUsage(app, file.User, file.Size)
	})

	app.OnModelBeforeUpdate("files").Add(func(e *core.ModelEvent) error {
		file, ok := e.Model.(*filemodels.File)
		if !ok {
			return nil
		}

		return validateFileUpdate(app, file)
	})

	app.OnModelAfterDelete("files").Add(func(e *core.ModelEvent) error {
		file, ok := e.Model.(*filemodels.File)
		if !ok {
			return nil
		}

		return filesvc.UpdateQuotaUsage(app, file.User, -file.Size)
	})
}

func registerFolderHooks(app core.App) {
	app.OnModelBeforeCreate("folders").Add(func(e *core.ModelEvent) error {
		folder, ok := e.Model.(*filemodels.Folder)
		if !ok {
			return nil
		}

		return validateFolderCreate(app, folder)
	})

	app.OnModelBeforeUpdate("folders").Add(func(e *core.ModelEvent) error {
		folder, ok := e.Model.(*filemodels.Folder)
		if !ok {
			return nil
		}

		return validateFolderUpdate(app, folder)
	})

	app.OnModelBeforeDelete("folders").Add(func(e *core.ModelEvent) error {
		folder, ok := e.Model.(*filemodels.Folder)
		if !ok {
			return nil
		}

		return validateFolderDelete(app, folder)
	})
}

func validateFileCreate(app core.App, file *filemodels.File) error {
	if file.User == "" {
		return fmt.Errorf("user field is required")
	}
	if file.Filename == "" {
		return fmt.Errorf("filename is required")
	}

	sanitizedFilename, err := filesvc.SanitizeFilename(file.Filename)
	if err != nil {
		return fmt.Errorf("invalid filename: %w", err)
	}
	file.Filename = sanitizedFilename

	if file.OriginalFilename == "" {
		file.OriginalFilename = sanitizedFilename
	} else if sanitizedOriginal, err := filesvc.SanitizeFilename(file.OriginalFilename); err == nil {
		file.OriginalFilename = sanitizedOriginal
	}

	if file.Path != "" {
		finalPath, err := filesvc.BuildFilePath(file.Path, sanitizedFilename)
		if err != nil {
			return fmt.Errorf("invalid path: %w", err)
		}
		file.Path = finalPath
	} else {
		file.Path = sanitizedFilename
	}

	if file.Size <= 0 {
		return fmt.Errorf("file size must be greater than 0")
	}

	available, err := filesvc.CheckQuotaAvailable(app, file.User, file.Size)
	if err != nil {
		return fmt.Errorf("failed to check quota: %w", err)
	}
	if !available {
		return fmt.Errorf("storage quota exceeded")
	}

	return nil
}

func validateFileUpdate(app core.App, file *filemodels.File) error {
	existing, err := findExistingFile(app, file.Id)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	if file.User != existing.User {
		return fmt.Errorf("cannot change file ownership")
	}
	if file.Size != existing.Size {
		return fmt.Errorf("cannot modify file size")
	}

	if file.Filename != "" {
		sanitizedFilename, err := filesvc.SanitizeFilename(file.Filename)
		if err != nil {
			return fmt.Errorf("invalid filename: %w", err)
		}
		file.Filename = sanitizedFilename

		if file.Path != "" {
			finalPath, err := filesvc.BuildFilePath(file.Path, sanitizedFilename)
			if err != nil {
				return fmt.Errorf("invalid path: %w", err)
			}
			file.Path = finalPath
		}
	}

	if file.Path != "" && file.Filename == existing.Filename {
		if !filesvc.ValidatePath(file.Path) {
			return fmt.Errorf("invalid path")
		}

		sanitizedPath, err := filesvc.SanitizePath(file.Path)
		if err != nil {
			return fmt.Errorf("invalid path: %w", err)
		}
		file.Path = sanitizedPath
	}

	return nil
}

func validateFolderCreate(app core.App, folder *filemodels.Folder) error {
	if folder.User == "" {
		return fmt.Errorf("user field is required")
	}
	if folder.Name == "" {
		return fmt.Errorf("folder name is required")
	}

	sanitizedName, err := filesvc.SanitizeFilename(folder.Name)
	if err != nil {
		return fmt.Errorf("invalid folder name: %w", err)
	}
	folder.Name = sanitizedName

	if folder.Parent == "" {
		return nil
	}

	parent, err := findExistingFolder(app, folder.Parent)
	if err != nil {
		return fmt.Errorf("parent folder not found: %w", err)
	}
	if parent.User != folder.User {
		return fmt.Errorf("unauthorized access to parent folder")
	}

	return nil
}

func validateFolderUpdate(app core.App, folder *filemodels.Folder) error {
	existing, err := findExistingFolder(app, folder.Id)
	if err != nil {
		return fmt.Errorf("folder not found: %w", err)
	}

	if folder.User != existing.User {
		return fmt.Errorf("cannot change folder ownership")
	}

	if folder.Name != "" {
		sanitizedName, err := filesvc.SanitizeFilename(folder.Name)
		if err != nil {
			return fmt.Errorf("invalid folder name: %w", err)
		}
		folder.Name = sanitizedName
	}

	return nil
}

func validateFolderDelete(app core.App, folder *filemodels.Folder) error {
	childFolders, err := listFoldersByUserParent(app, folder.User, folder.Id)
	if err == nil && len(childFolders) > 0 {
		return fmt.Errorf("cannot delete folder: it contains subfolders")
	}

	filesInFolder, err := listFilesByUserFolder(app, folder.User, folder.Id)
	if err == nil && len(filesInFolder) > 0 {
		return fmt.Errorf("cannot delete folder: it contains files")
	}

	return nil
}

func extractUserIDFromJWT(token string) (string, error) {
	claims, err := security.ParseUnverifiedJWT(token)
	if err != nil {
		return "", err
	}

	userID, _ := claims["id"].(string)
	if userID == "" {
		return "", fmt.Errorf("token claims missing id")
	}

	return userID, nil
}

func isAuthorizedRealtimeTopic(userID, topic string) bool {
	if !strings.Contains(topic, ":") {
		return true
	}

	parts := strings.Split(topic, ":")
	if len(parts) < 2 || len(parts) > 3 {
		return false
	}

	if parts[1] == userID {
		return true
	}

	switch parts[0] {
	case "system-broadcast", "maintenance", "announcements":
		return true
	default:
		return false
	}
}

func encryptTokenModel(item *basemodels.Token, existing *basemodels.Token, encryptionKey string) error {
	if item.AccessToken != "" && (existing == nil || item.AccessToken != existing.AccessToken) {
		encrypted, err := security.Encrypt([]byte(item.AccessToken), encryptionKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt access_token: %w", err)
		}
		item.AccessToken = encrypted
	}

	if item.RefreshToken != "" && (existing == nil || item.RefreshToken != existing.RefreshToken) {
		encrypted, err := security.Encrypt([]byte(item.RefreshToken), encryptionKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt refresh_token: %w", err)
		}
		item.RefreshToken = encrypted
	}

	return nil
}

func encryptDevTokenModel(item *basemodels.DevToken, existing *basemodels.DevToken, encryptionKey string) error {
	if item.Token == "" || (existing != nil && item.Token == existing.Token) {
		return nil
	}

	encrypted, err := security.Encrypt([]byte(item.Token), encryptionKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt dev token: %w", err)
	}

	item.Token = encrypted
	return nil
}

func decryptRecordFields(record *recordmodels.Record, encryptionKey string, fields ...string) {
	if record == nil || encryptionKey == "" {
		return
	}

	for _, field := range fields {
		encryptedValue := record.GetString(field)
		if encryptedValue == "" {
			continue
		}

		decrypted, err := security.Decrypt(encryptedValue, encryptionKey)
		if err != nil {
			record.Set(field, "[decryption failed]")
			continue
		}

		record.Set(field, string(decrypted))
	}
}

func isAdminUser(app core.App, userID string) bool {
	if app == nil || userID == "" {
		return false
	}

	superuser := &basemodels.Superuser{}
	return app.Dao().FindById(superuser, userID) == nil
}

func findExistingToken(app core.App, id string) (*basemodels.Token, error) {
	item := &basemodels.Token{}
	if err := app.Dao().FindById(item, id); err != nil {
		return nil, err
	}
	return item, nil
}

func findExistingDevToken(app core.App, id string) (*basemodels.DevToken, error) {
	item := &basemodels.DevToken{}
	if err := app.Dao().FindById(item, id); err != nil {
		return nil, err
	}
	return item, nil
}

func findExistingFile(app core.App, id string) (*filemodels.File, error) {
	item := &filemodels.File{}
	if err := app.Dao().FindById(item, id); err != nil {
		return nil, err
	}
	return item, nil
}

func findExistingFolder(app core.App, id string) (*filemodels.Folder, error) {
	item := &filemodels.Folder{}
	if err := app.Dao().FindById(item, id); err != nil {
		return nil, err
	}
	return item, nil
}

func findExistingQuota(app core.App, id string) (*filemodels.UserStorageQuota, error) {
	item := &filemodels.UserStorageQuota{}
	if err := app.Dao().FindById(item, id); err != nil {
		return nil, err
	}
	return item, nil
}

func listFilesByUserFolder(app core.App, userID, folderID string) ([]*filemodels.File, error) {
	var items []*filemodels.File
	err := app.Dao().
		ModelQuery(&filemodels.File{}).
		Where(dbx.HashExp{"user": userID, "folder": folderID}).
		All(&items)
	return items, err
}

func listFoldersByUserParent(app core.App, userID, parentID string) ([]*filemodels.Folder, error) {
	var items []*filemodels.Folder
	err := app.Dao().
		ModelQuery(&filemodels.Folder{}).
		Where(dbx.HashExp{"user": userID, "parent": parentID}).
		All(&items)
	return items, err
}
