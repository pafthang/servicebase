package app

import (
	"fmt"

	"github.com/pafthang/servicebase/core"
	"github.com/pafthang/servicebase/services/migrate/registry"
	"github.com/pafthang/servicebase/tools/migrate"
	"github.com/pocketbase/dbx"
)

type migrationsConnection struct {
	DB             *dbx.DB
	MigrationsList migrate.MigrationsList
}

func RunMigrations(app core.App) error {
	if app == nil || app.Dao() == nil || app.LogsDao() == nil {
		return fmt.Errorf("app is not bootstrapped")
	}

	dataDB, ok := app.Dao().DB().(*dbx.DB)
	if !ok || dataDB == nil {
		return fmt.Errorf("failed to resolve app db")
	}

	logsDB, ok := app.LogsDao().DB().(*dbx.DB)
	if !ok || logsDB == nil {
		return fmt.Errorf("failed to resolve logs db")
	}

	connections := []migrationsConnection{
		{DB: dataDB, MigrationsList: registry.AppMigrations},
		{DB: logsDB, MigrationsList: registry.LogsMigrations},
	}

	for _, c := range connections {
		runner, err := migrate.NewRunner(c.DB, c.MigrationsList)
		if err != nil {
			return err
		}
		if _, err := runner.Up(); err != nil {
			return err
		}
	}

	return nil
}
