package migrations

import "github.com/pocketbase/dbx"

func register002AppLogs(register RegisterFunc) {
	register(func(db dbx.Builder) error {
		statements := []string{
			`CREATE TABLE IF NOT EXISTS logging_projects (
				id TEXT PRIMARY KEY NOT NULL,
				name TEXT NOT NULL DEFAULT '',
				slug TEXT NOT NULL DEFAULT '',
				dev_token TEXT NOT NULL DEFAULT '',
				retention INTEGER NOT NULL DEFAULT 30,
				active BOOLEAN NOT NULL DEFAULT TRUE,
				created TEXT NOT NULL DEFAULT '',
				updated TEXT NOT NULL DEFAULT ''
			)`,
			`CREATE UNIQUE INDEX IF NOT EXISTS idx_logging_projects_slug ON logging_projects (slug)`,
			`CREATE INDEX IF NOT EXISTS idx_logging_projects_dev_token ON logging_projects (dev_token)`,

			`CREATE TABLE IF NOT EXISTS logs (
				id TEXT PRIMARY KEY NOT NULL,
				project TEXT NOT NULL DEFAULT '',
				level TEXT NOT NULL DEFAULT 'info',
				timestamp TEXT NOT NULL DEFAULT '',
				source TEXT NOT NULL DEFAULT '',
				message TEXT NOT NULL DEFAULT '',
				context JSON DEFAULT '{}',
				trace_id TEXT NOT NULL DEFAULT '',
				created TEXT NOT NULL DEFAULT '',
				updated TEXT NOT NULL DEFAULT ''
			)`,
			`CREATE INDEX IF NOT EXISTS idx_logs_project_timestamp ON logs (project, timestamp)`,
			`CREATE INDEX IF NOT EXISTS idx_logs_level ON logs (level)`,
			`CREATE INDEX IF NOT EXISTS idx_logs_trace_id ON logs (trace_id)`,
		}

		for _, stmt := range statements {
			if _, err := db.NewQuery(stmt).Execute(); err != nil {
				return err
			}
		}

		return nil
	}, func(db dbx.Builder) error {
		for _, stmt := range []string{
			`DROP TABLE IF EXISTS logs`,
			`DROP TABLE IF EXISTS logging_projects`,
		} {
			if _, err := db.NewQuery(stmt).Execute(); err != nil {
				return err
			}
		}

		return nil
	}, "002_log_app_logs.go")
}
