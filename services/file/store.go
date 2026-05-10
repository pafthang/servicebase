package file

import (
	"fmt"

	"github.com/pafthang/servicebase/core"
	filemodels "github.com/pafthang/servicebase/services/file/models"
	"github.com/pocketbase/dbx"
)

func findFileByID(app core.App, id string) (*filemodels.File, error) {
	if app == nil {
		return nil, fmt.Errorf("app is required")
	}

	item := &filemodels.File{}
	if err := app.Dao().FindById(item, id); err != nil {
		return nil, err
	}

	return item, nil
}

func findFolderByID(app core.App, id string) (*filemodels.Folder, error) {
	if app == nil {
		return nil, fmt.Errorf("app is required")
	}

	item := &filemodels.Folder{}
	if err := app.Dao().FindById(item, id); err != nil {
		return nil, err
	}

	return item, nil
}

func listFilesByUserFolder(app core.App, userID, folderID string) ([]*filemodels.File, error) {
	if app == nil {
		return nil, fmt.Errorf("app is required")
	}

	var items []*filemodels.File
	err := app.Dao().
		ModelQuery(&filemodels.File{}).
		Where(dbx.HashExp{"user": userID, "folder": folderID}).
		All(&items)

	return items, err
}

func listFoldersByUserParent(app core.App, userID, parentID string) ([]*filemodels.Folder, error) {
	if app == nil {
		return nil, fmt.Errorf("app is required")
	}

	var items []*filemodels.Folder
	err := app.Dao().
		ModelQuery(&filemodels.Folder{}).
		Where(dbx.HashExp{"user": userID, "parent": parentID}).
		All(&items)

	return items, err
}

func findUserStorageQuota(app core.App, userID string) (*filemodels.UserStorageQuota, error) {
	if app == nil {
		return nil, fmt.Errorf("app is required")
	}

	item := &filemodels.UserStorageQuota{}
	err := app.Dao().
		ModelQuery(item).
		Where(dbx.HashExp{"user": userID}).
		Limit(1).
		One(item)
	if err != nil {
		return nil, err
	}

	return item, nil
}

func saveUserStorageQuota(app core.App, item *filemodels.UserStorageQuota) error {
	if app == nil {
		return fmt.Errorf("app is required")
	}

	return app.Dao().Save(item)
}
