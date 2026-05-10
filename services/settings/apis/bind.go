package apis

import (
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/labstack/echo/v5"
	"github.com/pafthang/servicebase/core"
	baseforms "github.com/pafthang/servicebase/services/base/forms"
	settingsservice "github.com/pafthang/servicebase/services/settings"
	settingsmodels "github.com/pafthang/servicebase/services/settings/models"
)

type MiddlewareFactory func(app core.App) echo.MiddlewareFunc
type BadRequestErrorFunc func(message string, rawErr any) error

type BindDeps struct {
	ActivityLogger         MiddlewareFactory
	RequireAdminTeamAccess func() echo.MiddlewareFunc
	NewBadRequestError     BadRequestErrorFunc
}

// Bind registers the settings service routes on the provided route group.
func Bind(app core.App, rg *echo.Group, deps BindDeps) {
	api := settingsAPI{
		app:     app,
		service: settingsservice.New(app),
		deps:    deps,
	}

	subGroup := rg.Group("/settings", deps.ActivityLogger(app), deps.RequireAdminTeamAccess())
	subGroup.GET("", api.list)
	subGroup.PATCH("", api.set)
	subGroup.POST("/test/s3", api.testS3)
	subGroup.POST("/test/email", api.testEmail)
}

type settingsAPI struct {
	app     core.App
	service *settingsservice.Service
	deps    BindDeps
}

func (api *settingsAPI) list(c echo.Context) error {
	settings, err := api.service.RedactClone()
	if err != nil {
		return api.deps.NewBadRequestError("", err)
	}

	event := new(core.SettingsListEvent)
	event.HttpContext = c
	event.RedactedSettings = settings

	return api.app.OnSettingsListRequest().Trigger(event, func(e *core.SettingsListEvent) error {
		if e.HttpContext.Response().Committed {
			return nil
		}

		return e.HttpContext.JSON(http.StatusOK, e.RedactedSettings)
	})
}

func (api *settingsAPI) set(c echo.Context) error {
	form := api.service.NewUpsertForm()
	if err := c.Bind(form); err != nil {
		return api.deps.NewBadRequestError("An error occurred while loading the submitted data.", err)
	}

	event := new(core.SettingsUpdateEvent)
	event.HttpContext = c
	event.OldSettings = api.app.Settings()

	return api.service.SubmitUpsert(form, func(next baseforms.InterceptorNextFunc[*settingsmodels.Settings]) baseforms.InterceptorNextFunc[*settingsmodels.Settings] {
		return func(s *settingsmodels.Settings) error {
			event.NewSettings = s

			return api.app.OnSettingsBeforeUpdateRequest().Trigger(event, func(e *core.SettingsUpdateEvent) error {
				if err := next(e.NewSettings); err != nil {
					return api.deps.NewBadRequestError("An error occurred while submitting the form.", err)
				}

				return api.app.OnSettingsAfterUpdateRequest().Trigger(event, func(e *core.SettingsUpdateEvent) error {
					if e.HttpContext.Response().Committed {
						return nil
					}

					redactedSettings, err := api.service.RedactClone()
					if err != nil {
						return api.deps.NewBadRequestError("", err)
					}

					return e.HttpContext.JSON(http.StatusOK, redactedSettings)
				})
			})
		}
	})
}

func (api *settingsAPI) testS3(c echo.Context) error {
	form := api.service.NewTestS3Form()
	if err := c.Bind(form); err != nil {
		return api.deps.NewBadRequestError("An error occurred while loading the submitted data.", err)
	}

	if err := api.service.TestS3(form); err != nil {
		if fErr, ok := err.(validation.Errors); ok {
			return api.deps.NewBadRequestError("Failed to test the S3 filesystem.", fErr)
		}

		return api.deps.NewBadRequestError("Failed to test the S3 filesystem. Raw error: \n"+err.Error(), nil)
	}

	return c.NoContent(http.StatusNoContent)
}

func (api *settingsAPI) testEmail(c echo.Context) error {
	form := api.service.NewTestEmailForm()
	if err := c.Bind(form); err != nil {
		return api.deps.NewBadRequestError("An error occurred while loading the submitted data.", err)
	}

	if err := api.service.TestEmail(form); err != nil {
		if fErr, ok := err.(validation.Errors); ok {
			return api.deps.NewBadRequestError("Failed to send the test email.", fErr)
		}

		return api.deps.NewBadRequestError("Failed to send the test email. Raw error: \n"+err.Error(), nil)
	}

	return c.NoContent(http.StatusNoContent)
}
