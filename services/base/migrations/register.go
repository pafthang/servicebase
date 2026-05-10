package migrations

import "github.com/pocketbase/dbx"

// RegisterFunc matches the shared app migration register signature and allows
// the settings module to contribute its own migration pack without importing
// the root migrations package directly.
type RegisterFunc func(
	up func(dbx.Builder) error,
	down func(dbx.Builder) error,
	optFilename ...string,
)

// Register wires the settings module migrations into the shared migration list.
//
// The settings module currently relies on the shared `_params` settings storage,
// and `models.NewSettings()` already applies defaults for newly introduced
// fields when older serialized settings are loaded. The hook exists so future
// settings-specific migrations can be added in-module without changing the
// registration pattern again.
func Register(register RegisterFunc) {
	_ = register
}
