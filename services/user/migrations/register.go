package migrations

import "github.com/pocketbase/dbx"

// RegisterFunc matches the shared app migration register signature and allows
// the user module to own its migration pack.
type RegisterFunc func(
	up func(dbx.Builder) error,
	down func(dbx.Builder) error,
	optFilename ...string,
)

// Register wires user module migrations into the shared app migration list.
func Register(register RegisterFunc) {
	register1640988000UsersAuthInit(register)
	register1640988001ExternalAuthInit(register)
	register1743192484UpdatedUsers(register)
	register1743219280UpdatedUsers(register)
}
