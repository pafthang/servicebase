package health

import (
	"github.com/pafthang/servicebase/core"
	"github.com/pafthang/servicebase/services/backup"
	servicebase "github.com/pafthang/servicebase/services/base"
)

var Descriptor = servicebase.Descriptor{
	Name:    "health",
	Purpose: "Exposes lightweight runtime health checks for admin surfaces.",
	Dependencies: []string{
		"core.App",
	},
	Operations: []string{
		"Check",
	},
}

type Service struct {
	servicebase.Service
}

type CheckResult struct {
	CanBackup bool
}

func New(app core.App) *Service {
	return &Service{
		Service: servicebase.NewWithApp(app),
	}
}

func (s *Service) Check() CheckResult {
	return CheckResult{
		CanBackup: !s.App().Store().Has(backup.StoreKeyActiveBackup),
	}
}
