package updater

import (
	"context"
	"net/http"
)

// HttpClient is a base HTTP client interface (usually used for test purposes).
type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Config defines the updater service options.
//
// NB! This module is considered experimental and its config options may change.
type Config struct {
	// Owner specifies the account owner of the repository.
	Owner string

	// Repo specifies the repository name.
	Repo string

	// ArchiveExecutable specifies the executable file name inside the archive.
	ArchiveExecutable string

	// Context is used when fetching and downloading the latest release.
	Context context.Context

	// HttpClient is used when fetching and downloading the latest release.
	HttpClient HttpClient

	CreateBackup func(ctx context.Context, name string) error
}

func (c Config) normalized() Config {
	if c.Owner == "" {
		c.Owner = "pocketbase"
	}

	if c.Repo == "" {
		c.Repo = "pocketbase"
	}

	if c.ArchiveExecutable == "" {
		c.ArchiveExecutable = "pocketbase"
	}

	if c.HttpClient == nil {
		c.HttpClient = http.DefaultClient
	}

	if c.Context == nil {
		c.Context = context.Background()
	}

	return c
}
