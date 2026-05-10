package migrate

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pafthang/servicebase/core"
	"github.com/pafthang/servicebase/daos"
	collectionmodels "github.com/pafthang/servicebase/services/collection/models"
	migratetool "github.com/pafthang/servicebase/tools/migrate"
	"github.com/pocketbase/dbx"
)

const collectionsStoreKey = "migratecmd_collections"

// afterCollectionChange handles the automigration snapshot generation on
// collection change event (create/update/delete).
func (s *Service) afterCollectionChange() func(*core.ModelEvent) error {
	return func(e *core.ModelEvent) error {
		if e.Model.TableName() != "_collections" {
			return nil
		}

		oldCollections, err := s.getCachedCollections()
		if err != nil {
			return err
		}

		old := oldCollections[e.Model.GetId()]

		newCollection, err := s.App().Dao().FindCollectionByNameOrId(e.Model.GetId())
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		template, templateErr := s.diffTemplate(newCollection, old)
		if templateErr != nil {
			if errors.Is(templateErr, emptyTemplateErr) {
				return nil
			}
			return fmt.Errorf("failed to resolve template: %w", templateErr)
		}

		var action string
		switch {
		case newCollection == nil:
			action = "deleted_" + old.Name
		case old == nil:
			action = "created_" + newCollection.Name
		default:
			action = "updated_" + old.Name
		}

		name := fmt.Sprintf("%d_%s.%s", time.Now().Unix(), action, s.templateLang())
		filePath := filepath.Join(s.config.Dir, name)

		return s.App().Dao().RunInTransaction(func(txDao *daos.Dao) error {
			_, err := txDao.DB().Insert(migratetool.DefaultMigrationsTable, dbx.Params{
				"file":    name,
				"applied": time.Now().UnixMicro(),
			}).Execute()
			if err != nil {
				return err
			}

			if err := os.MkdirAll(s.config.Dir, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create migration dir: %w", err)
			}

			if err := os.WriteFile(filePath, []byte(template), 0644); err != nil {
				return fmt.Errorf("failed to save automigrate file: %w", err)
			}

			s.updateSingleCachedCollection(newCollection, old)
			return nil
		})
	}
}

func (s *Service) updateSingleCachedCollection(newCollection, old *collectionmodels.Collection) {
	cached, _ := s.App().Store().Get(collectionsStoreKey).(map[string]*collectionmodels.Collection)
	if cached == nil {
		cached = map[string]*collectionmodels.Collection{}
	}

	switch {
	case newCollection == nil && old != nil:
		delete(cached, old.Id)
	case newCollection != nil:
		cached[newCollection.Id] = newCollection
	}

	s.App().Store().Set(collectionsStoreKey, cached)
}

func (s *Service) refreshCachedCollections() error {
	if s.App().Dao() == nil {
		return errors.New("app is not initialized yet")
	}

	var collections []*collectionmodels.Collection
	if err := s.App().Dao().CollectionQuery().All(&collections); err != nil {
		return err
	}

	cached := map[string]*collectionmodels.Collection{}
	for _, collection := range collections {
		cached[collection.Id] = collection
	}

	s.App().Store().Set(collectionsStoreKey, cached)
	return nil
}

func (s *Service) getCachedCollections() (map[string]*collectionmodels.Collection, error) {
	if !s.App().Store().Has(collectionsStoreKey) {
		if err := s.refreshCachedCollections(); err != nil {
			return nil, err
		}
	}

	result, _ := s.App().Store().Get(collectionsStoreKey).(map[string]*collectionmodels.Collection)
	if result == nil {
		result = map[string]*collectionmodels.Collection{}
	}

	return result, nil
}
