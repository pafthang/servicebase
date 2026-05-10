package team

// Config defines team service options.
//
// The current team service has no custom configuration, but the type keeps the
// module aligned with the shared service template.
type Config struct{}

func (c Config) normalized() Config {
	return c
}
