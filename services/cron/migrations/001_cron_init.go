package migrations

import "github.com/pocketbase/dbx"

func register001CronInit(register RegisterFunc) {
	register(func(db dbx.Builder) error {
		_, err := db.NewQuery(`
			CREATE TABLE IF NOT EXISTS crons (
				id TEXT PRIMARY KEY NOT NULL,
				name TEXT NOT NULL DEFAULT '',
				schedule TEXT NOT NULL DEFAULT '',
				command TEXT NOT NULL DEFAULT '',
				is_system BOOLEAN NOT NULL DEFAULT FALSE,
				is_active BOOLEAN NOT NULL DEFAULT TRUE,
				last_run TEXT NOT NULL DEFAULT '',
				next_run TEXT NOT NULL DEFAULT '',
				last_error TEXT NOT NULL DEFAULT '',
				created TEXT NOT NULL DEFAULT '',
				updated TEXT NOT NULL DEFAULT ''
			);

			CREATE INDEX IF NOT EXISTS idx_crons_active ON crons (is_active);
			CREATE INDEX IF NOT EXISTS idx_crons_system ON crons (is_system);
		`).Execute()
		return err
	}, func(db dbx.Builder) error {
		_, err := db.DropTable("crons").Execute()
		return err
	}, "001_cron_init.go")
}
