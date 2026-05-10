package updater

import (
	"github.com/pafthang/servicebase/core"
	servicebase "github.com/pafthang/servicebase/services/base"
	"github.com/spf13/cobra"
)

var Descriptor = servicebase.Descriptor{
	Name:    "updater",
	Purpose: "Infra service for self-update command binding and GitHub release update orchestration.",
	Dependencies: []string{
		"core.App",
		"cobra",
		"github releases API",
	},
	Operations: []string{
		"Bind",
		"CreateCommand",
		"Update",
	},
}

// Service provides self-update command binding and execution.
type Service struct {
	servicebase.Service
	config         Config
	currentVersion string
}

// New creates an updater service bound to the provided app and version.
func New(app core.App, currentVersion string, config Config) *Service {
	return &Service{
		Service:        servicebase.NewWithApp(app),
		config:         config.normalized(),
		currentVersion: currentVersion,
	}
}

// MustRegister is a compatibility helper that binds the updater service and
// panics on error.
func MustRegister(app core.App, rootCmd *cobra.Command, config Config) {
	if err := Register(app, rootCmd, config); err != nil {
		panic(err)
	}
}

// Register is a compatibility helper that binds the updater service.
func Register(app core.App, rootCmd *cobra.Command, config Config) error {
	version := ""
	if rootCmd != nil {
		version = rootCmd.Version
	}

	return New(app, version, config).Bind(rootCmd)
}

// Bind attaches the update command to the provided root command.
func (s *Service) Bind(rootCmd *cobra.Command) error {
	if rootCmd != nil {
		rootCmd.AddCommand(s.CreateCommand())
	}

	return nil
}
