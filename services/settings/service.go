package settings

import (
	"github.com/pafthang/servicebase/core"
	servicebase "github.com/pafthang/servicebase/services/base"
)

var Descriptor = servicebase.Descriptor{
	Name:    "settings",
	Purpose: "Wraps settings reads and settings-related admin forms.",
	Dependencies: []string{
		"core.App",
		"services/base/forms",
	},
	Operations: []string{
		"RedactClone",
		"NewUpsertForm",
		"SubmitUpsert",
		"NewTestS3Form",
		"TestS3",
		"NewTestEmailForm",
		"TestEmail",
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
