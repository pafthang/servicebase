// Package migrations contains backup-owned schema migrations.
//
// Backups are currently stored by the configured backups filesystem, so this
// service doesn't own database tables yet. The package still exposes a Register
// hook to keep the service migration boundary explicit and ready for future
// backup metadata/audit tables.
package migrations
