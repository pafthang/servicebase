package record

// Config defines record service options.
//
// The record service currently has no custom configuration, but the dedicated
// config type keeps the module shape aligned with the shared service template.
type Config struct{}

func (c Config) normalized() Config {
	return c
}
