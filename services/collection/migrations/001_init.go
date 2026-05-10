package migrations

import "github.com/pocketbase/dbx"

func register1640988000CollectionInit(register RegisterFunc) {
	register(func(db dbx.Builder) error {
		_, tablesErr := db.NewQuery(`

			CREATE TABLE {{_collections}} (
				[[id]]         TEXT PRIMARY KEY NOT NULL,
				[[system]]     BOOLEAN DEFAULT FALSE NOT NULL,
				[[type]]       TEXT DEFAULT "base" NOT NULL,
				[[name]]       TEXT UNIQUE NOT NULL,
				[[schema]]     JSON DEFAULT "[]" NOT NULL,
				[[indexes]]    JSON DEFAULT "[]" NOT NULL,
				[[listRule]]   TEXT DEFAULT NULL,
				[[viewRule]]   TEXT DEFAULT NULL,
				[[createRule]] TEXT DEFAULT NULL,
				[[updateRule]] TEXT DEFAULT NULL,
				[[deleteRule]] TEXT DEFAULT NULL,
				[[options]]    JSON DEFAULT "{}" NOT NULL,
				[[created]]    TEXT DEFAULT (strftime('%Y-%m-%d %H:%M:%fZ')) NOT NULL,
				[[updated]]    TEXT DEFAULT (strftime('%Y-%m-%d %H:%M:%fZ')) NOT NULL
			);
		`).Execute()
		if tablesErr != nil {
			return tablesErr
		}

		return nil
	}, func(db dbx.Builder) error {
		tables := []string{

			"_collections",
		}

		for _, name := range tables {
			if _, err := db.DropTable(name).Execute(); err != nil {
				return err
			}
		}

		return nil
	}, "1640988000_collection_init.go")
}
