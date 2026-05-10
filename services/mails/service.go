package mails

import (
	"github.com/pafthang/servicebase/core"
)

// Service handles transactional emails.
type Service struct {
	app core.App
}

// New creates new mails service instance.
func New(app core.App) *Service {
	return &Service{
		app: app,
	}
}
