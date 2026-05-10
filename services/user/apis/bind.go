package apis

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pafthang/servicebase/core"
	"github.com/pafthang/servicebase/daos"
	baseforms "github.com/pafthang/servicebase/services/base/forms"
	collectioncache "github.com/pafthang/servicebase/services/cache"
	collectionmodels "github.com/pafthang/servicebase/services/collection/models"
	recordforms "github.com/pafthang/servicebase/services/record/forms"
	recordmodels "github.com/pafthang/servicebase/services/record/models"
	teammodels "github.com/pafthang/servicebase/services/team/models"
)

type MiddlewareFactory func(app core.App) echo.MiddlewareFunc
type MiddlewareByStringFactory func(string) echo.MiddlewareFunc
type ErrorFunc func(message string, rawErr any) error
type RequestInfoFunc func(c echo.Context) *recordmodels.RequestInfo
type HasAuthManageAccessFunc func(dao *daos.Dao, record *recordmodels.Record, requestInfo *recordmodels.RequestInfo) bool
type EnrichRecordFunc func(c echo.Context, dao *daos.Dao, record *recordmodels.Record, defaultExpands ...string) error

type BindUsersDeps struct {
	ActivityLogger                    MiddlewareFactory
	LoadFixedCollectionContext        func(app core.App, collectionName string, collectionType string) echo.MiddlewareFunc
	RequireSameFixedAuthRecord        MiddlewareByStringFactory
	RequireAdminTeamAccessOrOwnerAuth MiddlewareByStringFactory
	RequireSameContextRecordAuth      func() echo.MiddlewareFunc
	NewUnauthorizedError              ErrorFunc
	NewNotFoundError                  ErrorFunc
	NewBadRequestError                ErrorFunc
	RequestInfo                       RequestInfoFunc
	HasAuthManageAccess               HasAuthManageAccessFunc
	EnrichRecord                      EnrichRecordFunc
	ContextAuthRecordKey              string
	ListHandler                       echo.HandlerFunc
	CreateHandler                     echo.HandlerFunc
	ViewHandler                       echo.HandlerFunc
	UpdateHandler                     echo.HandlerFunc
	DeleteHandler                     echo.HandlerFunc
	AuthMethodsHandler                echo.HandlerFunc
	AuthRefreshHandler                echo.HandlerFunc
	AuthPasswordHandler               echo.HandlerFunc
	AuthOAuth2Handler                 echo.HandlerFunc
	PasswordResetRequestHandler       echo.HandlerFunc
	PasswordResetConfirmHandler       echo.HandlerFunc
	VerificationRequestHandler        echo.HandlerFunc
	VerificationConfirmHandler        echo.HandlerFunc
	EmailChangeRequestHandler         echo.HandlerFunc
	EmailChangeConfirmHandler         echo.HandlerFunc
	ListExternalAuthsHandler          echo.HandlerFunc
	UnlinkExternalAuthHandler         echo.HandlerFunc
}

// BindUsers registers dedicated users endpoints inside the user service module.
func BindUsers(app core.App, rg *echo.Group, deps BindUsersDeps) {
	group := rg.Group(
		"/users",
		deps.ActivityLogger(app),
		deps.LoadFixedCollectionContext(app, "users", collectionmodels.CollectionTypeUsers),
	)

	group.GET("", deps.ListHandler)
	group.POST("", deps.CreateHandler)
	group.GET("/me", currentUserView(app, deps), deps.RequireSameFixedAuthRecord("users"))
	group.PATCH("/me", currentUserUpdate(app, deps), deps.RequireSameFixedAuthRecord("users"))
	group.GET("/:id", deps.ViewHandler)
	group.PATCH("/:id", deps.UpdateHandler)
	group.DELETE("/:id", deps.DeleteHandler)
	group.GET("/:id/teams", listUserTeams(app, deps), deps.RequireAdminTeamAccessOrOwnerAuth("id"))

	group.GET("/auth-methods", deps.AuthMethodsHandler)
	group.POST("/auth/refresh", deps.AuthRefreshHandler, deps.RequireSameContextRecordAuth())
	group.POST("/auth/password", deps.AuthPasswordHandler)
	group.POST("/auth/oauth2", deps.AuthOAuth2Handler)
	group.POST("/password-reset/request", deps.PasswordResetRequestHandler)
	group.POST("/password-reset/confirm", deps.PasswordResetConfirmHandler)
	group.POST("/verification/request", deps.VerificationRequestHandler)
	group.POST("/verification/confirm", deps.VerificationConfirmHandler)
	group.POST("/email-change/request", deps.EmailChangeRequestHandler, deps.RequireSameContextRecordAuth())
	group.POST("/email-change/confirm", deps.EmailChangeConfirmHandler)
	group.GET("/:id/external-auths", deps.ListExternalAuthsHandler, deps.RequireAdminTeamAccessOrOwnerAuth("id"))
	group.DELETE("/:id/external-auths/:provider", deps.UnlinkExternalAuthHandler, deps.RequireAdminTeamAccessOrOwnerAuth("id"))
}

