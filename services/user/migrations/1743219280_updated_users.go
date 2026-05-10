package migrations

import "encoding/json"

func register1743219280UpdatedUsers(register RegisterFunc) {
	registerCollectionMigration(register, func(app *migrationTx) error {
		collection, err := app.FindCollectionByNameOrId("_pb_users_auth_")
		if err != nil {
			return err
		}

		if err := json.Unmarshal([]byte(`{
			"authRule": "verified = true",
			"createRule": "@request.auth.id = id",
			"deleteRule": "@request.auth.id = id",
			"listRule": "@request.auth.id = id",
			"updateRule": "@request.auth.id = id && \n(@request.body.id:isset = false || @request.auth.id = @request.body.id)",
			"viewRule": "@request.auth.id = id"
		}`), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app *migrationTx) error {
		collection, err := app.FindCollectionByNameOrId("_pb_users_auth_")
		if err != nil {
			return err
		}

		if err := json.Unmarshal([]byte(`{
			"authRule": "",
			"createRule": null,
			"deleteRule": null,
			"listRule": null,
			"updateRule": null,
			"viewRule": null
		}`), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	}, "1743219280_updated_users.go")
}
