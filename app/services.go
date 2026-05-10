package app

import (
	"fmt"

	"github.com/pafthang/servicebase/core"
	backupsvc "github.com/pafthang/servicebase/services/backup"
	cronsvc "github.com/pafthang/servicebase/services/cron"
	logsvc "github.com/pafthang/servicebase/services/log"
	mailsvc "github.com/pafthang/servicebase/services/mails"
	migratesvc "github.com/pafthang/servicebase/services/migrate"
	updatersvc "github.com/pafthang/servicebase/services/updater"
)

type Services struct {
	Backup        *backupsvc.Service
	CronExecutor  *cronsvc.Executor
	CronScheduler *cronsvc.Scheduler
	Cron          *cronsvc.CronService
	Log           *logsvc.Service
	Mail          *mailsvc.MailService
	Migrate       *migratesvc.Service
	Updater       *updatersvc.Service
}

type ServicesConfig struct {
	App core.App

	CurrentVersion string

	SearchCollections []string

	MigrateConfig migratesvc.Config
	UpdaterConfig updatersvc.Config
}

func NewServices(cfg ServicesConfig) (*Services, error) {
	if cfg.App == nil {
		return nil, fmt.Errorf("app is required")
	}

	app := cfg.App

	backupService := backupsvc.New(app)

	cfg.UpdaterConfig.CreateBackup = backupService.CreateBackup

	cronExecutor := cronsvc.NewExecutorWithApp(app)
	cronScheduler := cronsvc.NewScheduler(cronExecutor)
	cronService := cronsvc.NewCronService(cronScheduler, cronExecutor)

	return &Services{
		Backup:        backupService,
		CronExecutor:  cronExecutor,
		CronScheduler: cronScheduler,
		Cron:          cronService,
		Log:           logsvc.New(app),
		Mail:          mailsvc.NewMailServiceWithApp(app),
		Migrate:       migratesvc.New(app, cfg.MigrateConfig),
		Updater:       updatersvc.New(app, cfg.CurrentVersion, cfg.UpdaterConfig),
	}, nil
}
