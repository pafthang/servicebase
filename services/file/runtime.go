package file

import (
	"log/slog"

	"github.com/pafthang/servicebase/core"
)

func defaultLogger(app core.App) *slog.Logger {
	if app != nil && app.Logger() != nil {
		return app.Logger()
	}

	return slog.Default()
}
