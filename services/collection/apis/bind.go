package apis

import (
	"net/http"

	collectionmodels "github.com/pafthang/servicebase/services/collection/models"

	"github.com/labstack/echo/v5"
	"github.com/pafthang/servicebase/core"
	baseforms "github.com/pafthang/servicebase/services/base/forms"
	collectionservice "github.com/pafthang/servicebase/services/collection"
)

type MiddlewareFactory func(app core.App) echo.MiddlewareFunc
type BadRequestErrorFunc func(message string, rawErr any) error
type NotFoundErrorFunc func(message string, rawErr any) error

type BindDeps struct {
	ActivityLogger         MiddlewareFactory
	RequireAdminTeamAccess func() echo.MiddlewareFunc
	NewBadRequestError     BadRequestErrorFunc
	NewNotFoundError       NotFoundErrorFunc
}

// Bind registers the collection service routes on the provided route group.
func Bind(app core.App, rg *echo.Group, deps BindDeps) {
	api := collectionAPI{
		app:     app,
		service: collectionservice.New(app),
		deps:    deps,
	}

	subGroup := rg.Group("/collections", deps.ActivityLogger(app), deps.RequireAdminTeamAccess())
	subGroup.GET("", api.list)
	subGroup.POST("", api.create)
	subGroup.GET("/:collection", api.view)
	subGroup.PATCH("/:collection", api.update)
	subGroup.DELETE("/:collection", api.delete)
	subGroup.PUT("/import", api.bulkImport)
}

type collectionAPI struct {
	app     core.App
	service *collectionservice.Service
	deps    BindDeps
}

func (api *collectionAPI) list(c echo.Context) error {
	result, collections, err := api.service.List(c.QueryParams().Encode())
	if err != nil {
		return api.deps.NewBadRequestError("", err)
	}

	event := new(core.CollectionsListEvent)
	event.HttpContext = c
	event.Collections = collections
	event.Result = result

	return api.app.OnCollectionsListRequest().Trigger(event, func(e *core.CollectionsListEvent) error {
		if e.HttpContext.Response().Committed {
			return nil
		}

		return e.HttpContext.JSON(http.StatusOK, e.Result)
	})
}

func (api *collectionAPI) view(c echo.Context) error {
	collection, err := api.service.FindByNameOrID(c.PathParam("collection"))
	if err != nil || collection == nil {
		return api.deps.NewNotFoundError("", err)
	}

	event := new(core.CollectionViewEvent)
	event.HttpContext = c
	event.Collection = collection

	return api.app.OnCollectionViewRequest().Trigger(event, func(e *core.CollectionViewEvent) error {
		if e.HttpContext.Response().Committed {
			return nil
		}

		return e.HttpContext.JSON(http.StatusOK, e.Collection)
	})
}

func (api *collectionAPI) create(c echo.Context) error {
	collection := &collectionmodels.Collection{}
	form := api.service.NewUpsertForm(collection)
	if err := c.Bind(form); err != nil {
		return api.deps.NewBadRequestError("Failed to load the submitted data due to invalid formatting.", err)
	}

	event := new(core.CollectionCreateEvent)
	event.HttpContext = c
	event.Collection = collection

	return api.service.SubmitUpsert(form, func(next baseforms.InterceptorNextFunc[*collectionmodels.Collection]) baseforms.InterceptorNextFunc[*collectionmodels.Collection] {
		return func(m *collectionmodels.Collection) error {
			event.Collection = m

			return api.app.OnCollectionBeforeCreateRequest().Trigger(event, func(e *core.CollectionCreateEvent) error {
				if err := next(e.Collection); err != nil {
					return api.deps.NewBadRequestError("Failed to create the collection.", err)
				}

				return api.app.OnCollectionAfterCreateRequest().Trigger(event, func(e *core.CollectionCreateEvent) error {
					if e.HttpContext.Response().Committed {
						return nil
					}

					return e.HttpContext.JSON(http.StatusOK, e.Collection)
				})
			})
		}
	})
}

func (api *collectionAPI) update(c echo.Context) error {
	collection, err := api.service.FindByNameOrID(c.PathParam("collection"))
	if err != nil || collection == nil {
		return api.deps.NewNotFoundError("", err)
	}

	form := api.service.NewUpsertForm(collection)
	if err := c.Bind(form); err != nil {
		return api.deps.NewBadRequestError("Failed to load the submitted data due to invalid formatting.", err)
	}

	event := new(core.CollectionUpdateEvent)
	event.HttpContext = c
	event.Collection = collection

	return api.service.SubmitUpsert(form, func(next baseforms.InterceptorNextFunc[*collectionmodels.Collection]) baseforms.InterceptorNextFunc[*collectionmodels.Collection] {
		return func(m *collectionmodels.Collection) error {
			event.Collection = m

			return api.app.OnCollectionBeforeUpdateRequest().Trigger(event, func(e *core.CollectionUpdateEvent) error {
				if err := next(e.Collection); err != nil {
					return api.deps.NewBadRequestError("Failed to update the collection.", err)
				}

				return api.app.OnCollectionAfterUpdateRequest().Trigger(event, func(e *core.CollectionUpdateEvent) error {
					if e.HttpContext.Response().Committed {
						return nil
					}

					return e.HttpContext.JSON(http.StatusOK, e.Collection)
				})
			})
		}
	})
}

func (api *collectionAPI) delete(c echo.Context) error {
	collection, err := api.service.FindByNameOrID(c.PathParam("collection"))
	if err != nil || collection == nil {
		return api.deps.NewNotFoundError("", err)
	}

	event := new(core.CollectionDeleteEvent)
	event.HttpContext = c
	event.Collection = collection

	return api.app.OnCollectionBeforeDeleteRequest().Trigger(event, func(e *core.CollectionDeleteEvent) error {
		if err := api.service.Delete(e.Collection); err != nil {
			return api.deps.NewBadRequestError("Failed to delete collection due to existing dependency.", err)
		}

		return api.app.OnCollectionAfterDeleteRequest().Trigger(event, func(e *core.CollectionDeleteEvent) error {
			if e.HttpContext.Response().Committed {
				return nil
			}

			return e.HttpContext.NoContent(http.StatusNoContent)
		})
	})
}

func (api *collectionAPI) bulkImport(c echo.Context) error {
	form := api.service.NewImportForm()
	if err := c.Bind(form); err != nil {
		return api.deps.NewBadRequestError("Failed to load the submitted data due to invalid formatting.", err)
	}

	event := new(core.CollectionsImportEvent)
	event.HttpContext = c
	event.Collections = form.Collections

	return api.service.SubmitImport(form, func(next baseforms.InterceptorNextFunc[[]*collectionmodels.Collection]) baseforms.InterceptorNextFunc[[]*collectionmodels.Collection] {
		return func(imports []*collectionmodels.Collection) error {
			event.Collections = imports

			return api.app.OnCollectionsBeforeImportRequest().Trigger(event, func(e *core.CollectionsImportEvent) error {
				if err := next(e.Collections); err != nil {
					return api.deps.NewBadRequestError("Failed to import the submitted collections.", err)
				}

				return api.app.OnCollectionsAfterImportRequest().Trigger(event, func(e *core.CollectionsImportEvent) error {
					if e.HttpContext.Response().Committed {
						return nil
					}

					return e.HttpContext.NoContent(http.StatusNoContent)
				})
			})
		}
	})
}
