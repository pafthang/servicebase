package apis

import (
	"github.com/labstack/echo/v5"
	"github.com/pafthang/servicebase/core"
	updaterservice "github.com/pafthang/servicebase/services/updater"
)

// Bind registers updater-owned HTTP endpoints.
//
// The updater service is currently CLI-only. Keep this hook so the module has
// the same public shape as other services and future update-status endpoints can
// live here instead of the root apis package.
func Bind(app core.App, rg *echo.Group, service *updaterservice.Service) {
	if app == nil || rg == nil || service == nil {
		return
	}
}
