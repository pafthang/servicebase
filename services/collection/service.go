package collection

import (
	"github.com/pafthang/servicebase/core"
	servicebase "github.com/pafthang/servicebase/services/base"
)

var Descriptor = servicebase.Descriptor{
	Name:    "collection",
	Purpose: "Provides collection listing, lookup and mutation helpers for the admin API.",
	Dependencies: []string{
		"core.App",
		"services/base/forms",
		"team service collection exclusions",
	},
	Operations: []string{
		"List",
		"FindByNameOrID",
		"NewUpsertForm",
		"SubmitUpsert",
		"Delete",
		"NewImportForm",
		"SubmitImport",
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
