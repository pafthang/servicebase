package registry

import (
	"github.com/pafthang/servicebase/tools/migrate"

	collectionmigrations "github.com/pafthang/servicebase/services/collection/migrations"
	cronmigrations "github.com/pafthang/servicebase/services/cron/migrations"
	filemigrations "github.com/pafthang/servicebase/services/file/migrations"
	logmigrations "github.com/pafthang/servicebase/services/log/migrations"
	settingsmigrations "github.com/pafthang/servicebase/services/settings/migrations"
	usermigrations "github.com/pafthang/servicebase/services/user/migrations"
)

var AppMigrations migrate.MigrationsList
var LogsMigrations migrate.MigrationsList

func init() {
	register := AppMigrations.Register

	collectionmigrations.Register(register)

	cronmigrations.Register(register)
	filemigrations.Register(register)
	settingsmigrations.Register(register)
	usermigrations.Register(register)

	logmigrations.Register(LogsMigrations.Register)
}
