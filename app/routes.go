package app

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/pafthang/servicebase/core"
	backupapis "github.com/pafthang/servicebase/services/backup/apis"
	collectioncache "github.com/pafthang/servicebase/services/cache"
	collectionapis "github.com/pafthang/servicebase/services/collection/apis"
	collectionmodels "github.com/pafthang/servicebase/services/collection/models"
	cronapis "github.com/pafthang/servicebase/services/cron/apis"
	fileapis "github.com/pafthang/servicebase/services/file/apis"
	healthapis "github.com/pafthang/servicebase/services/health/apis"
	logapis "github.com/pafthang/servicebase/services/log/apis"
	realtimesvc "github.com/pafthang/servicebase/services/realtime"
	realtimeapis "github.com/pafthang/servicebase/services/realtime/apis"
	recordmodels "github.com/pafthang/servicebase/services/record/models"
	settingsapis "github.com/pafthang/servicebase/services/settings/apis"
	teamservice "github.com/pafthang/servicebase/services/team"
	updaterapis "github.com/pafthang/servicebase/services/updater/apis"
	"github.com/pafthang/servicebase/tools/list"
	"github.com/pafthang/servicebase/tools/rest"
	"github.com/pafthang/servicebase/tools/security"
	"github.com/pafthang/servicebase/tools/tokens"
)

const (
	ContextAdminTeamAccessKey string = "adminTeamAccess"
	ContextAuthRecordKey      string = "authRecord"
	ContextCollectionKey      string = "collection"
	ContextExecStartKey       string = "execStart"
	fieldsQueryParam          string = "fields"
)

func InitRouter(app core.App, services *Services) (*echo.Echo, error) {
	e := echo.New()
	e.Debug = false
	e.Binder = &rest.MultiBinder{}
	e.JSONSerializer = &rest.Serializer{FieldsParam: fieldsQueryParam}

	e.ResetRouterCreator(func(ec *echo.Echo) echo.Router {
		return echo.NewRouter(echo.RouterConfig{
			UnescapePathParamValues: true,
			AllowOverwritingRoute:   true,
		})
	})

	e.Pre(middleware.RemoveTrailingSlashWithConfig(middleware.RemoveTrailingSlashConfig{
		Skipper: func(c echo.Context) bool {
			return !strings.HasPrefix(c.Request().URL.Path, "/api/")
		},
	}))
	e.Pre(LoadAuthContext(app))
	e.Use(middleware.Recover())
	e.Use(middleware.Secure())
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(ContextExecStartKey, time.Now())
			return next(c)
		}
	})

	e.HTTPErrorHandler = func(c echo.Context, err error) {
		handleHTTPError(app, c, err)
	}

	api := e.Group("/api", eagerRequestInfoCache(app))
	RegisterRoutes(app, e, api, services)

	api.Any("/*", func(c echo.Context) error { return echo.ErrNotFound }, ActivityLogger(app))

	return e, nil
}

func RegisterRoutes(app core.App, router *echo.Echo, api *echo.Group, services *Services) {
	if services == nil {
		return
	}

	commonBadRequest := func(message string, rawErr any) error { return NewBadRequestError(message, rawErr) }
	commonForbidden := func(message string, rawErr any) error { return NewForbiddenError(message, rawErr) }
	commonNotFound := func(message string, rawErr any) error { return NewNotFoundError(message, rawErr) }
	backupapis.Bind(app, api, backupapis.BindDeps{
		ActivityLogger:         ActivityLogger,
		RequireAdminTeamAccess: RequireAdminTeamAccess,
		NewBadRequestError:     commonBadRequest,
		NewForbiddenError:      commonForbidden,
	})
	collectionapis.Bind(app, api, collectionapis.BindDeps{
		ActivityLogger:         ActivityLogger,
		RequireAdminTeamAccess: RequireAdminTeamAccess,
		NewBadRequestError:     commonBadRequest,
		NewNotFoundError:       commonNotFound,
	})
	healthapis.Bind(app, api)
	settingsapis.Bind(app, api, settingsapis.BindDeps{
		ActivityLogger:         ActivityLogger,
		RequireAdminTeamAccess: RequireAdminTeamAccess,
		NewBadRequestError:     commonBadRequest,
	})

	cronapis.Bind(app, api, services.Cron)
	logapis.BindSystem(app, api, services.Log)
	logapis.BindProject(app, api, services.Log, RequireRecordAuth())
	updaterapis.Bind(app, api, services.Updater)

	fileapis.BindProject(app, api, fileapis.ProjectDeps{
		CurrentUserID:          CurrentUserID,
		CurrentUserIDFromToken: CurrentUserIDFromRecordToken,
		GenerateToken:          GenerateFileToken,
	})
	fileapis.Bind(app, api, LoadCollectionContext(app))

	realtimeapis.Bind(app, api, realtimesvc.New(app), ActivityLogger(app))

}

func ActivityLogger(app core.App) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(c)
		}
	}
}

func RequireRecordAuth(optCollectionNames ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			record, _ := c.Get(ContextAuthRecordKey).(*recordmodels.Record)
			if record == nil {
				return NewUnauthorizedError("The request requires valid record authorization token to be set.", nil)
			}
			if len(optCollectionNames) > 0 && !list.ExistInSlice(record.Collection().Name, optCollectionNames) {
				return NewForbiddenError("The authorized record model is not allowed to perform this action.", nil)
			}
			return next(c)
		}
	}
}

