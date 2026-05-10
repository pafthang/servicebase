package apis

import (
	"context"
	"net/http"
	"path/filepath"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/pafthang/servicebase/core"
	backupservice "github.com/pafthang/servicebase/services/backup"
	baseforms "github.com/pafthang/servicebase/services/base/forms"
	"github.com/pafthang/servicebase/tools/filesystem"
	"github.com/pafthang/servicebase/tools/rest"
)

type MiddlewareFactory func(app core.App) echo.MiddlewareFunc
type ErrorFunc func(message string, rawErr any) error

type BindDeps struct {
	ActivityLogger         MiddlewareFactory
	RequireAdminTeamAccess func() echo.MiddlewareFunc
	NewBadRequestError     ErrorFunc
	NewForbiddenError      ErrorFunc
}

// Bind registers the backup service api endpoints and handlers.
func Bind(app core.App, rg *echo.Group, deps BindDeps) {
	api := backupAPI{
		app:     app,
		service: backupservice.New(app),
		deps:    deps,
	}

	subGroup := rg.Group("/backups", deps.ActivityLogger(app))
	subGroup.GET("", api.list, deps.RequireAdminTeamAccess())
	subGroup.POST("", api.create, deps.RequireAdminTeamAccess())
	subGroup.POST("/upload", api.upload, deps.RequireAdminTeamAccess())
	subGroup.GET("/:key", api.download)
	subGroup.DELETE("/:key", api.delete, deps.RequireAdminTeamAccess())
	subGroup.POST("/:key/restore", api.restore, deps.RequireAdminTeamAccess())
}

type backupAPI struct {
	app     core.App
	service *backupservice.Service
	deps    BindDeps
}

func (api *backupAPI) list(c echo.Context) error {
	backups, err := api.service.List()
	if err != nil {
		return api.deps.NewBadRequestError("Failed to retrieve backup items. Raw error: \n"+err.Error(), nil)
	}

	return c.JSON(http.StatusOK, backups)
}

func (api *backupAPI) create(c echo.Context) error {
	if api.service.HasActiveBackup() {
		return api.deps.NewBadRequestError("Try again later - another backup/restore process has already been started", nil)
	}

	form := api.service.NewCreateForm()
	form.SetContext(c.Request().Context())
	if err := c.Bind(form); err != nil {
		return api.deps.NewBadRequestError("An error occurred while loading the submitted data.", err)
	}

	return api.service.SubmitCreate(form, func(next baseforms.InterceptorNextFunc[string]) baseforms.InterceptorNextFunc[string] {
		return func(name string) error {
			if err := next(name); err != nil {
				return api.deps.NewBadRequestError("Failed to create backup.", err)
			}

			// We don't retrieve the generated backup file because it may not be
			// available yet due to the eventually consistent nature of some S3 providers.
			return c.NoContent(http.StatusNoContent)
		}
	})
}

func (api *backupAPI) upload(c echo.Context) error {
	files, err := rest.FindUploadedFiles(c.Request(), "file")
	if err != nil {
		return api.deps.NewBadRequestError("Missing or invalid uploaded file.", err)
	}

	form := api.service.NewUploadForm()
	form.SetContext(c.Request().Context())
	form.File = files[0]

	return api.service.SubmitUpload(form, func(next baseforms.InterceptorNextFunc[*filesystem.File]) baseforms.InterceptorNextFunc[*filesystem.File] {
		return func(file *filesystem.File) error {
			if err := next(file); err != nil {
				return api.deps.NewBadRequestError("Failed to upload backup.", err)
			}

			// We don't retrieve the generated backup file because it may not be
			// available yet due to the eventually consistent nature of some S3 providers.
			return c.NoContent(http.StatusNoContent)
		}
	})
}

func (api *backupAPI) download(c echo.Context) error {
	fileToken := c.QueryParam("token")

	if err := api.service.CanDownload(fileToken); err != nil {
		return api.deps.NewForbiddenError("Insufficient permissions to access the resource.", err)
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Minute)
	defer cancel()

	fsys, err := api.app.NewBackupsFilesystem()
	if err != nil {
		return api.deps.NewBadRequestError("Failed to load backups filesystem.", err)
	}
	defer fsys.Close()

	fsys.SetContext(ctx)

	key := c.PathParam("key")

	br, err := fsys.GetFile(key)
	if err != nil {
		return api.deps.NewBadRequestError("Failed to retrieve backup item. Raw error: \n"+err.Error(), nil)
	}
	defer br.Close()

	return fsys.Serve(
		c.Response(),
		c.Request(),
		key,
		filepath.Base(key), // without the path prefix (if any)
	)
}

func (api *backupAPI) restore(c echo.Context) error {
	if api.service.HasActiveBackup() {
		return api.deps.NewBadRequestError("Try again later - another backup/restore process has already been started.", nil)
	}

	key := c.PathParam("key")

	if exists, err := api.service.BackupExists(key); !exists {
		return api.deps.NewBadRequestError("Missing or invalid backup file.", err)
	}
	api.service.RestoreAsync(key)

	return c.NoContent(http.StatusNoContent)
}

func (api *backupAPI) delete(c echo.Context) error {
	key := c.PathParam("key")

	if !api.service.CanDelete(key) {
		return api.deps.NewBadRequestError("The backup is currently being used and cannot be deleted.", nil)
	}

	if err := api.service.Delete(key); err != nil {
		return api.deps.NewBadRequestError("Invalid or already deleted backup file. Raw error: \n"+err.Error(), nil)
	}

	return c.NoContent(http.StatusNoContent)
}
