package migrations

import "github.com/pocketbase/dbx"

// RegisterFunc matches the shared app migration register signature and allows
// the backup module to contribute migrations without importing the root
// migrations package directly.
type RegisterFunc func(
	up func(dbx.Builder) error,
	down func(dbx.Builder) error,
	optFilename ...string,
)

// Register wires backup module migrations into the shared migration list.
//
// No migrations are registered for now because backup archives live in the
// configured backups filesystem and there is no backup-owned DB schema yet.
func Register(register RegisterFunc) {}
