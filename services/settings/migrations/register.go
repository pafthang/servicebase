package migrations

import "github.com/pocketbase/dbx"

// RegisterFunc matches the shared app migration register signature and allows
// the settings module to contribute migrations without importing the root
// migrations package directly.
type RegisterFunc func(
	up func(dbx.Builder) error,
	down func(dbx.Builder) error,
	optFilename ...string,
)

// Register wires settings module migrations into the shared migration list.
func Register(register RegisterFunc) {
	register001Init(register)
}
