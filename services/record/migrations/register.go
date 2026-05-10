package migrations

import "github.com/pocketbase/dbx"

// RegisterFunc matches the shared app migration register signature and allows
// the record module to own its migration pack.
type RegisterFunc func(
	up func(dbx.Builder) error,
	down func(dbx.Builder) error,
	optFilename ...string,
)

// Register wires record migrations into the shared app migration list.
func Register(register RegisterFunc) {
	_ = register
}
