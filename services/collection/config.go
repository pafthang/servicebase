package collection

// Config defines collection service options.
//
// The collection service currently doesn't need custom knobs, but the explicit
// config type keeps the module layout uniform with the shared service template.
type Config struct{}

func (c Config) normalized() Config {
	return c
}
