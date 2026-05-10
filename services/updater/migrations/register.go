package migrations

import "github.com/pocketbase/dbx"

type RegisterFunc func(
	up func(dbx.Builder) error,
	down func(dbx.Builder) error,
	optFilename ...string,
)

// Register registers updater-owned migrations.
//
// Updater is an infra/CLI service and currently has no database-owned schema.
func Register(register RegisterFunc) {
	_ = register
}
