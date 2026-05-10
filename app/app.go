package app

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/pafthang/servicebase/core"
	"github.com/pafthang/servicebase/tools/list"
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

type ServeConfig struct {
	ShowStartBanner bool
	HttpAddr        string
	HttpsAddr       string

	CertificateDomains []string
	AllowedOrigins     []string

	Services      *Services
	EncryptionKey string
}

// Serve starts the application web server through appinit.
func Serve(app core.App, config ServeConfig) (*http.Server, error) {
	if app == nil {
		return nil, NewApiError(http.StatusInternalServerError, "app is required", nil)
	}

	if len(config.AllowedOrigins) == 0 {
		config.AllowedOrigins = []string{"*"}
	}

	if config.EncryptionKey == "" {
		config.EncryptionKey = os.Getenv("ENCRYPTION_KEY")
	}

	if err := RunMigrations(app); err != nil {
		return nil, err
	}

	if err := app.RefreshSettings(); err != nil {
		color.Yellow("=====================================")
		color.Yellow("WARNING: Settings load error! \n%v", err)
		color.Yellow("Fallback to the application defaults.")
		color.Yellow("=====================================")
	}

	services := config.Services
	if services == nil {
		var err error
		services, err = NewServices(ServicesConfig{App: app})
		if err != nil {
			return nil, err
		}
	}

	if err := RegisterHooks(HookConfig{
		App:            app,
		EncryptionKey:  config.EncryptionKey,
		LoggingService: services.Log,
	}); err != nil {
		return nil, err
	}

	if err := RegisterCrons(CronConfig{
		App:             app,
		CronScheduler:   services.CronScheduler,
		CronService:     services.Cron,
		EnableStatsJobs: true,
		EnableLogJobs:   true,
	}); err != nil {
		return nil, err
	}

	router, err := InitRouter(app, services)
	if err != nil {
		return nil, err
	}

	router.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		Skipper:      middleware.DefaultSkipper,
		AllowOrigins: config.AllowedOrigins,
		AllowMethods: []string{
			http.MethodGet,
			http.MethodHead,
			http.MethodPut,
			http.MethodPatch,
			http.MethodPost,
			http.MethodDelete,
		},
	}))

	mainAddr := config.HttpAddr
	if config.HttpsAddr != "" {
		mainAddr = config.HttpsAddr
	}

	var wwwRedirects []string
	hostNames := append([]string(nil), config.CertificateDomains...)
	if len(hostNames) == 0 {
		host, _, _ := net.SplitHostPort(mainAddr)
		if host != "" {
			hostNames = append(hostNames, host)
		}
	}

	for _, host := range hostNames {
		if strings.HasPrefix(host, "www.") {
			continue
		}

		wwwHost := "www." + host
		if !list.ExistInSlice(wwwHost, hostNames) {
			hostNames = append(hostNames, wwwHost)
			wwwRedirects = append(wwwRedirects, wwwHost)
		}
	}

	if len(wwwRedirects) > 0 {
		router.Pre(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				host := c.Request().Host
				if strings.HasPrefix(host, "www.") && list.ExistInSlice(host, wwwRedirects) {
					return c.Redirect(http.StatusTemporaryRedirect, c.Scheme()+"://"+host[4:]+c.Request().RequestURI)
				}
				return next(c)
			}
		})
	}

	certManager := &autocert.Manager{
		Prompt: autocert.AcceptTOS,
		Cache:  autocert.DirCache(filepath.Join(app.DataDir(), ".autocert_cache")),
	}
	if len(hostNames) > 0 {
		certManager.HostPolicy = autocert.HostWhitelist(hostNames...)
	}

	baseCtx, cancelBaseCtx := context.WithCancel(context.Background())
	defer cancelBaseCtx()

	server := &http.Server{
		TLSConfig: &tls.Config{
			MinVersion:     tls.VersionTLS12,
			GetCertificate: certManager.GetCertificate,
			NextProtos:     []string{acme.ALPNProto},
		},
		ReadTimeout:       10 * time.Minute,
		ReadHeaderTimeout: 30 * time.Second,
		Handler:           router,
		Addr:              mainAddr,
		BaseContext: func(l net.Listener) context.Context {
			return baseCtx
		},
		ErrorLog: log.New(&serverErrorLogWriter{app: app}, "", 0),
	}

	serveEvent := &core.ServeEvent{
		App:         app,
		Router:      router,
		Server:      server,
		CertManager: certManager,
	}
	if err := app.OnBeforeServe().Trigger(serveEvent); err != nil {
		return nil, err
	}

	if config.ShowStartBanner {
		printStartBanner(config, server)
	}

	var wg sync.WaitGroup
	app.OnTerminate().Add(func(e *core.TerminateEvent) error {
		cancelBaseCtx()

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		wg.Add(1)
		_ = server.Shutdown(ctx)
		if e.IsRestart {
			time.AfterFunc(5*time.Second, func() { wg.Done() })
		} else {
			wg.Done()
		}
		return nil
	})
	defer wg.Wait()

	if config.HttpsAddr != "" {
		if config.HttpAddr != "" {
			go func() { _ = http.ListenAndServe(config.HttpAddr, certManager.HTTPHandler(nil)) }()
		}
		return server, server.ListenAndServeTLS("", "")
	}

	return server, server.ListenAndServe()
}

func printStartBanner(config ServeConfig, server *http.Server) {
	schema := "http"
	addr := server.Addr
	if config.HttpsAddr != "" {
		schema = "https"
		if len(config.CertificateDomains) > 0 {
			addr = config.CertificateDomains[0]
		}
	}

	date := new(strings.Builder)
	log.New(date, "", log.LstdFlags).Print()

	bold := color.New(color.Bold).Add(color.FgGreen)
	bold.Printf("%s Server started at %s\n", strings.TrimSpace(date.String()), color.CyanString("%s://%s", schema, addr))

	regular := color.New()
	regular.Printf("├─ REST API: %s\n", color.CyanString("%s://%s/api/", schema, addr))
	regular.Printf("└─ Dashboard UI: %s\n", color.CyanString("%s://%s/_/", schema, addr))
}

type serverErrorLogWriter struct{ app core.App }

func (s *serverErrorLogWriter) Write(p []byte) (int, error) {
	if s.app != nil {
		s.app.Logger().Debug(strings.TrimSpace(string(p)))
	}
	return len(p), nil
}
