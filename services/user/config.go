package user

// Config defines user service options.
//
// The service currently has no dedicated knobs, but the type keeps the module
// aligned with the shared service layout.
type Config struct{}

func (c Config) normalized() Config {
	return c
}
