package realtime

type Config struct {
	Enabled bool
}

func DefaultConfig() Config {
	return Config{
		Enabled: true,
	}
}
