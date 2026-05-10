package log

import (
	"time"

	baseconfig "github.com/pafthang/servicebase/services/base/config"
)

type Config struct {
	BatchSize      int
	FlushInterval  time.Duration
	BufferSize     int
	WorkerPoolSize int
	RetryAttempts  int
	RetryBackoff   time.Duration
}

func DefaultConfig() *Config {
	return &Config{
		BatchSize:      baseconfig.Int(0, "LOGGING_BATCH_SIZE", 100),
		FlushInterval:  baseconfig.Duration(0, "LOGGING_FLUSH_INTERVAL", 5*time.Second),
		BufferSize:     baseconfig.Int(0, "LOGGING_BUFFER_SIZE", 5000),
		WorkerPoolSize: baseconfig.Int(0, "LOGGING_WORKER_POOL_SIZE", 5),
		RetryAttempts:  baseconfig.Int(0, "LOGGING_RETRY_ATTEMPTS", 3),
		RetryBackoff:   baseconfig.Duration(0, "LOGGING_RETRY_BACKOFF", 1*time.Second),
	}
}
