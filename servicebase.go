package servicebase

import (
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/fatih/color"
	"github.com/pafthang/servicebase/core"
	"github.com/pafthang/servicebase/tools/list"
	"github.com/spf13/cobra"
)

var _ core.App = (*ServiceBase)(nil)

// Version of ServiceBase.
var Version = "(untracked)"

// appWrapper serves as a private App instance wrapper.
type appWrapper struct {
	core.App
}

// ServiceBase defines a ServiceBase app launcher.
//
// It implements core.App via embedding and all of the app interface methods
// could be accessed directly through the instance.
type ServiceBase struct {
	*appWrapper

	devFlag           bool
	dataDirFlag       string
	encryptionEnvFlag string
	hideStartBanner   bool

	// RootCmd is the main console command.
	RootCmd *cobra.Command
}

// Config is the ServiceBase initialization config struct.
type Config struct {
	// optional default values for the console flags
	DefaultDev           bool
	DefaultDataDir       string
	DefaultEncryptionEnv string

	// hide the default console server info on app startup
	HideStartBanner bool

	// optional DB configurations
	DataMaxOpenConns int
	DataMaxIdleConns int
	LogsMaxOpenConns int
	LogsMaxIdleConns int
}

// New creates a new ServiceBase instance with the default configuration.
func New() *ServiceBase {
	_, isUsingGoRun := inspectRuntime()

	return NewWithConfig(Config{
		DefaultDev: isUsingGoRun,
	})
}

// NewWithConfig creates a new ServiceBase instance with the provided config.
func NewWithConfig(config Config) *ServiceBase {
	if config.DefaultDataDir == "" {
		baseDir, _ := inspectRuntime()
		config.DefaultDataDir = filepath.Join(baseDir, "data")
	}

	pb := &ServiceBase{
		RootCmd: &cobra.Command{
			Use:     filepath.Base(os.Args[0]),
			Short:   "ServiceBase CLI",
			Version: Version,
			FParseErrWhitelist: cobra.FParseErrWhitelist{
				UnknownFlags: true,
			},
			CompletionOptions: cobra.CompletionOptions{
				DisableDefaultCmd: true,
			},
		},
		devFlag:           config.DefaultDev,
		dataDirFlag:       config.DefaultDataDir,
		encryptionEnvFlag: config.DefaultEncryptionEnv,
		hideStartBanner:   config.HideStartBanner,
	}

	pb.RootCmd.SetErr(newErrWriter())
	_ = pb.eagerParseFlags(&config)

	pb.appWrapper = &appWrapper{core.NewBaseApp(core.BaseAppConfig{
		IsDev:            pb.devFlag,
		DataDir:          pb.dataDirFlag,
		EncryptionEnv:    pb.encryptionEnvFlag,
		DataMaxOpenConns: config.DataMaxOpenConns,
		DataMaxIdleConns: config.DataMaxIdleConns,
		LogsMaxOpenConns: config.LogsMaxOpenConns,
		LogsMaxIdleConns: config.LogsMaxIdleConns,
	})}

	pb.RootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	return pb
}

// ShowStartBanner reports whether the default server start banner should be shown.
func (pb *ServiceBase) ShowStartBanner() bool {
	if pb == nil {
		return true
	}
	return !pb.hideStartBanner
}

// Start executes the root command.
//
// Command registration is intentionally owned by cmd/appinit.
func (pb *ServiceBase) Start() error {
	return pb.Execute()
}

// Execute initializes the application (if not already) and executes
// the root command with graceful shutdown support.
func (pb *ServiceBase) Execute() error {
	if !pb.skipBootstrap() {
		if err := pb.Bootstrap(); err != nil {
			return err
		}
	}

	done := make(chan bool, 1)

	go func() {
		sigch := make(chan os.Signal, 1)
		signal.Notify(sigch, os.Interrupt, syscall.SIGTERM)
		<-sigch
		done <- true
	}()

	go func() {
		_ = pb.RootCmd.Execute()
		done <- true
	}()

	<-done

	return pb.OnTerminate().Trigger(&core.TerminateEvent{
		App: pb,
	}, func(e *core.TerminateEvent) error {
		return e.App.ResetBootstrapState()
	})
}

func (pb *ServiceBase) eagerParseFlags(config *Config) error {
	pb.RootCmd.PersistentFlags().StringVar(
		&pb.dataDirFlag,
		"dir",
		config.DefaultDataDir,
		"the ServiceBase data directory",
	)

	pb.RootCmd.PersistentFlags().StringVar(
		&pb.encryptionEnvFlag,
		"encryptionEnv",
		config.DefaultEncryptionEnv,
		"the env variable whose value of 32 characters will be used \nas encryption key for the app settings (default none)",
	)

	pb.RootCmd.PersistentFlags().BoolVar(
		&pb.devFlag,
		"dev",
		config.DefaultDev,
		"enable dev mode, aka. printing logs and sql statements to the console",
	)

	return pb.RootCmd.ParseFlags(os.Args[1:])
}

func (pb *ServiceBase) skipBootstrap() bool {
	flags := []string{
		"-h",
		"--help",
		"-v",
		"--version",
	}

	if pb.IsBootstrapped() {
		return true
	}

	cmd, _, err := pb.RootCmd.Find(os.Args[1:])
	if err != nil {
		return true
	}

	for _, arg := range os.Args {
		if !list.ExistInSlice(arg, flags) {
			continue
		}

		trimmed := strings.TrimLeft(arg, "-")
		if len(trimmed) > 1 && cmd.Flags().Lookup(trimmed) == nil {
			return true
		}
		if len(trimmed) == 1 && cmd.Flags().ShorthandLookup(trimmed) == nil {
			return true
		}
	}

	return false
}

func inspectRuntime() (baseDir string, withGoRun bool) {
	if strings.HasPrefix(os.Args[0], os.TempDir()) {
		withGoRun = true
		baseDir, _ = os.Getwd()
	} else {
		withGoRun = false
		baseDir = filepath.Dir(os.Args[0])
	}

	return baseDir, withGoRun
}

func newErrWriter() *coloredWriter {
	return &coloredWriter{
		w: os.Stderr,
		c: color.New(color.FgRed),
	}
}

type coloredWriter struct {
	w io.Writer
	c *color.Color
}

func (colored *coloredWriter) Write(p []byte) (n int, err error) {
	colored.c.SetWriter(colored.w)
	defer colored.c.UnsetWriter(colored.w)

	return colored.c.Print(string(p))
}
