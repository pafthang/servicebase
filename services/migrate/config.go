package migrate

import "strings"

// Config defines the migrate service options.
type Config struct {
	// Dir specifies the directory with the user defined migrations.
	//
	// If not set it defaults to a path relative to the app data directory.
	Dir string

	// Automigrate specifies whether to enable collection automigrations.
	Automigrate bool
}

func (c Config) normalized() Config {
	c.Dir = strings.TrimSpace(c.Dir)

	return c
}
