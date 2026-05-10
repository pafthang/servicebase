package mails

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/pafthang/servicebase/core"
)

var (
	mailApp   core.App
	mailAppMu sync.RWMutex
	logger    mailLoggerShim
)

type mailLoggerShim struct {
	Debug mailDebugShim
}

type mailDebugShim struct{}

func (mailLoggerShim) LogInfo(msg string, args ...any) {
	defaultLogger(getMailApp()).Info(msg, normalizeArgs(args)...)
}

func (mailLoggerShim) LogWarning(msg string, args ...any) {
	defaultLogger(getMailApp()).Warn(msg, normalizeArgs(args)...)
}

func (mailLoggerShim) LogError(msg string, args ...any) {
	defaultLogger(getMailApp()).Error(msg, normalizeArgs(args)...)
}

func (mailDebugShim) Printf(format string, args ...any) {
	defaultLogger(getMailApp()).Debug(fmt.Sprintf(format, args...))
}

func defaultLogger(app core.App) *slog.Logger {
	if app != nil && app.Logger() != nil {
		return app.Logger()
	}
	return slog.Default()
}

func setMailApp(app core.App) {
	mailAppMu.Lock()
	mailApp = app
	mailAppMu.Unlock()
}

func getMailApp() core.App {
	mailAppMu.RLock()
	defer mailAppMu.RUnlock()
	return mailApp
}

func normalizeArgs(args ...any) []any {
	if len(args) == 1 {
		if err, ok := args[0].(error); ok {
			return []any{"error", err}
		}
	}
	return args
}
