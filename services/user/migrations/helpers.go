package migrations

import (
	"encoding/json"

	collectionmodels "github.com/pafthang/servicebase/services/collection/models"
	"github.com/pocketbase/dbx"

	"github.com/pafthang/servicebase/daos"
)

func registerCollectionMigration(
	register RegisterFunc,
	up func(tx *migrationTx) error,
	down func(tx *migrationTx) error,
	filename string,
) {
	register(func(db dbx.Builder) error {
		return up(newMigrationTx(db))
	}, func(db dbx.Builder) error {
		return down(newMigrationTx(db))
	}, filename)
}

type migrationTx struct {
	dao *daos.Dao
}

func newMigrationTx(db dbx.Builder) *migrationTx {
	return &migrationTx{dao: daos.New(db)}
}

func (tx *migrationTx) FindCollectionByNameOrId(nameOrID string) (*migrationCollection, error) {
	collection, err := tx.dao.FindCollectionByNameOrId(nameOrID)
	if err != nil {
		return nil, err
	}

	return newMigrationCollection(collection), nil
}

func (tx *migrationTx) Save(collection *migrationCollection) error {
	if collection == nil || collection.Collection == nil {
		return nil
	}

	return tx.dao.SaveCollection(collection.Collection)
}

type migrationCollection struct {
	*collectionmodels.Collection
	Fields *migrationFields `json:"-"`
}

func newMigrationCollection(collection *collectionmodels.Collection) *migrationCollection {
	result := &migrationCollection{Collection: collection}
	result.syncFields()
	return result
}

func (c *migrationCollection) syncFields() {
	c.Fields = &migrationFields{collection: c}
}

type migrationFields struct {
	collection *migrationCollection
}

func (f *migrationFields) RemoveById(id string) {
	if f == nil || f.collection == nil || f.collection.Collection == nil {
		return
	}

	f.collection.Schema.RemoveField(id)
}

func (f *migrationFields) AddMarshaledJSONAt(index int, data []byte) error {
	if f == nil || f.collection == nil || f.collection.Collection == nil {
		return nil
	}

	field := &collectionmodels.SchemaField{}
	if err := json.Unmarshal(data, field); err != nil {
		return err
	}

	if err := field.InitOptions(); err != nil {
		return err
	}

	fields := append([]*collectionmodels.SchemaField(nil), f.collection.Schema.Fields()...)
	if index < 0 {
		index = 0
	}
	if index > len(fields) {
		index = len(fields)
	}

	reordered := make([]*collectionmodels.SchemaField, 0, len(fields)+1)
	reordered = append(reordered, fields[:index]...)
	reordered = append(reordered, field)
	reordered = append(reordered, fields[index:]...)

	f.collection.Schema = collectionmodels.NewSchema(reordered...)
	return nil
}
