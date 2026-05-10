package migrations

import (
	"github.com/pafthang/servicebase/daos"
	collectionmodels "github.com/pafthang/servicebase/services/collection/models"
	"github.com/pafthang/servicebase/tools/types"
	"github.com/pocketbase/dbx"
)

func register1640988000UsersAuthInit(register RegisterFunc) {
	register(func(db dbx.Builder) error {
		collection := &collectionmodels.Collection{
			BaseModel: collectionmodels.BaseModel{
				Id: "_pb_users_auth_",
			},
			Name:   "users",
			Type:   collectionmodels.CollectionTypeUsers,
			System: false,
			Schema: collectionmodels.NewSchema(),
			Options: types.JsonMap{
				"allowEmailAuth":    true,
				"allowOAuth2Auth":   true,
				"allowUsernameAuth": true,
				"requireEmail":      false,
				"minPasswordLength": 8,
			},
		}

		return daos.New(db).SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db)

		collection, err := dao.FindCollectionByNameOrId("_pb_users_auth_")
		if err != nil {
			return err
		}

		return dao.DeleteCollection(collection)
	}, "1640988000_users_auth_init.go")
}
