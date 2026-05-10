package file

import (
	"archive/zip"
	"fmt"
	"io"
	"path/filepath"

	"github.com/pafthang/servicebase/core"
	filemodels "github.com/pafthang/servicebase/services/file/models"
)

func StreamSelectionAsZip(userId string, folderIds []string, fileIds []string, w io.Writer) error {
	return StreamSelectionAsZipWithApp(nil, userId, folderIds, fileIds, w)
}

func StreamSelectionAsZipWithApp(app core.App, userId string, folderIds []string, fileIds []string, w io.Writer) error {
	logger := defaultLogger(app)
	logger.Info("Starting ZIP stream", "userId", userId, "foldersCount", len(folderIds), "filesCount", len(fileIds))

	zw := zip.NewWriter(w)

	if app == nil {
		return fmt.Errorf("app is required")
	}

	filesCol, err := app.Dao().FindCollectionByNameOrId("files")
	if err != nil {
		logger.Error("Failed to find files collection", "error", err)
		return fmt.Errorf("failed to find files collection: %w", err)
	}

	totalFiles := 0

	for _, fileId := range fileIds {
		if fileId == "" {
			continue
		}
		file, err := findFileByID(app, fileId)
		if err != nil {
			logger.Warn("File not found for zip", "fileId", fileId)
			continue
		}

		if file.User != userId {
			logger.Warn("Unauthorized individual file skip", "fileId", fileId, "owner", file.User, "requester", userId)
			continue
		}

		if err := addFileToZip(app, filesCol.Id, zw, file, ""); err != nil {
			logger.Error("Error adding individual file to zip", "error", err, "fileId", file.Id)
			continue
		}
		totalFiles++
	}

	for _, folderId := range folderIds {
		if folderId == "" {
			continue
		}
		folder, err := findFolderByID(app, folderId)
		if err != nil {
			logger.Warn("Folder not found for zip", "folderId", folderId)
			continue
		}

		if folder.User != userId {
			logger.Warn("Unauthorized folder skip", "folderId", folderId, "owner", folder.User, "requester", userId)
			continue
		}

		count, err := addFolderToZip(app, filesCol.Id, userId, folder, "", zw)
		if err != nil {
			logger.Error("Error processing folder for zip", "error", err, "folderId", folder.Id)
			continue
		}
		totalFiles += count
	}

	if err := zw.Close(); err != nil {
		logger.Error("Error closing zip writer", "error", err)
		return err
	}

	logger.Info("ZIP stream completed successfully", "totalFiles", totalFiles)
	return nil
}

func addFolderToZip(app core.App, filesColId string, userId string, folder *filemodels.Folder, basePath string, zw *zip.Writer) (int, error) {
	currentZipDir := filepath.Join(basePath, folder.Name)
	fileCount := 0

	files, err := listFilesByUserFolder(app, userId, folder.Id)

	if err == nil && len(files) > 0 {
		for _, file := range files {
			if err := addFileToZip(app, filesColId, zw, file, currentZipDir); err != nil {
				return fileCount, err
			}
			fileCount++
		}
	}

	subfolders, err := listFoldersByUserParent(app, userId, folder.Id)

	if err == nil && len(subfolders) > 0 {
		for _, subfolder := range subfolders {
			count, err := addFolderToZip(app, filesColId, userId, subfolder, currentZipDir, zw)
			if err != nil {
				return fileCount, err
			}
			fileCount += count
		}
	}

	return fileCount, nil
}

func addFileToZip(app core.App, filesColId string, zw *zip.Writer, file *filemodels.File, zipDir string) error {
	logger := defaultLogger(app)
	fs, err := app.NewFilesystem()
	if err != nil {
		return fmt.Errorf("failed to initialize filesystem: %w", err)
	}
	defer fs.Close()

	if file.File == "" {
		logger.Warn("File field is empty in database record, skipping", "fileId", file.Id)
		return nil
	}

	storagePath := filesColId + "/" + file.Id + "/" + file.File

	src, err := fs.GetFile(storagePath)
	if err != nil {
		src, err = fs.GetFile("files/" + file.Id + "/" + file.File)
		if err != nil {
			logger.Warn("File skip: not found in storage", "primaryPath", storagePath, "fileId", file.Id, "error", err)
			return nil
		}
	}
	defer src.Close()

	targetFilename := file.OriginalFilename
	if targetFilename == "" {
		targetFilename = file.Filename
	}
	if targetFilename == "" {
		targetFilename = file.File
	}

	zipPath := filepath.Join(zipDir, targetFilename)

	header := &zip.FileHeader{
		Name:   zipPath,
		Method: zip.Deflate,
	}
	header.Modified = file.Updated.Time()

	writer, err := zw.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("failed to create zip header for %s: %w", zipPath, err)
	}

	if _, err := io.Copy(writer, src); err != nil {
		return fmt.Errorf("failed to copy file %s to zip: %w", zipPath, err)
	}

	return nil
}
