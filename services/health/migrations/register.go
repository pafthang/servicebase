package migrations

import "github.com/pocketbase/dbx"

type RegisterFunc func(
	up func(dbx.Builder) error,
	down func(dbx.Builder) error,
	optFilename ...string,
)

// Register wires health-owned migrations into the application migration runner.
//
// Health is currently runtime-only, so there are no schema migrations yet. The
// register hook exists to keep the service module shape consistent and to give
// future health-owned persistence a local place to live.
func Register(register RegisterFunc) {}
