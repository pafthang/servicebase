package log

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pafthang/servicebase/core"
	"github.com/pafthang/servicebase/daos"
	servicebase "github.com/pafthang/servicebase/services/base"
	logmodels "github.com/pafthang/servicebase/services/log/models"
	"github.com/pocketbase/dbx"

	"github.com/pafthang/servicebase/tools/search"
)

var Descriptor = servicebase.Descriptor{
	Name:    "log",
	Purpose: "Provides system log read-side APIs plus project app-log ingest, batching, retention cleanup and stats.",
	Dependencies: []string{
		"core.App",
		"services/base/models",
		"tools/search",
	},
	Operations: []string{
		"List",
		"Stats",
		"FindByID",
		"IngestLogs",
		"Shutdown",
		"GetStats",
	},
}

type Service struct {
	app core.App

	config      *Config
	buffer      chan *LogEntry
	flushTicker *time.Ticker
	workers     chan struct{}
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc

	logsBuffered    int64
	logsFlushed     int64
	errors          int64
	bufferOverflows int64
}

func New(app core.App) *Service {
	config := DefaultConfig()
	ctx, cancel := context.WithCancel(context.Background())

	s := &Service{
		app:         app,
		config:      config,
		buffer:      make(chan *LogEntry, config.BufferSize),
		flushTicker: time.NewTicker(config.FlushInterval),
		workers:     make(chan struct{}, config.WorkerPoolSize),
		ctx:         ctx,
		cancel:      cancel,
	}

	s.wg.Add(1)
	go s.worker()

	return s
}

func (s *Service) List(query string, fieldResolver *search.SimpleFieldResolver) (*search.Result, error) {
	return search.NewProvider(fieldResolver).
		Query(s.app.LogsDao().LogQuery()).
		ParseAndExec(query, &[]*logmodels.Log{})
}

func (s *Service) Stats(expr dbx.Expression) ([]*daos.LogsStatsItem, error) {
	return s.app.LogsDao().LogsStats(expr)
}

func (s *Service) FindByID(id string) (*logmodels.Log, error) {
	return s.app.LogsDao().FindLogById(id)
}

type LogEntry struct {
	Level     string                 `json:"level"`
	Timestamp string                 `json:"timestamp"`
	Source    string                 `json:"source"`
	Message   string                 `json:"message"`
	Context   map[string]interface{} `json:"context"`
	TraceID   string                 `json:"trace_id"`
	ProjectId string                 `json:"-"`
}

type IngestResult struct {
	Success   bool     `json:"success"`
	Processed int      `json:"processed"`
	Failed    int      `json:"failed"`
	Errors    []string `json:"errors,omitempty"`
}

type LoggingStats struct {
	LogsBuffered    int64 `json:"logs_buffered"`
	LogsFlushed     int64 `json:"logs_flushed"`
	Errors          int64 `json:"errors"`
	BufferOverflows int64 `json:"buffer_overflows"`
}

func (s *Service) IngestLogs(ctx context.Context, projectId string, entries []LogEntry) (*IngestResult, error) {
	defaultLogger(s.app).Debug("Ingesting logs", "count", len(entries), "projectId", projectId)
	if len(entries) > 1000 {
		return nil, errors.New("batch size exceeds limit of 1000")
	}

	result := &IngestResult{Errors: []string{}}
	now := time.Now()

	for i := range entries {
		entry := &entries[i]
		entry.ProjectId = projectId

		if entry.Level == "" {
			entry.Level = "info"
		}

		if entry.Message == "" {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("entry %d: message is required", i))
			continue
		}
		if len(entry.Message) > 64*1024 {
			entry.Message = entry.Message[:64*1024-3] + "..."
		}

		if entry.Timestamp != "" {
			ts, err := time.Parse(time.RFC3339, entry.Timestamp)
			if err != nil {
				result.Failed++
				result.Errors = append(result.Errors, fmt.Sprintf("entry %d: invalid timestamp format", i))
				continue
			}

			if ts.Before(now.Add(-24*time.Hour)) || ts.After(now.Add(24*time.Hour)) {
				result.Failed++
				result.Errors = append(result.Errors, fmt.Sprintf("entry %d: timestamp outside ±24h window", i))
				continue
			}
		} else {
			entry.Timestamp = now.Format(time.RFC3339)
		}

		atomic.AddInt64(&s.logsBuffered, 1)
		select {
		case s.buffer <- entry:
			result.Processed++
		default:
			atomic.AddInt64(&s.bufferOverflows, 1)
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("entry %d: buffer full, log dropped", i))
		}
	}

	result.Success = result.Failed == 0
	return result, nil
}

func (s *Service) Shutdown(ctx context.Context) error {
	defaultLogger(s.app).Info("Shutting down log service...")

	s.flushTicker.Stop()
	s.cancel()

	remainingLogs := s.drainBuffer()
	if len(remainingLogs) > 0 {
		if err := s.writeBatch(ctx, remainingLogs); err != nil {
			defaultLogger(s.app).Error("Error flushing logs during shutdown", "error", err)
		}
	}

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		defaultLogger(s.app).Info("Log service shut down successfully")
		return nil
	case <-ctx.Done():
		return fmt.Errorf("shutdown timeout: %w", ctx.Err())
	}
}

func (s *Service) GetStats() LoggingStats {
	return LoggingStats{
		LogsBuffered:    atomic.LoadInt64(&s.logsBuffered),
		LogsFlushed:     atomic.LoadInt64(&s.logsFlushed),
		Errors:          atomic.LoadInt64(&s.errors),
		BufferOverflows: atomic.LoadInt64(&s.bufferOverflows),
	}
}
