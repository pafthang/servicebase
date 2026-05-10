package apis

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pafthang/servicebase/core"
	"github.com/pafthang/servicebase/daos"
	basemodels "github.com/pafthang/servicebase/services/base/models"
	logservice "github.com/pafthang/servicebase/services/log"
	logmodels "github.com/pafthang/servicebase/services/log/models"
	"github.com/pafthang/servicebase/tools/search"
	"github.com/pocketbase/dbx"
)

var logFilterFields = []string{"rowid", "id", "created", "updated", "level", "message", "data", `^data\.[\w\.\:]*\w+$`}

func BindSystem(app core.App, rg *echo.Group, service *logservice.Service) {
	api := logsAPI{service: service}
	rg.GET("/logs", api.list)
	rg.GET("/logs/stats", api.stats)
	rg.GET("/logs/:id", api.view)
}

type logsAPI struct{ service *logservice.Service }

func (api *logsAPI) list(c echo.Context) error {
	result, err := api.service.List(c.QueryParams().Encode(), search.NewSimpleFieldResolver(logFilterFields...))
	if err != nil {
		return httpError(http.StatusBadRequest, "", err)
	}
	return c.JSON(http.StatusOK, result)
}

func (api *logsAPI) stats(c echo.Context) error {
	fieldResolver := search.NewSimpleFieldResolver(logFilterFields...)
	filter := c.QueryParam(search.FilterQueryParam)
	var expr dbx.Expression
	if filter != "" {
		var err error
		expr, err = search.FilterData(filter).BuildExpr(fieldResolver)
		if err != nil {
			return httpError(http.StatusBadRequest, "Invalid filter format.", err)
		}
	}
	stats, err := api.service.Stats(expr)
	if err != nil {
		return httpError(http.StatusBadRequest, "Failed to generate logs stats.", err)
	}
	return c.JSON(http.StatusOK, stats)
}

func (api *logsAPI) view(c echo.Context) error {
	id := c.PathParam("id")
	if id == "" {
		return httpError(http.StatusNotFound, "", nil)
	}
	log, err := api.service.FindByID(id)
	if err != nil || log == nil {
		return httpError(http.StatusNotFound, "", err)
	}
	return c.JSON(http.StatusOK, log)
}

func BindProject(app core.App, rg *echo.Group, service *logservice.Service, protectedMiddleware ...echo.MiddlewareFunc) {
	protected := rg.Group("/logging", protectedMiddleware...)
	protected.POST("/projects", func(c echo.Context) error {
		userID, ok := currentProjectUserID(c)
		if !ok {
			return httpError(http.StatusUnauthorized, "Authentication required.", nil)
		}

		var body struct {
			Name      string `json:"name"`
			Slug      string `json:"slug"`
			Retention int    `json:"retention"`
		}
		if err := c.Bind(&body); err != nil {
			return httpError(http.StatusBadRequest, "Invalid request body.", err)
		}
		if body.Name == "" || body.Slug == "" {
			return httpError(http.StatusBadRequest, "Name and slug are required.", nil)
		}

		var plaintextToken string
		var project *logmodels.LoggingProject
		err := app.Dao().RunInTransaction(func(txDao *daos.Dao) error {
			plaintext, err := generateSecureProjectToken("dev")
			if err != nil {
				return err
			}
			plaintextToken = plaintext
			devToken := &basemodels.DevToken{User: userID, Token: plaintextToken, Name: "Logging: " + body.Name, Environment: "production", IsActive: true}
			devToken.RefreshId()
			devToken.RefreshCreated()
			devToken.RefreshUpdated()
			if err := txDao.Save(devToken); err != nil {
				return err
			}
			project = &logmodels.LoggingProject{Name: body.Name, Slug: body.Slug, Retention: body.Retention, Active: true, DevToken: devToken.Id}
			project.RefreshId()
			project.RefreshCreated()
			project.RefreshUpdated()
			return txDao.Save(project)
		})
		if err != nil {
			return httpError(http.StatusInternalServerError, "Failed to create logging project.", err)
		}
		return c.JSON(http.StatusCreated, map[string]any{"project": project, "token": plaintextToken})
	})

	ingest := rg.Group("/logs", requireProjectLoggingAPIKey(app))
	ingest.POST("/ingest", func(c echo.Context) error {
		project, ok := currentProjectLoggingProject(c)
		if !ok {
			return httpError(http.StatusForbidden, "Project ID not found in context.", nil)
		}
		var req struct {
			logservice.LogEntry
			Logs []logservice.LogEntry `json:"logs"`
		}
		if err := c.Bind(&req); err != nil {
			return httpError(http.StatusBadRequest, "Invalid request body.", err)
		}
		var entries []logservice.LogEntry
		if len(req.Logs) > 0 {
			entries = req.Logs
		} else if req.Message != "" {
			entries = []logservice.LogEntry{req.LogEntry}
		} else {
			return httpError(http.StatusUnprocessableEntity, "No logs provided.", nil)
		}
		result, err := service.IngestLogs(c.Request().Context(), project.Id, entries)
		if err != nil {
			return httpError(http.StatusBadRequest, "Failed to ingest logs.", err)
		}
		status := http.StatusOK
		if result.Failed > 0 {
			status = http.StatusMultiStatus
		}
		return c.JSON(status, result)
	})
}
