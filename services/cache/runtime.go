package cache

import (
	"strings"

	"github.com/pafthang/servicebase/core"
	collectionmodels "github.com/pafthang/servicebase/services/collection/models"
)

const StoreCachedCollectionsKey = "@cachedCollectionsContext"

func Reload(app core.App) error {
	collections := []*collectionmodels.Collection{}

	if err := app.Dao().CollectionQuery().All(&collections); err != nil {
		return err
	}

	app.Store().Set(StoreCachedCollectionsKey, collections)

	return nil
}

func FindByNameOrId(app core.App, nameOrId string) (*collectionmodels.Collection, error) {
	collections, _ := app.Store().Get(StoreCachedCollectionsKey).([]*collectionmodels.Collection)

	for _, c := range collections {
		if strings.EqualFold(c.Name, nameOrId) || c.Id == nameOrId {
			return c, nil
		}
	}

	found, err := app.Dao().FindCollectionByNameOrId(nameOrId)
	if err != nil {
		return nil, err
	}

	_ = Reload(app)

	return found, nil
}
