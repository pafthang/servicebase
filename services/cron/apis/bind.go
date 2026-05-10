package apis

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/pafthang/servicebase/core"
	cronservice "github.com/pafthang/servicebase/services/cron"
	cronmodels "github.com/pafthang/servicebase/services/cron/models"
	recordmodels "github.com/pafthang/servicebase/services/record/models"
	"github.com/pocketbase/dbx"
)

const contextAuthRecordKey = "authRecord"

func currentProjectUserID(c echo.Context) (string, bool) {
	record, _ := c.Get(contextAuthRecordKey).(*recordmodels.Record)
	if record == nil || record.Id == "" {
		return "", false
	}
	return record.Id, true
}

func projectQueryInt(c echo.Context, key string, fallback int) int {
	value := c.QueryParam(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func newHTTPError(status int, message string, data any) error {
	err := echo.NewHTTPError(status, message)
	if internal, ok := data.(error); ok {
		err.Internal = internal
	}
	return err
}

func projectCronError(err error) error {
	if err == nil {
		return nil
	}
	apiErr, ok := err.(*cronservice.APIError)
	if !ok {
		return newHTTPError(http.StatusInternalServerError, "Request failed.", err)
	}
	return newHTTPError(apiErr.Status, apiErr.Message, apiErr.Err)
}

// Bind registers cron-owned project API endpoints under /crons.
func Bind(app core.App, rg *echo.Group, service *cronservice.CronService) {
	group := rg.Group("/crons")

	group.POST("/:id/test", func(c echo.Context) error {
		userID, ok := currentProjectUserID(c)
		if !ok {
			return newHTTPError(http.StatusUnauthorized, "Authentication required.", nil)
		}

		cronID := c.PathParam("id")
		if cronID == "" {
			return newHTTPError(http.StatusBadRequest, "Cron ID is required.", nil)
		}

		if err := service.TestCron(cronID, userID); err != nil {
			return projectCronError(err)
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "cron test execution triggered"})
	})

	group.POST("/:id/clone", func(c echo.Context) error {
		userID, ok := currentProjectUserID(c)
		if !ok {
			return newHTTPError(http.StatusUnauthorized, "Authentication required.", nil)
		}

		cronID := c.PathParam("id")
		if cronID == "" {
			return newHTTPError(http.StatusBadRequest, "Cron ID is required.", nil)
		}

		var body struct {
			Name string `json:"name"`
		}
		if err := c.Bind(&body); err != nil {
			return newHTTPError(http.StatusBadRequest, "Invalid request body.", err)
		}
		if body.Name == "" {
			return newHTTPError(http.StatusBadRequest, "Name is required.", nil)
		}

		cloned, err := service.CloneCron(cronID, userID, body.Name)
		if err != nil {
			return projectCronError(err)
		}

		return c.JSON(http.StatusCreated, cloned)
	})

	group.GET("/:id/stats", func(c echo.Context) error {
		userID, ok := currentProjectUserID(c)
		if !ok {
			return newHTTPError(http.StatusUnauthorized, "Authentication required.", nil)
		}

		cronID := c.PathParam("id")
		if cronID == "" {
			return newHTTPError(http.StatusBadRequest, "Cron ID is required.", nil)
		}

		if _, err := service.GetCron(cronID, userID); err != nil {
			return projectCronError(err)
		}

		days := projectQueryInt(c, "days", 30)
		if days > 90 {
			days = 90
		}

		executions, err := listProjectCronExecutions(app, cronID, userID, days)
		if err != nil {
			return newHTTPError(http.StatusInternalServerError, "Failed to load cron stats.", err)
		}

		var totalRuns, successRuns, failureRuns int
		var totalDuration int64
		var lastRun *time.Time

		for _, exec := range executions {
			totalRuns++
			if exec.Status == "success" {
				successRuns++
			} else if exec.Status == "failure" {
				failureRuns++
			}
			totalDuration += exec.DurationMs
			if !exec.CompletedAt.IsZero() {
				completedAt := exec.CompletedAt.Time()
				if lastRun == nil || completedAt.After(*lastRun) {
					lastRun = &completedAt
				}
			}
		}

		successRate := 0.0
		avgDuration := 0.0
		if totalRuns > 0 {
			successRate = float64(successRuns) / float64(totalRuns) * 100
			avgDuration = float64(totalDuration) / float64(totalRuns)
		}

		var lastRunValue *string
		if lastRun != nil {
			formatted := lastRun.Format(time.RFC3339)
			lastRunValue = &formatted
		}

		return c.JSON(http.StatusOK, map[string]any{
			"total_runs":      totalRuns,
			"success_runs":    successRuns,
			"failure_runs":    failureRuns,
			"success_rate":    successRate,
			"avg_duration_ms": avgDuration,
			"last_run":        lastRunValue,
		})
	})

	group.GET("/:id/metrics", func(c echo.Context) error {
		userID, ok := currentProjectUserID(c)
		if !ok {
			return newHTTPError(http.StatusUnauthorized, "Authentication required.", nil)
		}

		cronID := c.PathParam("id")
		if cronID == "" {
			return newHTTPError(http.StatusBadRequest, "Cron ID is required.", nil)
		}

		if _, err := service.GetCron(cronID, userID); err != nil {
			return projectCronError(err)
		}

		days := projectQueryInt(c, "days", 30)
		if days > 90 {
			days = 90
		}

		executions, err := listProjectCronExecutions(app, cronID, userID, days)
		if err != nil {
			return newHTTPError(http.StatusInternalServerError, "Failed to load cron metrics.", err)
		}

		type dataPoint struct {
			Date           string  `json:"date"`
			SuccessCount   int     `json:"success_count"`
			FailureCount   int     `json:"failure_count"`
			AvgDuration    float64 `json:"avg_duration_ms"`
			ExecutionCount int     `json:"execution_count"`
		}

		dateMap := make(map[string]*dataPoint)
		statusBreakdown := make(map[string]int)
		for _, exec := range executions {
			statusBreakdown[exec.Status]++
			dateKey := exec.Created.Time().Format("2006-01-02")
			point := dateMap[dateKey]
			if point == nil {
				point = &dataPoint{Date: dateKey}
				dateMap[dateKey] = point
			}
			point.ExecutionCount++
			point.AvgDuration += float64(exec.DurationMs)
			if exec.Status == "success" {
				point.SuccessCount++
			} else if exec.Status == "failure" {
				point.FailureCount++
			}
		}

		points := make([]dataPoint, 0, len(dateMap))
		for _, point := range dateMap {
			if point.ExecutionCount > 0 {
				point.AvgDuration = point.AvgDuration / float64(point.ExecutionCount)
			}
			points = append(points, *point)
		}
		sort.Slice(points, func(i, j int) bool { return points[i].Date < points[j].Date })

		return c.JSON(http.StatusOK, map[string]any{
			"success_rate":     points,
			"execution_count":  points,
			"duration_trend":   points,
			"error_frequency":  points,
			"status_breakdown": statusBreakdown,
		})
	})
}

func listProjectCronExecutions(app core.App, cronID, userID string, days int) ([]*cronmodels.CronExecution, error) {
	startDate := time.Now().AddDate(0, 0, -days)
	var items []*cronmodels.CronExecution
	err := app.Dao().ModelQuery(&cronmodels.CronExecution{}).
		Where(dbx.HashExp{"cron": cronID, "user": userID}).
		AndWhere(dbx.NewExp("`created` >= {:startDate}", dbx.Params{"startDate": startDate})).
		All(&items)
	return items, err
}