func RequireAdminTeamAccess() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !hasAdminTeamAccessContext(c) {
				return NewUnauthorizedError("The request requires valid authorization from a user with admin team access.", nil)
			}
			return next(c)
		}
	}
}

func RequireAdminTeamAccessOrRecordAuth(optCollectionNames ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if hasAdminTeamAccessContext(c) {
				return next(c)
			}

			record, _ := c.Get(ContextAuthRecordKey).(*recordmodels.Record)
			if record == nil {
				return NewUnauthorizedError("The request requires either admin team access or a valid record authorization token.", nil)
			}
			if len(optCollectionNames) > 0 && !list.ExistInSlice(record.Collection().Name, optCollectionNames) {
				return NewForbiddenError("The authorized record model is not allowed to perform this action.", nil)
			}

			return next(c)
		}
	}
}

func LoadCollectionContext(app core.App, optCollectionTypes ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			collectionName := c.PathParam("collection")
			if collectionName == "" {
				return next(c)
			}

			collection, err := collectioncache.FindByNameOrId(app, collectionName)
			if err != nil || collection == nil {
				return NewNotFoundError("Collection not found.", err)
			}
			if len(optCollectionTypes) > 0 && !list.ExistInSlice(collection.Type, optCollectionTypes) {
				return NewNotFoundError("Collection not found.", nil)
			}

			c.Set(ContextCollectionKey, collection)
			return next(c)
		}
	}
}

func LoadFixedCollectionContext(app core.App, collectionName string, collectionType string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			collection, err := collectioncache.FindByNameOrId(app, collectionName)
			if err != nil || collection == nil {
				return NewNotFoundError("Collection not found.", err)
			}
			if collectionType != "" && collection.Type != collectionType {
				return NewNotFoundError("Collection not found.", nil)
			}

			c.Set(ContextCollectionKey, collection)
			return next(c)
		}
	}
}

func LoadAuthContext(app core.App) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := strings.TrimSpace(c.Request().Header.Get("Authorization"))
			if token == "" {
				return next(c)
			}

			token = strings.TrimSpace(strings.TrimPrefix(token, "Bearer "))
			if token == "" {
				return next(c)
			}

			claims, _ := security.ParseUnverifiedJWT(token)
			tokenType, _ := claims["type"].(string)
			if tokenType != tokens.TypeAuthRecord {
				return next(c)
			}

			record, err := app.Dao().FindUserRecordByToken(token, app.Settings().RecordAuthToken.Secret)
			if err == nil && record != nil {
				c.Set(ContextAuthRecordKey, record)
				if teamservice.New(app).HasAdminTeamAccess(record) {
					c.Set(ContextAdminTeamAccessKey, true)
				}
			}

			return next(c)
		}
	}
}

func eagerRequestInfoCache(app core.App) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(c)
		}
	}
}

func CurrentUserID(c echo.Context) (string, bool) {
	record, _ := c.Get(ContextAuthRecordKey).(*recordmodels.Record)
	if record == nil || record.Id == "" {
		return "", false
	}
	return record.Id, true
}

func CurrentUserIDFromRecordToken(app core.App, token string) (string, bool) {
	token = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(token), "Bearer "))
	if token == "" {
		return "", false
	}

	record, err := app.Dao().FindUserRecordByToken(token, app.Settings().RecordAuthToken.Secret)
	if err != nil || record == nil || record.Id == "" {
		return "", false
	}

	return record.Id, true
}

func GenerateFileToken(userID string) (string, error) {
	return security.RandomString(32), nil
}

func hasAdminTeamAccessContext(c echo.Context) bool {
	v, _ := c.Get(ContextAdminTeamAccessKey).(bool)
	return v
}

func asApiError(err error) *ApiError {
	if err == nil {
		return NewApiError(http.StatusInternalServerError, "Internal server error.", nil)
	}

	var apiErr *ApiError
	if errors.As(err, &apiErr) {
		return apiErr
	}

	return NewApiError(http.StatusInternalServerError, err.Error(), err)
}

func handleHTTPError(app core.App, c echo.Context, err error) {
	if err == nil {
		return
	}

	var apiErr *ApiError
	if errors.As(err, &apiErr) {
		// keep it
	} else if v := new(echo.HTTPError); errors.As(err, &v) {
		apiErr = NewApiError(v.Code, fmt.Sprintf("%v", v.Message), v)
	} else if errors.Is(err, sql.ErrNoRows) {
		apiErr = asApiError(NewNotFoundError("", err))
	} else {
		apiErr = asApiError(NewBadRequestError("", err))
	}

	if c.Response().Committed {
		return
	}

	event := new(core.ApiErrorEvent)
	event.HttpContext = c
	event.Error = apiErr

	hookErr := app.OnBeforeApiError().Trigger(event, func(e *core.ApiErrorEvent) error {
		if e.HttpContext.Response().Committed {
			return nil
		}
		if e.HttpContext.Request().Method == http.MethodHead {
			return e.HttpContext.NoContent(apiErr.Code)
		}
		return e.HttpContext.JSON(apiErr.Code, apiErr)
	})
	if hookErr == nil {
		if err := app.OnAfterApiError().Trigger(event); err != nil {
			app.Logger().Debug("OnAfterApiError failure", slog.String("error", err.Error()))
		}
	} else {
		app.Logger().Debug("OnBeforeApiError error", slog.String("error", hookErr.Error()))
	}
}

var _ = collectionmodels.CollectionTypeBase
