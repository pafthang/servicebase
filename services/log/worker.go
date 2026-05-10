package log

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/pafthang/servicebase/core"
	logmodels "github.com/pafthang/servicebase/services/log/models"
)

func (s *Service) worker() {
	defer s.wg.Done()

	var batch []*LogEntry

	for {
		select {
		case <-s.ctx.Done():
			if len(batch) > 0 {
				s.processBatch(batch)
			}
			return
		case <-s.flushTicker.C:
			if len(batch) > 0 {
				s.processBatch(batch)
				batch = nil
			}

			remaining := s.drainBuffer()
			if len(remaining) > 0 {
				s.processBatch(remaining)
			}
		case entry := <-s.buffer:
			batch = append(batch, entry)
			if len(batch) >= s.config.BatchSize {
				s.processBatch(batch)
				batch = nil
			}
		}
	}
}

func (s *Service) processBatch(logs []*LogEntry) {
	s.workers <- struct{}{}
	s.wg.Add(1)
	go func(lgs []*LogEntry) {
		defer func() { <-s.workers }()
		defer s.wg.Done()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := s.writeBatch(ctx, lgs); err != nil {
			defaultLogger(s.app).Error("Error flushing logs", "error", err)
		}
	}(logs)
}

func (s *Service) drainBuffer() []*LogEntry {
	logs := make([]*LogEntry, 0, s.config.BatchSize)
	for {
		select {
		case entry := <-s.buffer:
			logs = append(logs, entry)
		default:
			return logs
		}
	}
}

func (s *Service) writeBatch(ctx context.Context, entries []*LogEntry) error {
	if len(entries) == 0 {
		return nil
	}

	defaultLogger(s.app).Info("Flushing logs to database", "count", len(entries))

	var lastErr error
	for attempt := 0; attempt < s.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			defaultLogger(s.app).Warn("Retrying log batch write", "attempt", attempt+1)
			select {
			case <-time.After(s.config.RetryBackoff * time.Duration(attempt)):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		lastErr = runInTransaction(s.app, func(txApp core.App) error {
			for _, entry := range entries {
				ts, _ := time.Parse(time.RFC3339, entry.Timestamp)

				logModel := &logmodels.AppLog{
					Project: entry.ProjectId,
					Level:   entry.Level,
					Source:  entry.Source,
					Message: entry.Message,
					TraceID: entry.TraceID,
				}
				logModel.Timestamp.Scan(ts)

				if err := logModel.SetContextMap(entry.Context); err != nil {
					defaultLogger(s.app).Error("Failed to set log context map", "error", err)
					continue
				}

				if err := saveAppLog(txApp, logModel); err != nil {
					atomic.AddInt64(&s.errors, 1)
					defaultLogger(s.app).Error("Failed to save log record", "error", err, "project", entry.ProjectId)
					return err
				}
				atomic.AddInt64(&s.logsFlushed, 1)
			}
			return nil
		})

		if lastErr == nil {
			defaultLogger(s.app).Info("Successfully flushed log batch", "count", len(entries))
			return nil
		}
	}

	return fmt.Errorf("failed to write batch after %d attempts: %w", s.config.RetryAttempts, lastErr)
}
