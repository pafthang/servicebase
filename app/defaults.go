package app

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pafthang/servicebase/core"
	"github.com/pafthang/servicebase/services/updater"
	"github.com/spf13/cobra"
)

// TinybaseDefaultsConfig contains optional defaults for the application CLI/runtime wiring.
type TinybaseDefaultsConfig struct {
	PublicDir       string
	IndexFallback   bool
	QueryTimeoutSec int

	// HideStartBanner disables the default serve command banner.
	HideStartBanner bool
}

// RegisterTinybaseDefaults wires the default Tinybase CLI flags, commands,
// updater, query timeout and static file serving into the app.
func RegisterTinybaseDefaults(app core.App, rootCmd *cobra.Command, cfg TinybaseDefaultsConfig) {
	if app == nil || rootCmd == nil {
		return
	}

	if cfg.PublicDir == "" {
		cfg.PublicDir = defaultPublicDir()
	}
	if cfg.QueryTimeoutSec <= 0 {
		cfg.QueryTimeoutSec = 30
	}

	var hooksDir string
	rootCmd.PersistentFlags().StringVar(
		&hooksDir,
		"hooksDir",
		"",
		"the directory with the JS app hooks",
	)

	var hooksWatch bool
	rootCmd.PersistentFlags().BoolVar(
		&hooksWatch,
		"hooksWatch",
		true,
		"auto restart the app on pb_hooks file change",
	)

	var hooksPool int
	rootCmd.PersistentFlags().IntVar(
		&hooksPool,
		"hooksPool",
		25,
		"the total prewarm goja.Runtime instances for the JS app hooks execution",
	)

	var migrationsDir string
	rootCmd.PersistentFlags().StringVar(
		&migrationsDir,
		"migrationsDir",
		"",
		"the directory with the user defined migrations",
	)

	var automigrate bool
	rootCmd.PersistentFlags().BoolVar(
		&automigrate,
		"automigrate",
		true,
		"enable/disable auto migrations",
	)

	publicDir := cfg.PublicDir
	rootCmd.PersistentFlags().StringVar(
		&publicDir,
		"publicDir",
		cfg.PublicDir,
		"the directory to serve static files",
	)

	indexFallback := true
	if cfg.IndexFallback {
		indexFallback = cfg.IndexFallback
	}
	rootCmd.PersistentFlags().BoolVar(
		&indexFallback,
		"indexFallback",
		indexFallback,
		"fallback the request to index.html on missing static path (eg. when pretty urls are used with SPA)",
	)

	queryTimeout := cfg.QueryTimeoutSec
	rootCmd.PersistentFlags().IntVar(
		&queryTimeout,
		"queryTimeout",
		cfg.QueryTimeoutSec,
		"the default SELECT queries timeout in seconds",
	)

	_ = rootCmd.ParseFlags(os.Args[1:])

	updater.MustRegister(app, rootCmd, updater.Config{})

	app.OnAfterBootstrap().PreAdd(func(e *core.BootstrapEvent) error {
		app.Dao().ModelQueryTimeout = time.Duration(queryTimeout) * time.Second
		return nil
	})

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.GET("/*", StaticDirectoryHandler(os.DirFS(publicDir), indexFallback))
		return nil
	})

	rootCmd.AddCommand(NewServeCommand(app, !cfg.HideStartBanner))
}

func defaultPublicDir() string {
	if strings.HasPrefix(os.Args[0], os.TempDir()) {
		return "./pb_public"
	}

	return filepath.Join(os.Args[0], "../pb_public")
}
