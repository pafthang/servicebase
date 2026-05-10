package log

import (
	"fmt"

	"github.com/pafthang/servicebase/core"
	"github.com/pafthang/servicebase/daos"
	logmodels "github.com/pafthang/servicebase/services/log/models"
)

func saveAppLog(app core.App, item *logmodels.AppLog) error {
	if app == nil {
		return fmt.Errorf("app is required")
	}

	return app.Dao().Save(item)
}

func runInTransaction(app core.App, fn func(txApp core.App) error) error {
	if app == nil {
		return fmt.Errorf("app is required")
	}

	return app.Dao().RunInTransaction(func(txDao *daos.Dao) error {
		return fn(&txApp{App: app, dao: txDao})
	})
}

type txApp struct {
	core.App
	dao *daos.Dao
}

func (a *txApp) Dao() *daos.Dao {
	return a.dao
}
