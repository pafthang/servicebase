package apis

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pafthang/servicebase/core"
	healthservice "github.com/pafthang/servicebase/services/health"
)

// Bind registers the health service routes on the provided route group.
func Bind(app core.App, rg *echo.Group) {
	api := healthAPI{service: healthservice.New(app)}

	subGroup := rg.Group("/health")
	subGroup.HEAD("", api.healthCheck)
	subGroup.GET("", api.healthCheck)
}

type healthAPI struct {
	service *healthservice.Service
}

type healthCheckResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	Data    struct {
		CanBackup bool `json:"canBackup"`
	} `json:"data"`
}

func (api *healthAPI) healthCheck(c echo.Context) error {
	if c.Request().Method == http.MethodHead {
		return c.NoContent(http.StatusOK)
	}

	resp := new(healthCheckResponse)
	resp.Code = http.StatusOK
	resp.Message = "API is healthy."
	resp.Data.CanBackup = api.service.Check().CanBackup

	return c.JSON(http.StatusOK, resp)
}