func currentUserView(app core.App, deps BindUsersDeps) echo.HandlerFunc {
	return func(c echo.Context) error {
		record, _ := c.Get(deps.ContextAuthRecordKey).(*recordmodels.Record)
		if record == nil {
			return deps.NewUnauthorizedError("The request requires valid record authorization token to be set.", nil)
		}

		freshRecord, err := app.Dao().FindRecordById(record.Collection().Id, record.Id)
		if err != nil || freshRecord == nil {
			return deps.NewNotFoundError("", err)
		}

		event := new(core.RecordViewEvent)
		event.HttpContext = c
		event.Collection = freshRecord.Collection()
		event.Record = freshRecord

		return app.OnRecordViewRequest().Trigger(event, func(e *core.RecordViewEvent) error {
			if e.HttpContext.Response().Committed {
				return nil
			}

			if err := deps.EnrichRecord(e.HttpContext, app.Dao(), e.Record); err != nil {
				app.Logger().Debug("Failed to enrich current user record", "error", err.Error(), "id", e.Record.Id)
			}

			return e.HttpContext.JSON(http.StatusOK, e.Record)
		})
	}
}

func currentUserUpdate(app core.App, deps BindUsersDeps) echo.HandlerFunc {
	return func(c echo.Context) error {
		record, _ := c.Get(deps.ContextAuthRecordKey).(*recordmodels.Record)
		if record == nil {
			return deps.NewUnauthorizedError("The request requires valid record authorization token to be set.", nil)
		}

		requestInfo := deps.RequestInfo(c)

		freshRecord, err := app.Dao().FindRecordById(record.Collection().Id, record.Id)
		if err != nil || freshRecord == nil {
			return deps.NewNotFoundError("", err)
		}

		if requestInfo.HasModifierDataKeys() {
			requestInfo.Data = freshRecord.ReplaceModifers(requestInfo.Data)
		}

		form := recordforms.NewRecordUpsert(app, freshRecord)
		form.SetFullManageAccess(requestInfo.AdminTeamAccess || deps.HasAuthManageAccess(app.Dao(), freshRecord, requestInfo))
		if err := form.LoadRequest(c.Request(), ""); err != nil {
			return deps.NewBadRequestError("Failed to load the submitted data due to invalid formatting.", err)
		}

		event := new(core.RecordUpdateEvent)
		event.HttpContext = c
		event.Collection = freshRecord.Collection()
		event.Record = freshRecord
		event.UploadedFiles = form.FilesToUpload()

		return form.Submit(func(next baseforms.InterceptorNextFunc[*recordmodels.Record]) baseforms.InterceptorNextFunc[*recordmodels.Record] {
			return func(m *recordmodels.Record) error {
				event.Record = m

				return app.OnRecordBeforeUpdateRequest().Trigger(event, func(e *core.RecordUpdateEvent) error {
					if err := next(e.Record); err != nil {
						return deps.NewBadRequestError("Failed to update record.", err)
					}

					if err := deps.EnrichRecord(e.HttpContext, app.Dao(), e.Record); err != nil {
						app.Logger().Debug("Failed to enrich updated current user record", "error", err.Error(), "id", e.Record.Id)
					}

					return app.OnRecordAfterUpdateRequest().Trigger(event, func(e *core.RecordUpdateEvent) error {
						if e.HttpContext.Response().Committed {
							return nil
						}

						return e.HttpContext.JSON(http.StatusOK, e.Record)
					})
				})
			}
		})
	}
}

func listUserTeams(app core.App, deps BindUsersDeps) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.PathParam("id")
		if userID == "" {
			return deps.NewNotFoundError("", nil)
		}

		userCollection, err := collectioncache.FindByNameOrId(app, "users")
		if err != nil || userCollection == nil {
			return deps.NewNotFoundError("", err)
		}

		memberships, err := app.Dao().FindTeamMembersByUser(userID, userCollection.Id)
		if err != nil {
			return deps.NewBadRequestError("Failed to fetch user team memberships.", err)
		}

		teams := make([]*teammodels.Team, 0, len(memberships))
		for _, membership := range memberships {
			team, err := app.Dao().FindTeamById(membership.Team)
			if err != nil || team == nil {
				continue
			}

			teams = append(teams, team)
		}

		return c.JSON(http.StatusOK, teams)
	}
}
