package migrations

import "github.com/pocketbase/dbx"

func register001Init(register RegisterFunc) {
	register(func(db dbx.Builder) error {
		statements := []string{
			`CREATE TABLE IF NOT EXISTS files (id TEXT PRIMARY KEY NOT NULL, user TEXT NOT NULL DEFAULT '', filename TEXT NOT NULL DEFAULT '', original_filename TEXT NOT NULL DEFAULT '', file TEXT NOT NULL DEFAULT '', mime_type TEXT NOT NULL DEFAULT '', size INTEGER NOT NULL DEFAULT 0, path TEXT NOT NULL DEFAULT '', folder TEXT NOT NULL DEFAULT '', description TEXT NOT NULL DEFAULT '', tags JSON DEFAULT '[]', created TEXT NOT NULL DEFAULT '', updated TEXT NOT NULL DEFAULT '')`,
			`CREATE INDEX IF NOT EXISTS idx_files_user ON files (user)`,
			`CREATE INDEX IF NOT EXISTS idx_files_folder ON files (folder)`,
			`CREATE TABLE IF NOT EXISTS folders (id TEXT PRIMARY KEY NOT NULL, user TEXT NOT NULL DEFAULT '', name TEXT NOT NULL DEFAULT '', parent TEXT NOT NULL DEFAULT '', color TEXT NOT NULL DEFAULT '', created TEXT NOT NULL DEFAULT '', updated TEXT NOT NULL DEFAULT '')`,
			`CREATE INDEX IF NOT EXISTS idx_folders_user ON folders (user)`,
			`CREATE INDEX IF NOT EXISTS idx_folders_parent ON folders (parent)`,
			`CREATE TABLE IF NOT EXISTS user_storage_quotas (id TEXT PRIMARY KEY NOT NULL, user TEXT NOT NULL DEFAULT '', quota_bytes INTEGER NOT NULL DEFAULT 0, used_bytes INTEGER NOT NULL DEFAULT 0, created TEXT NOT NULL DEFAULT '', updated TEXT NOT NULL DEFAULT '')`,
			`CREATE UNIQUE INDEX IF NOT EXISTS idx_user_storage_quotas_user ON user_storage_quotas (user)`,
		}
		for _, stmt := range statements {
			if _, err := db.NewQuery(stmt).Execute(); err != nil {
				return err
			}
		}
		return nil
	}, func(db dbx.Builder) error {
		statements := []string{
			`DROP TABLE IF EXISTS user_storage_quotas`,
			`DROP TABLE IF EXISTS folders`,
			`DROP TABLE IF EXISTS files`,
		}
		for _, stmt := range statements {
			if _, err := db.NewQuery(stmt).Execute(); err != nil {
				return err
			}
		}
		return nil
	}, "001_file_init.go")
}
