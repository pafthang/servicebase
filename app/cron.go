package app

import (
	"fmt"

	"github.com/pafthang/servicebase/core"
	cronservice "github.com/pafthang/servicebase/services/cron"
	cronmodels "github.com/pafthang/servicebase/services/cron/models"
	logsvc "github.com/pafthang/servicebase/services/log"
	toolscron "github.com/pafthang/servicebase/tools/cron"
)

type CronConfig struct {
	App             core.App
	CronScheduler   *cronservice.Scheduler
	CronService     *cronservice.CronService
	EnableStatsJobs bool
	EnableLogJobs   bool
}

func RegisterCrons(cfg CronConfig) error {
	if cfg.App == nil {
		return fmt.Errorf("app is required")
	}

	if cfg.CronScheduler != nil {
		bindDynamicCronScheduler(cfg.App, cfg.CronScheduler)
		registerDynamicCronHooks(cfg.App, cfg.CronScheduler)
	}

	registerSystemCronRuntime(cfg)

	return nil
}

func bindDynamicCronScheduler(app core.App, scheduler *cronservice.Scheduler) {
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		return scheduler.Start()
	})

	app.OnTerminate().Add(func(e *core.TerminateEvent) error {
		scheduler.Stop()
		return nil
	})
}

func registerDynamicCronHooks(app core.App, scheduler *cronservice.Scheduler) {
	app.OnModelAfterCreate("crons").Add(func(e *core.ModelEvent) error {
		item, ok := e.Model.(*cronmodels.Cron)
		if !ok || item.IsSystem || !item.IsActive {
			return nil
		}

		return scheduler.AddCron(item)
	})

	app.OnModelAfterUpdate("crons").Add(func(e *core.ModelEvent) error {
		item, ok := e.Model.(*cronmodels.Cron)
		if !ok || item.IsSystem {
			return nil
		}

		if item.IsActive {
			return scheduler.AddCron(item)
		}

		return scheduler.RemoveCron(item.Id)
	})

	app.OnModelAfterDelete("crons").Add(func(e *core.ModelEvent) error {
		item, ok := e.Model.(*cronmodels.Cron)
		if !ok {
			return nil
		}

		return scheduler.RemoveCron(item.Id)
	})
}

func registerSystemCronRuntime(cfg CronConfig) {
	systemCron := toolscron.New()

	if cfg.EnableLogJobs {
		_ = systemCron.Add("logging-retention-cleanup", "0 2 * * *", func() {
			_ = logsvc.CleanupLogs(cfg.App)
		})
	}

	cfg.App.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		systemCron.Start()
		return nil
	})

	cfg.App.OnTerminate().Add(func(e *core.TerminateEvent) error {
		systemCron.Stop()
		return nil
	})
}
