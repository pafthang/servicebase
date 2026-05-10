package realtime

import (
	"github.com/pafthang/servicebase/core"
	servicebase "github.com/pafthang/servicebase/services/base"
)

var Descriptor = servicebase.Descriptor{
	Name:    "realtime",
	Purpose: "Provides helpers for realtime subscription forms and auth model synchronization.",
	Dependencies: []string{
		"core.App",
		"services/realtime/forms",
		"subscriptions",
	},
	Operations: []string{
		"NewSubscribeForm",
		"ResolveRecord",
		"ResolveRecordCollection",
		"UpdateClientsAuthModel",
		"UnregisterClientsByAuthModel",
		"ExtractAuthID",
	},
}

type Service struct {
	servicebase.Service
}

type Getter interface {
	Get(string) any
}

func New(app core.App) *Service {
	return &Service{
		Service: servicebase.NewWithApp(app),
	}
}
