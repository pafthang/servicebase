package record

import (
	"github.com/pafthang/servicebase/core"
	servicebase "github.com/pafthang/servicebase/services/base"
)

var Descriptor = servicebase.Descriptor{
	Name:    "record",
	Purpose: "Provides record lookup and record form helpers for admin handlers.",
	Dependencies: []string{
		"core.App",
		"services/base/forms",
	},
	Operations: []string{
		"FindByID",
		"NewUpsertForm",
		"Delete",
	},
}

type Service struct {
	servicebase.Service
}

func New(app core.App) *Service {
	return &Service{
		Service: servicebase.NewWithApp(app),
	}
}
