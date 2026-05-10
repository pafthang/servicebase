// Package apis owns file service HTTP bindings.
package apis

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/pafthang/servicebase/core"
	collectionmodels "github.com/pafthang/servicebase/services/collection/models"
	fileservice "github.com/pafthang/servicebase/services/file"
	recordmodels "github.com/pafthang/servicebase/services/record/models"
	teamservice "github.com/pafthang/servicebase/services/team"
	"github.com/pafthang/servicebase/tools/filesystem"
	"github.com/pafthang/servicebase/tools/list"
	"golang.org/x/sync/semaphore"
	"golang.org/x/sync/singleflight"
)

const (
	contextAuthRecordKey = "authRecord"
	contextCollectionKey = "collection"
)

var imageContentTypes = []string{"image/png", "image/jpg", "image/jpeg", "image/gif"}
var defaultThumbSizes = []string{"100x100"}

type ProjectDeps struct {
	CurrentUserID          func(echo.Context) (string, bool)
	CurrentUserIDFromToken func(core.App, string) (string, bool)
	GenerateToken          func(string) (string, error)
}

// Bind registers the core /files API endpoints.
func Bind(app core.App, rg *echo.Group, loadCollection echo.MiddlewareFunc) {
	api := fileAPI{
		app:             app,
		service:         fileservice.New(app),
		thumbGenSem:     semaphore.NewWeighted(int64(runtime.NumCPU() + 2)),
		thumbGenPending: new(singleflight.Group),
		thumbGenMaxWait: 60 * time.Second,
	}

	subGroup := rg.Group("/files")
	subGroup.POST("/token", api.fileToken)
	subGroup.HEAD("/:collection/:recordId/:filename", api.download, loadCollection)
	subGroup.GET("/:collection/:recordId/:filename", api.download, loadCollection)
}

// BindProject registers project file/export endpoints.
func BindProject(app core.App, rg *echo.Group, deps ProjectDeps) {
	rg.GET("/custom/files/download-zip", func(c echo.Context) error {
		userID, ok := "", false
		if deps.CurrentUserID != nil {
			userID, ok = deps.CurrentUserID(c)
		}
		if !ok {
			authToken := c.Request().Header.Get("Authorization")
			if authToken == "" {
				authToken = c.QueryParam("token")
			}
			if deps.CurrentUserIDFromToken != nil {
				userID, ok = deps.CurrentUserIDFromToken(app, authToken)
			}
			if !ok {
				return httpError(http.StatusUnauthorized, "Authentication required.", nil)
			}
		}

		folderIDs := splitCSVParam(c.QueryParam("folderIds"))
		fileIDs := splitCSVParam(c.QueryParam("fileIds"))
		if len(folderIDs) == 0 && len(fileIDs) == 0 {
			return httpError(http.StatusBadRequest, "No folders or files selected for download.", nil)
		}

		suffix := "archive"
		if deps.GenerateToken != nil {
			if token, err := deps.GenerateToken("zip"); err == nil && len(token) >= 8 {
				suffix = token[:8]
			}
		}

		c.Response().Header().Set(echo.HeaderContentType, "application/zip")
		c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=\"download-%s.zip\"", suffix))
		c.Response().Header().Set(echo.HeaderCacheControl, "no-cache")
		c.Response().Header().Set("Connection", "keep-alive")

		if err := fileservice.StreamSelectionAsZipWithApp(app, userID, folderIDs, fileIDs, c.Response()); err != nil {
			return httpError(http.StatusInternalServerError, "Failed to generate ZIP archive.", err)
		}
		return nil
	})
}

type fileAPI struct {
	app             core.App
	service         *fileservice.Service
	thumbGenSem     *semaphore.Weighted
	thumbGenPending *singleflight.Group
	thumbGenMaxWait time.Duration
}

func httpError(status int, message string, data any) error {
	err := echo.NewHTTPError(status, message)
	if internal, ok := data.(error); ok {
		err.Internal = internal
	}
	return err
}

func (api *fileAPI) fileToken(c echo.Context) error {
	event := new(core.FileTokenEvent)
	event.HttpContext = c

	if record, _ := c.Get(contextAuthRecordKey).(*recordmodels.Record); record != nil {
		event.Model = record
		event.Token, _ = api.service.NewFileToken(record)
	}

	return api.app.OnFileBeforeTokenRequest().Trigger(event, func(e *core.FileTokenEvent) error {
		if e.Model == nil || e.Token == "" {
			return httpError(http.StatusBadRequest, "Failed to generate file token.", nil)
		}
		return api.app.OnFileAfterTokenRequest().Trigger(event, func(e *core.FileTokenEvent) error {
			if e.HttpContext.Response().Committed {
				return nil
			}
			return e.HttpContext.JSON(http.StatusOK, map[string]string{"token": e.Token})
		})
	})
}

