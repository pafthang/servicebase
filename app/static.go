package app

import (
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v5"
)

// StaticDirectoryHandler is similar to echo.StaticDirectoryHandler, but without
// the directory redirect that conflicts with RemoveTrailingSlash middleware.
// If a file resource is missing and indexFallback is true, index.html is served.
func StaticDirectoryHandler(fileSystem fs.FS, indexFallback bool) echo.HandlerFunc {
	return func(c echo.Context) error {
		p := c.PathParam("*")

		tmpPath, err := url.PathUnescape(p)
		if err != nil {
			return fmt.Errorf("failed to unescape path variable: %w", err)
		}
		p = tmpPath

		name := filepath.ToSlash(filepath.Clean(strings.TrimPrefix(p, "/")))
		fileErr := c.FileFS(name, fileSystem)
		if fileErr != nil && indexFallback && errors.Is(fileErr, echo.ErrNotFound) {
			return c.FileFS("index.html", fileSystem)
		}

		return fileErr
	}
}
