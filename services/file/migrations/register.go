package migrations

import "github.com/pocketbase/dbx"

type RegisterFunc func(
	up func(dbx.Builder) error,
	down func(dbx.Builder) error,
	optFilename ...string,
)

func Register(register RegisterFunc) {
	register001Init(register)
}