func (api *fileAPI) download(c echo.Context) error {
	collection, _ := c.Get(contextCollectionKey).(*collectionmodels.Collection)
	if collection == nil {
		return httpError(http.StatusNotFound, "", nil)
	}

	recordID := c.PathParam("recordId")
	if recordID == "" {
		return httpError(http.StatusNotFound, "", nil)
	}

	record, err := api.app.Dao().FindRecordById(collection.Id, recordID)
	if err != nil {
		return httpError(http.StatusNotFound, "", err)
	}

	filename := c.PathParam("filename")
	fileField := record.FindFileFieldByFile(filename)
	if fileField == nil {
		return httpError(http.StatusNotFound, "", nil)
	}

	options, ok := fileField.Options.(*collectionmodels.FileOptions)
	if !ok {
		return httpError(http.StatusBadRequest, "", errors.New("failed to load file options"))
	}

	if options.Protected {
		authRecord, _ := api.service.FindAuthRecordByFileToken(c.QueryParam("token"))
		requestInfo := protectedFileRequestInfo(c, api.app, authRecord)
		if ok, _ := api.app.Dao().CanAccessRecord(record, requestInfo, record.Collection().ViewRule); !ok {
			return httpError(http.StatusForbidden, "Insufficient permissions to access the file resource.", nil)
		}
	}

	baseFilesPath := record.BaseFilesPath()
	if collection.IsView() {
		fileRecord, err := api.app.Dao().FindRecordByViewFile(collection.Id, fileField.Name, filename)
		if err != nil {
			return httpError(http.StatusNotFound, "", fmt.Errorf("failed to fetch view file field record: %w", err))
		}
		baseFilesPath = fileRecord.BaseFilesPath()
	}

	fsys, err := api.app.NewFilesystem()
	if err != nil {
		return httpError(http.StatusBadRequest, "Filesystem initialization failure.", err)
	}
	defer fsys.Close()

	originalPath := baseFilesPath + "/" + filename
	servedPath := originalPath
	servedName := filename

	thumbSize := c.QueryParam("thumb")
	if thumbSize != "" && (list.ExistInSlice(thumbSize, defaultThumbSizes) || list.ExistInSlice(thumbSize, options.Thumbs)) {
		oAttrs, oAttrsErr := fsys.Attributes(originalPath)
		if oAttrsErr != nil {
			return httpError(http.StatusNotFound, "", oAttrsErr)
		}
		if list.ExistInSlice(oAttrs.ContentType, imageContentTypes) {
			servedName = thumbSize + "_" + filename
			servedPath = baseFilesPath + "/thumbs_" + filename + "/" + servedName
			if exists, _ := fsys.Exists(servedPath); !exists {
				if err := api.createThumb(c, fsys, originalPath, servedPath, thumbSize); err != nil {
					api.app.Logger().Warn("Fallback to original - failed to create thumb "+servedName, slog.Any("error", err), slog.String("original", originalPath), slog.String("thumb", servedPath))
					servedName = filename
					servedPath = originalPath
				}
			}
		}
	}

	event := new(core.FileDownloadEvent)
	event.HttpContext = c
	event.Collection = collection
	event.Record = record
	event.FileField = fileField
	event.ServedPath = servedPath
	event.ServedName = servedName

	c.Response().Header().Del("X-Frame-Options")

	return api.app.OnFileDownloadRequest().Trigger(event, func(e *core.FileDownloadEvent) error {
		if e.HttpContext.Response().Committed {
			return nil
		}
		if err := fsys.Serve(e.HttpContext.Response(), e.HttpContext.Request(), e.ServedPath, e.ServedName); err != nil {
			return httpError(http.StatusNotFound, "", err)
		}
		return nil
	})
}

func protectedFileRequestInfo(c echo.Context, app core.App, authRecord *recordmodels.Record) *recordmodels.RequestInfo {
	requestInfo := &recordmodels.RequestInfo{
		Context:         recordmodels.RequestInfoContextProtectedFile,
		Method:          c.Request().Method,
		Query:           map[string]any{},
		Data:            map[string]any{},
		Headers:         map[string]any{},
		AuthRecord:      authRecord,
		AdminTeamAccess: false,
	}
	if authRecord != nil {
		requestInfo.AdminTeamAccess = teamservice.New(app).HasAdminTeamAccess(authRecord)
	}
	return requestInfo
}

func (api *fileAPI) createThumb(c echo.Context, fsys *filesystem.System, originalPath string, thumbPath string, thumbSize string) error {
	ch := api.thumbGenPending.DoChan(thumbPath, func() (any, error) {
		ctx, cancel := context.WithTimeout(c.Request().Context(), api.thumbGenMaxWait)
		defer cancel()
		if err := api.thumbGenSem.Acquire(ctx, 1); err != nil {
			return nil, err
		}
		defer api.thumbGenSem.Release(1)
		return nil, fsys.CreateThumb(originalPath, thumbPath, thumbSize)
	})
	res := <-ch
	api.thumbGenPending.Forget(thumbPath)
	return res.Err
}

func splitCSVParam(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}
