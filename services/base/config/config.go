package config

import (
	"os"
	"strconv"
	"time"
)

// String resolves a string setting using the following precedence:
// explicit value, env var, default value.
func String(value string, envKey string, defaultValue string) string {
	if value != "" {
		return value
	}

	if envKey != "" {
		if envValue := os.Getenv(envKey); envValue != "" {
			return envValue
		}
	}

	return defaultValue
}

// Int resolves an int setting using the following precedence:
// explicit value, env var, default value.
func Int(value int, envKey string, defaultValue int) int {
	if value != 0 {
		return value
	}

	if envKey != "" {
		if envValue := os.Getenv(envKey); envValue != "" {
			if parsed, err := strconv.Atoi(envValue); err == nil {
				return parsed
			}
		}
	}

	return defaultValue
}

// Bool resolves a bool setting using the following precedence:
// explicit value, env var, default value.
func Bool(value *bool, envKey string, defaultValue bool) bool {
	if value != nil {
		return *value
	}

	if envKey != "" {
		if envValue := os.Getenv(envKey); envValue != "" {
			if parsed, err := strconv.ParseBool(envValue); err == nil {
				return parsed
			}
		}
	}

	return defaultValue
}

// Duration resolves a duration setting using the following precedence:
// explicit value, env var, default value.
func Duration(value time.Duration, envKey string, defaultValue time.Duration) time.Duration {
	if value != 0 {
		return value
	}

	if envKey != "" {
		if envValue := os.Getenv(envKey); envValue != "" {
			if parsed, err := time.ParseDuration(envValue); err == nil {
				return parsed
			}
		}
	}

	return defaultValue
}
