package cron

import (
	"fmt"

	"github.com/pafthang/servicebase/core"
	cronmodels "github.com/pafthang/servicebase/services/cron/models"
	"github.com/pocketbase/dbx"
)

func countCrons(app core.App, filter map[string]any) (int64, error) {
	if app == nil {
		return 0, fmt.Errorf("app is required")
	}

	var total int64
	err := app.Dao().
		ModelQuery(&cronmodels.Cron{}).
		Where(dbx.HashExp(filter)).
		Select("count(*)").
		Row(&total)

	return total, err
}

func findCronByID(app core.App, id string) (*cronmodels.Cron, error) {
	if app == nil {
		return nil, fmt.Errorf("app is required")
	}

	item := &cronmodels.Cron{}
	if err := app.Dao().FindById(item, id); err != nil {
		return nil, err
	}

	return item, nil
}

func findCronByFilter(app core.App, filter map[string]any) (*cronmodels.Cron, error) {
	if app == nil {
		return nil, fmt.Errorf("app is required")
	}

	item := &cronmodels.Cron{}
	err := app.Dao().
		ModelQuery(item).
		Where(dbx.HashExp(filter)).
		Limit(1).
		One(item)
	if err != nil {
		return nil, err
	}

	return item, nil
}

func listCrons(app core.App, filter map[string]any) ([]*cronmodels.Cron, error) {
	if app == nil {
		return nil, fmt.Errorf("app is required")
	}

	var items []*cronmodels.Cron
	err := app.Dao().
		ModelQuery(&cronmodels.Cron{}).
		Where(dbx.HashExp(filter)).
		All(&items)

	return items, err
}

func saveCron(app core.App, item *cronmodels.Cron) error {
	if app == nil {
		return fmt.Errorf("app is required")
	}

	return app.Dao().Save(item)
}

func deleteCron(app core.App, item *cronmodels.Cron) error {
	if app == nil {
		return fmt.Errorf("app is required")
	}

	return app.Dao().Delete(item)
}

func saveCronExecution(app core.App, item *cronmodels.CronExecution) error {
	if app == nil {
		return fmt.Errorf("app is required")
	}

	return app.Dao().Save(item)
}
