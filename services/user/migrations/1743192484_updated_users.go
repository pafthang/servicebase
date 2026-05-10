package migrations

func register1743192484UpdatedUsers(register RegisterFunc) {
	registerCollectionMigration(register, func(app *migrationTx) error {
		collection, err := app.FindCollectionByNameOrId("_pb_users_auth_")
		if err != nil {
			return err
		}

		if err := collection.Fields.AddMarshaledJSONAt(9, []byte(`{
			"autogeneratePattern": "",
			"hidden": false,
			"id": "text760939060",
			"max": 0,
			"min": 0,
			"name": "city",
			"pattern": "",
			"presentable": false,
			"primaryKey": false,
			"required": false,
			"system": false,
			"type": "text"
		}`)); err != nil {
			return err
		}

		if err := collection.Fields.AddMarshaledJSONAt(10, []byte(`{
			"hidden": false,
			"id": "number2499937429",
			"max": null,
			"min": null,
			"name": "lat",
			"onlyInt": false,
			"presentable": false,
			"required": false,
			"system": false,
			"type": "number"
		}`)); err != nil {
			return err
		}

		if err := collection.Fields.AddMarshaledJSONAt(11, []byte(`{
			"hidden": false,
			"id": "number4142125153",
			"max": null,
			"min": null,
			"name": "lon",
			"onlyInt": false,
			"presentable": false,
			"required": false,
			"system": false,
			"type": "number"
		}`)); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app *migrationTx) error {
		collection, err := app.FindCollectionByNameOrId("_pb_users_auth_")
		if err != nil {
			return err
		}

		collection.Fields.RemoveById("text760939060")
		collection.Fields.RemoveById("number2499937429")
		collection.Fields.RemoveById("number4142125153")

		return app.Save(collection)
	}, "1743192484_updated_users.go")
}
