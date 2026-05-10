package apis

import (
	"github.com/labstack/echo/v5"
	"github.com/pafthang/servicebase/core"
)

type MiddlewareFactory func(app core.App) echo.MiddlewareFunc
type MiddlewareByCollectionTypesFactory func(app core.App, collectionTypes ...string) echo.MiddlewareFunc
type MiddlewareByStringFactory func(string) echo.MiddlewareFunc

type BindDeps struct {
	ActivityLogger                    MiddlewareFactory
	LoadCollectionContext             MiddlewareByCollectionTypesFactory
	DisallowCanonicalCollectionAccess MiddlewareByStringFactory
	ListHandler                       echo.HandlerFunc
	ViewHandler                       echo.HandlerFunc
	CreateHandler                     echo.HandlerFunc
	UpdateHandler                     echo.HandlerFunc
	DeleteHandler                     echo.HandlerFunc
}

// Bind registers the generic record CRUD routes inside the record service module.
func Bind(app core.App, rg *echo.Group, deps BindDeps) {
	subGroup := rg.Group(
		"/collections/:collection",
		deps.ActivityLogger(app),
	)

	subGroup.GET("/records", deps.ListHandler, deps.LoadCollectionContext(app), deps.DisallowCanonicalCollectionAccess("users"))
	subGroup.GET("/records/:id", deps.ViewHandler, deps.LoadCollectionContext(app), deps.DisallowCanonicalCollectionAccess("users"))
	subGroup.POST("/records", deps.CreateHandler, deps.LoadCollectionContext(app, "base", "auth"), deps.DisallowCanonicalCollectionAccess("users"))
	subGroup.PATCH("/records/:id", deps.UpdateHandler, deps.LoadCollectionContext(app, "base", "auth"), deps.DisallowCanonicalCollectionAccess("users"))
	subGroup.DELETE("/records/:id", deps.DeleteHandler, deps.LoadCollectionContext(app, "base", "auth"), deps.DisallowCanonicalCollectionAccess("users"))
}
