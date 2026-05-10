// Package migrate provides migration command orchestration, migration file
// generation and collection automigration hooks.
package migrate

import (
	"path/filepath"

	"github.com/pafthang/servicebase/core"
	servicebase "github.com/pafthang/servicebase/services/base"
	"github.com/spf13/cobra"
)

var Descriptor = servicebase.Descriptor{
	Name:    "migrate",
	Purpose: "Provides migration command binding, migration file generation and collection automigration hooks.",
	Dependencies: []string{
		"core.App",
		"cobra",
		"dbase migrations registry",
	},
	RuntimeState: []string{
		"cached collection snapshots in app store when automigrate is enabled",
	},
	Operations: []string{
		"Bind",
		"CreateCommand",
		"CreateMigration",
		"CreateCollectionsSnapshot",
	},
}

// Service provides migration command orchestration and automigrate hooks.
type Service struct {
	servicebase.Service
	config Config
}

// New creates a migrate service bound to the provided app and config.
func New(app core.App, config Config) *Service {
	config = config.normalized()
	if config.Dir == "" {
		config.Dir = filepath.Join(app.DataDir(), "../migrations")
	}

	return &Service{
		Service: servicebase.NewWithApp(app),
		config:  config,
	}
}

// MustRegister is a compatibility helper that binds the service to the root
// command and panics on error.
func MustRegister(app core.App, rootCmd *cobra.Command, config Config) {
	if err := Register(app, rootCmd, config); err != nil {
		panic(err)
	}
}

// Register is a compatibility helper that binds the service to the root
// command.
func Register(app core.App, rootCmd *cobra.Command, config Config) error {
	return New(app, config).Bind(rootCmd)
}

// Bind attaches the migrate command to the provided root command and wires
// automigrate hooks when enabled.
func (s *Service) Bind(rootCmd *cobra.Command) error {
	if rootCmd != nil {
		rootCmd.AddCommand(s.CreateCommand())
	}

	if s.config.Automigrate {
		s.App().OnAfterBootstrap().Add(func(e *core.BootstrapEvent) error {
			s.refreshCachedCollections()
			return nil
		})

		s.App().OnBeforeServe().Add(func(e *core.ServeEvent) error {
			s.refreshCachedCollections()
			return nil
		})

		s.App().OnModelAfterCreate().Add(s.afterCollectionChange())
		s.App().OnModelAfterUpdate().Add(s.afterCollectionChange())
		s.App().OnModelAfterDelete().Add(s.afterCollectionChange())
	}

	return nil
}
