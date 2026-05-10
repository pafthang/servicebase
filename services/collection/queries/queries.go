package queries

import (
	"database/sql"

	collectionmodels "github.com/pafthang/servicebase/services/collection/models"

	"github.com/pafthang/servicebase/daos"
	teamservice "github.com/pafthang/servicebase/services/team"
	"github.com/pafthang/servicebase/tools/search"
	"github.com/pocketbase/dbx"
)

type ListResult = search.Result

func List(dao *daos.Dao, query string) (*ListResult, []*collectionmodels.Collection, error) {
	fieldResolver := search.NewSimpleFieldResolver(
		"id", "created", "updated", "name", "system", "type",
	)

	collections := []*collectionmodels.Collection{}

	baseQuery := dao.CollectionQuery().
		AndWhere(dbx.NotIn("name",
			teamservice.TeamsCollectionName,
			teamservice.TeamMembersCollectionName,
		))

	result, err := search.NewProvider(fieldResolver).
		Query(baseQuery).
		ParseAndExec(query, &collections)

	return result, collections, err
}

func FindByNameOrID(dao *daos.Dao, nameOrID string) (*collectionmodels.Collection, error) {
	collection, err := dao.FindCollectionByNameOrId(nameOrID)
	if err != nil || collection == nil {
		return collection, err
	}

	if collection.Name == teamservice.TeamsCollectionName || collection.Name == teamservice.TeamMembersCollectionName {
		return nil, sql.ErrNoRows
	}

	return collection, nil
}

func Delete(dao *daos.Dao, collection *collectionmodels.Collection) error {
	return dao.DeleteCollection(collection)
}
