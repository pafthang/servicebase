package migrations

import "github.com/pocketbase/dbx"

func register1640988001ExternalAuthInit(register RegisterFunc) {
	register(func(db dbx.Builder) error {
		_, err := db.NewQuery(`
			CREATE TABLE {{_externalAuths}} (
				[[id]]           TEXT PRIMARY KEY NOT NULL,
				[[collectionId]] TEXT NOT NULL,
				[[recordId]]     TEXT NOT NULL,
				[[provider]]     TEXT NOT NULL,
				[[providerId]]   TEXT NOT NULL,
				[[created]]      TEXT DEFAULT (strftime('%Y-%m-%d %H:%M:%fZ')) NOT NULL,
				[[updated]]      TEXT DEFAULT (strftime('%Y-%m-%d %H:%M:%fZ')) NOT NULL,
				---
				FOREIGN KEY ([[collectionId]]) REFERENCES {{_collections}} ([[id]]) ON UPDATE CASCADE ON DELETE CASCADE
			);

			CREATE UNIQUE INDEX _externalAuths_record_provider_idx on {{_externalAuths}} ([[collectionId]], [[recordId]], [[provider]]);
			CREATE UNIQUE INDEX _externalAuths_collection_provider_idx on {{_externalAuths}} ([[collectionId]], [[provider]], [[providerId]]);
		`).Execute()
		return err
	}, func(db dbx.Builder) error {
		_, err := db.DropTable("_externalAuths").Execute()
		return err
	}, "1640988001_external_auth_init.go")
}
