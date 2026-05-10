package migrations

import "github.com/pocketbase/dbx"

func register001SystemLogs(register RegisterFunc) {
	register(func(db dbx.Builder) error {
		_, err := db.NewQuery(`
			CREATE TABLE IF NOT EXISTS {{_logs}} (
				[[id]]      TEXT PRIMARY KEY DEFAULT ('r'||lower(hex(randomblob(7)))) NOT NULL,
				[[level]]   INTEGER DEFAULT 0 NOT NULL,
				[[message]] TEXT DEFAULT "" NOT NULL,
				[[data]]    JSON DEFAULT "{}" NOT NULL,
				[[created]] TEXT DEFAULT (strftime('%Y-%m-%d %H:%M:%fZ')) NOT NULL,
				[[updated]] TEXT DEFAULT (strftime('%Y-%m-%d %H:%M:%fZ')) NOT NULL
			);

			CREATE INDEX IF NOT EXISTS _logs_level_idx ON {{_logs}} ([[level]]);
			CREATE INDEX IF NOT EXISTS _logs_message_idx ON {{_logs}} ([[message]]);
			CREATE INDEX IF NOT EXISTS _logs_created_hour_idx ON {{_logs}} (strftime('%Y-%m-%d %H:00:00', [[created]]));
		`).Execute()

		return err
	}, func(db dbx.Builder) error {
		_, err := db.DropTable("_logs").Execute()
		return err
	}, "001_log_system_logs.go")
}
