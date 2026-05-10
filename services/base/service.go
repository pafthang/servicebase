package base

import (
	"log/slog"

	"github.com/pafthang/servicebase/core"
	"github.com/pafthang/servicebase/daos"
)

// Descriptor defines a compact, repeatable way to describe a service package.
//
// The intent is architectural clarity rather than runtime behavior. Packages can
// expose a package-level `Descriptor` value to document their role, key
// dependencies and public operations in a uniform format.
type Descriptor struct {
	Name         string
	Purpose      string
	Dependencies []string
	RuntimeState []string
	Operations   []string
}

// Deps captures the common runtime dependencies shared by most services.
type Deps struct {
	App    core.App
	Logger *slog.Logger
}

// Service provides a small reusable base for app-bound services.
//
// Services can embed it to get consistent accessors for App/Dao/Logger without
// introducing a heavy lifecycle framework.
type Service struct {
	app    core.App
	logger *slog.Logger
}

// New creates a base service from the provided dependencies.
func New(deps Deps) Service {
	logger := deps.Logger
	if logger == nil {
		if deps.App != nil && deps.App.Logger() != nil {
			logger = deps.App.Logger()
		} else {
			logger = slog.Default()
		}
	}

	return Service{
		app:    deps.App,
		logger: logger,
	}
}

// NewWithApp creates a base service using the app and its logger.
func NewWithApp(app core.App) Service {
	return New(Deps{App: app})
}

// App returns the bound app instance.
func (s *Service) App() core.App {
	return s.app
}

// Dao returns the primary dao instance, or nil when the app is unset.
func (s *Service) Dao() *daos.Dao {
	if s.app == nil {
		return nil
	}

	return s.app.Dao()
}

// Logger returns the bound logger.
func (s *Service) Logger() *slog.Logger {
	return s.logger
}

// SetApp rebinds the service to a different app and refreshes the logger if it
// wasn't explicitly overridden.
func (s *Service) SetApp(app core.App) {
	s.app = app

	if app != nil && app.Logger() != nil {
		s.logger = app.Logger()
	} else if s.logger == nil {
		s.logger = slog.Default()
	}
}
