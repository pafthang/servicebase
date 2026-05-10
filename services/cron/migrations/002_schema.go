package migrations

import "github.com/pocketbase/dbx"

type tableColumn struct {
	Name string `db:"name"`
}

func register002Schema(register RegisterFunc) {
	register(func(db dbx.Builder) error {
		columns := map[string]string{
			"user":                 `TEXT NOT NULL DEFAULT ''`,
			"description":          `TEXT NOT NULL DEFAULT ''`,
			"webhook_url":          `TEXT NOT NULL DEFAULT ''`,
			"webhook_method":       `TEXT NOT NULL DEFAULT 'POST'`,
			"webhook_headers":      `JSON DEFAULT '{}'`,
			"webhook_payload":      `JSON DEFAULT '{}'`,
			"is_system":            `BOOLEAN NOT NULL DEFAULT FALSE`,
			"system_type":          `TEXT NOT NULL DEFAULT ''`,
			"last_run":             `TEXT NOT NULL DEFAULT ''`,
			"next_run":             `TEXT NOT NULL DEFAULT ''`,
			"timeout_seconds":      `INTEGER NOT NULL DEFAULT 30`,
			"notify_on_success":    `BOOLEAN NOT NULL DEFAULT FALSE`,
			"notify_on_failure":    `BOOLEAN NOT NULL DEFAULT FALSE`,
			"notification_webhook": `TEXT NOT NULL DEFAULT ''`,
			"max_retries":          `INTEGER NOT NULL DEFAULT 0`,
			"retry_delay_seconds":  `INTEGER NOT NULL DEFAULT 60`,
		}

		for name, definition := range columns {
			exists, err := columnExists(db, "crons", name)
			if err != nil {
				return err
			}
			if exists {
				continue
			}
			if _, err := db.NewQuery("ALTER TABLE crons ADD COLUMN " + name + " " + definition).Execute(); err != nil {
				return err
			}
		}

		statements := []string{
			`CREATE INDEX IF NOT EXISTS idx_crons_user ON crons (user)`,
			`CREATE INDEX IF NOT EXISTS idx_crons_active ON crons (is_active)`,
			`CREATE TABLE IF NOT EXISTS cron_executions (
				id TEXT PRIMARY KEY NOT NULL,
				cron TEXT NOT NULL DEFAULT '',
				user TEXT NOT NULL DEFAULT '',
				status TEXT NOT NULL DEFAULT '',
				started_at TEXT NOT NULL DEFAULT '',
				completed_at TEXT NOT NULL DEFAULT '',
				duration_ms INTEGER NOT NULL DEFAULT 0,
				http_status INTEGER NOT NULL DEFAULT 0,
				request_url TEXT NOT NULL DEFAULT '',
				request_method TEXT NOT NULL DEFAULT '',
				request_headers JSON DEFAULT '{}',
				request_payload JSON DEFAULT '{}',
				response_headers JSON DEFAULT '{}',
				response_body TEXT NOT NULL DEFAULT '',
				error_message TEXT NOT NULL DEFAULT '',
				error_stack TEXT NOT NULL DEFAULT '',
				retry_count INTEGER NOT NULL DEFAULT 0,
				is_retry BOOLEAN NOT NULL DEFAULT FALSE,
				metadata JSON DEFAULT '{}',
				created TEXT NOT NULL DEFAULT '',
				updated TEXT NOT NULL DEFAULT ''
			)`,
			`CREATE INDEX IF NOT EXISTS idx_cron_executions_cron_created ON cron_executions (cron, created)`,
			`CREATE INDEX IF NOT EXISTS idx_cron_executions_user_created ON cron_executions (user, created)`,
		}
		for _, stmt := range statements {
			if _, err := db.NewQuery(stmt).Execute(); err != nil {
				return err
			}
		}
		return nil
	}, func(db dbx.Builder) error {
		_, err := db.DropTable("cron_executions").Execute()
		return err
	}, "1746810100_cron_schema.go")
}

func columnExists(db dbx.Builder, tableName, columnName string) (bool, error) {
	var columns []tableColumn
	if err := db.NewQuery("PRAGMA table_info(" + tableName + ")").All(&columns); err != nil {
		return false, err
	}
	for _, column := range columns {
		if column.Name == columnName {
			return true, nil
		}
	}
	return false, nil
}
