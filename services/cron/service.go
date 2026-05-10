package cron

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"regexp"
	"strings"

	"github.com/pafthang/servicebase/core"
	cronmodels "github.com/pafthang/servicebase/services/cron/models"
	"github.com/pafthang/servicebase/tools/types"
)

type CronService struct {
	scheduler            *Scheduler
	executor             *Executor
	app                  core.App
	logger               *slog.Logger
	maxCronsPerUser      int
	minExecutionInterval time.Duration
}

func NewCronService(scheduler *Scheduler, executor *Executor) *CronService {
	var app core.App
	if scheduler != nil && scheduler.app != nil {
		app = scheduler.app
	} else if executor != nil {
		app = executor.app
	}

	return &CronService{
		scheduler:            scheduler,
		executor:             executor,
		app:                  app,
		logger:               defaultLogger(app),
		maxCronsPerUser:      10,
		minExecutionInterval: 1 * time.Minute,
	}
}

type CreateCronData struct {
	Name                string                 `json:"name"`
	Description         string                 `json:"description"`
	Schedule            string                 `json:"schedule"`
	WebhookURL          string                 `json:"webhook_url"`
	WebhookMethod       string                 `json:"webhook_method"`
	WebhookHeaders      map[string]interface{} `json:"webhook_headers"`
	WebhookPayload      map[string]interface{} `json:"webhook_payload"`
	TimeoutSeconds      int                    `json:"timeout_seconds"`
	NotifyOnSuccess     bool                   `json:"notify_on_success"`
	NotifyOnFailure     bool                   `json:"notify_on_failure"`
	NotificationWebhook string                 `json:"notification_webhook"`
	MaxRetries          int                    `json:"max_retries"`
	RetryDelaySeconds   int                    `json:"retry_delay_seconds"`
}

type UpdateCronData struct {
	Name                *string                `json:"name"`
	Description         *string                `json:"description"`
	Schedule            *string                `json:"schedule"`
	WebhookURL          *string                `json:"webhook_url"`
	WebhookMethod       *string                `json:"webhook_method"`
	WebhookHeaders      map[string]interface{} `json:"webhook_headers"`
	WebhookPayload      map[string]interface{} `json:"webhook_payload"`
	TimeoutSeconds      *int                   `json:"timeout_seconds"`
	IsActive            *bool                  `json:"is_active"`
	NotifyOnSuccess     *bool                  `json:"notify_on_success"`
	NotifyOnFailure     *bool                  `json:"notify_on_failure"`
	NotificationWebhook *string                `json:"notification_webhook"`
	MaxRetries          *int                   `json:"max_retries"`
	RetryDelaySeconds   *int                   `json:"retry_delay_seconds"`
}

func (s *CronService) CreateCron(userId string, data CreateCronData) (*cronmodels.Cron, error) {
	count, err := countCrons(s.app, map[string]any{
		"user":      userId,
		"is_system": false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to count crons: %w", err)
	}

	if int(count) >= s.maxCronsPerUser {
		return nil, newValidationError("cron", fmt.Sprintf("maximum %d crons allowed per user", s.maxCronsPerUser))
	}

	if err := s.validateCronData(data.Schedule, data.WebhookURL, data.WebhookMethod, data.TimeoutSeconds); err != nil {
		return nil, err
	}

	cronRecord := &cronmodels.Cron{
		User:                userId,
		Name:                data.Name,
		Description:         data.Description,
		Schedule:            data.Schedule,
		WebhookURL:          data.WebhookURL,
		WebhookMethod:       s.getWebhookMethod(data.WebhookMethod),
		IsActive:            true,
		IsSystem:            false,
		TimeoutSeconds:      s.getTimeoutSeconds(data.TimeoutSeconds),
		NotifyOnSuccess:     data.NotifyOnSuccess,
		NotifyOnFailure:     data.NotifyOnFailure,
		NotificationWebhook: data.NotificationWebhook,
		MaxRetries:          data.MaxRetries,
		RetryDelaySeconds:   s.getRetryDelaySeconds(data.RetryDelaySeconds),
	}

	if data.WebhookHeaders != nil {
		headersBytes, err := json.Marshal(data.WebhookHeaders)
		if err != nil {
			return nil, newValidationError("webhook_headers", "invalid JSON")
		}
		cronRecord.WebhookHeaders = types.JsonRaw(headersBytes)
	}

	if data.WebhookPayload != nil {
		payloadBytes, err := json.Marshal(data.WebhookPayload)
		if err != nil {
			return nil, newValidationError("webhook_payload", "invalid JSON")
		}
		cronRecord.WebhookPayload = types.JsonRaw(payloadBytes)
	}

	cronRecord.RefreshCreated()
	cronRecord.RefreshUpdated()

	if err := saveCron(s.app, cronRecord); err != nil {
		return nil, fmt.Errorf("failed to save cron: %w", err)
	}

	return cronRecord, nil
}

func (s *CronService) UpdateCron(cronId string, userId string, data UpdateCronData) (*cronmodels.Cron, error) {
	cronRecord, err := findCronByID(s.app, cronId)
	if err != nil {
		return nil, newNotFoundError("cron not found")
	}

	if cronRecord.User != userId && !cronRecord.IsSystem {
		return nil, errForbidden
	}

	if cronRecord.IsSystem {
		return nil, newValidationError("cron", "system crons cannot be updated")
	}

	needsReschedule := false

	if data.Name != nil {
		cronRecord.Name = *data.Name
	}
	if data.Description != nil {
		cronRecord.Description = *data.Description
	}
	if data.Schedule != nil {
		if err := s.validateSchedule(*data.Schedule); err != nil {
			return nil, err
		}
		cronRecord.Schedule = *data.Schedule
		needsReschedule = true
	}
	if data.WebhookURL != nil {
		if err := s.validateWebhookURL(*data.WebhookURL); err != nil {
			return nil, err
		}
		cronRecord.WebhookURL = *data.WebhookURL
	}
	if data.WebhookMethod != nil {
		cronRecord.WebhookMethod = s.getWebhookMethod(*data.WebhookMethod)
	}
	if data.WebhookHeaders != nil {
		headersBytes, err := json.Marshal(data.WebhookHeaders)
		if err != nil {
			return nil, newValidationError("webhook_headers", "invalid JSON")
		}
		cronRecord.WebhookHeaders = types.JsonRaw(headersBytes)
	}
	if data.WebhookPayload != nil {
		payloadBytes, err := json.Marshal(data.WebhookPayload)
		if err != nil {
			return nil, newValidationError("webhook_payload", "invalid JSON")
		}
		cronRecord.WebhookPayload = types.JsonRaw(payloadBytes)
	}
	if data.TimeoutSeconds != nil {
		cronRecord.TimeoutSeconds = s.getTimeoutSeconds(*data.TimeoutSeconds)
	}
	if data.IsActive != nil {
		cronRecord.IsActive = *data.IsActive
		needsReschedule = true
	}
	if data.NotifyOnSuccess != nil {
		cronRecord.NotifyOnSuccess = *data.NotifyOnSuccess
	}
	if data.NotifyOnFailure != nil {
		cronRecord.NotifyOnFailure = *data.NotifyOnFailure
	}
	if data.NotificationWebhook != nil {
		cronRecord.NotificationWebhook = *data.NotificationWebhook
	}
	if data.MaxRetries != nil {
		cronRecord.MaxRetries = *data.MaxRetries
	}
	if data.RetryDelaySeconds != nil {
		cronRecord.RetryDelaySeconds = s.getRetryDelaySeconds(*data.RetryDelaySeconds)
	}

	cronRecord.RefreshUpdated()

	if err := saveCron(s.app, cronRecord); err != nil {
		return nil, fmt.Errorf("failed to update cron: %w", err)
	}

	if needsReschedule {
		if cronRecord.IsActive {
			if err := s.scheduler.AddCron(cronRecord); err != nil {
				s.logger.Error("Failed to update cron in scheduler", "cronId", cronId, "error", err)
			}
		} else {
			if err := s.scheduler.RemoveCron(cronId); err != nil {
				s.logger.Error("Failed to remove cron from scheduler", "cronId", cronId, "error", err)
			}
		}
	}

	return cronRecord, nil
}

func (s *CronService) DeleteCron(cronId string, userId string) error {
	cronRecord, err := findCronByID(s.app, cronId)
	if err != nil {
		return newNotFoundError("cron not found")
	}

	if cronRecord.User != userId && !cronRecord.IsSystem {
		return errForbidden
	}

	if cronRecord.IsSystem {
		return newValidationError("cron", "system crons cannot be deleted")
	}

	if err := s.scheduler.RemoveCron(cronId); err != nil {
		s.logger.Error("Failed to remove cron from scheduler", "cronId", cronId, "error", err)
	}

	if err := deleteCron(s.app, cronRecord); err != nil {
		return fmt.Errorf("failed to delete cron: %w", err)
	}

	return nil
}

func (s *CronService) GetCron(cronId string, userId string) (*cronmodels.Cron, error) {
	cronRecord, err := findCronByID(s.app, cronId)
	if err != nil {
		return nil, newNotFoundError("cron not found")
	}

	if cronRecord.User != userId && !cronRecord.IsSystem {
		return nil, errForbidden
	}

	return cronRecord, nil
}

func (s *CronService) ListCrons(userId string, filters map[string]interface{}) ([]*cronmodels.Cron, error) {
	filterMap := map[string]interface{}{
		"user": userId,
	}

	if isSystem, ok := filters["is_system"].(bool); ok {
		filterMap["is_system"] = isSystem
	} else {
		filterMap["is_system"] = false
	}

	if isActive, ok := filters["is_active"].(bool); ok {
		filterMap["is_active"] = isActive
	}

	crons, err := listCrons(s.app, filterMap)
	if err != nil {
		return nil, fmt.Errorf("failed to list crons: %w", err)
	}

	return crons, nil
}

func (s *CronService) PauseCron(cronId string, userId string) error {
	_, err := s.UpdateCron(cronId, userId, UpdateCronData{
		IsActive: boolPtr(false),
	})
	return err
}

func (s *CronService) ResumeCron(cronId string, userId string) error {
	_, err := s.UpdateCron(cronId, userId, UpdateCronData{
		IsActive: boolPtr(true),
	})
	return err
}

func (s *CronService) TestCron(cronId string, userId string) error {
	cronRecord, err := s.GetCron(cronId, userId)
	if err != nil {
		return err
	}

	if s.scheduler != nil {
		return s.scheduler.ExecuteCron(cronId)
	}

	err = s.executor.Execute(cronRecord)
	if err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}

	return nil
}

func (s *CronService) CloneCron(cronId string, userId string, newName string) (*cronmodels.Cron, error) {
	originalCron, err := s.GetCron(cronId, userId)
	if err != nil {
		return nil, err
	}

	var webhookHeaders map[string]interface{}
	if originalCron.WebhookHeaders != nil {
		if err := json.Unmarshal(originalCron.WebhookHeaders, &webhookHeaders); err != nil {
		}
	}

	var webhookPayload map[string]interface{}
	if originalCron.WebhookPayload != nil {
		if err := json.Unmarshal(originalCron.WebhookPayload, &webhookPayload); err != nil {
		}
	}

	createData := CreateCronData{
		Name:                newName,
		Description:         originalCron.Description,
		Schedule:            originalCron.Schedule,
		WebhookURL:          originalCron.WebhookURL,
		WebhookMethod:       originalCron.WebhookMethod,
		WebhookHeaders:      webhookHeaders,
		WebhookPayload:      webhookPayload,
		TimeoutSeconds:      originalCron.TimeoutSeconds,
		NotifyOnSuccess:     originalCron.NotifyOnSuccess,
		NotifyOnFailure:     originalCron.NotifyOnFailure,
		NotificationWebhook: originalCron.NotificationWebhook,
		MaxRetries:          originalCron.MaxRetries,
		RetryDelaySeconds:   originalCron.RetryDelaySeconds,
	}

	return s.CreateCron(userId, createData)
}

func (s *CronService) CreateSystemCron(systemType string, config CreateCronData) (*cronmodels.Cron, error) {
	existing, err := findCronByFilter(s.app, map[string]any{
		"is_system":   true,
		"system_type": systemType,
	})
	if err == nil {
		return existing, nil
	}

	cronRecord := &cronmodels.Cron{
		Name:                config.Name,
		Description:         config.Description,
		Schedule:            config.Schedule,
		WebhookURL:          config.WebhookURL,
		WebhookMethod:       s.getWebhookMethod(config.WebhookMethod),
		IsActive:            true,
		IsSystem:            true,
		SystemType:          systemType,
		TimeoutSeconds:      s.getTimeoutSeconds(config.TimeoutSeconds),
		NotifyOnSuccess:     config.NotifyOnSuccess,
		NotifyOnFailure:     config.NotifyOnFailure,
		NotificationWebhook: config.NotificationWebhook,
		MaxRetries:          config.MaxRetries,
		RetryDelaySeconds:   s.getRetryDelaySeconds(config.RetryDelaySeconds),
	}

	if config.WebhookHeaders != nil {
		headersBytes, _ := json.Marshal(config.WebhookHeaders)
		cronRecord.WebhookHeaders = types.JsonRaw(headersBytes)
	}

	if config.WebhookPayload != nil {
		payloadBytes, _ := json.Marshal(config.WebhookPayload)
		cronRecord.WebhookPayload = types.JsonRaw(payloadBytes)
	}

	cronRecord.RefreshCreated()
	cronRecord.RefreshUpdated()

	if err := saveCron(s.app, cronRecord); err != nil {
		return nil, fmt.Errorf("failed to save system cron: %w", err)
	}

	if err := s.scheduler.AddCron(cronRecord); err != nil {
		s.logger.Error("Failed to add system cron to scheduler", "systemType", systemType, "error", err)
	}

	return cronRecord, nil
}

func (s *CronService) DeleteSystemCron(systemType string) error {
	cronRecord, err := findCronByFilter(s.app, map[string]any{
		"is_system":   true,
		"system_type": systemType,
	})
	if err != nil {
		return newNotFoundError("system cron not found")
	}

	if err := s.scheduler.RemoveCron(cronRecord.Id); err != nil {
		s.logger.Error("Failed to remove system cron from scheduler", "systemType", systemType, "error", err)
	}

	return deleteCron(s.app, cronRecord)
}

func (s *CronService) GetSystemCron(systemType string) (*cronmodels.Cron, error) {
	cronRecord, err := findCronByFilter(s.app, map[string]any{
		"is_system":   true,
		"system_type": systemType,
	})
	if err != nil {
		return nil, newNotFoundError("system cron not found")
	}

	return cronRecord, nil
}

func (s *CronService) validateCronData(schedule, webhookURL, webhookMethod string, timeoutSeconds int) error {
	if err := s.validateSchedule(schedule); err != nil {
		return err
	}

	if err := s.validateWebhookURL(webhookURL); err != nil {
		return err
	}

	if timeoutSeconds < 1 || timeoutSeconds > 300 {
		return newValidationError("timeout_seconds", "must be between 1 and 300")
	}

	return nil
}

func (s *CronService) validateSchedule(schedule string) error {
	if schedule == "" {
		return newValidationError("schedule", "schedule cannot be empty")
	}

	parts := strings.Fields(schedule)
	if len(parts) != 5 {
		return newValidationError("schedule", "cron expression must have 5 fields: minute hour day month weekday")
	}

	cronFieldPattern := regexp.MustCompile(`^(\*|(\d+(-\d+)?)((,\d+(-\d+)?)*)(/\d+)?)$`)
	for i, part := range parts {
		if !cronFieldPattern.MatchString(part) {
			fieldNames := []string{"minute", "hour", "day", "month", "weekday"}
			return newValidationError("schedule", fmt.Sprintf("invalid %s field: %s", fieldNames[i], part))
		}
	}

	return nil
}

func (s *CronService) validateWebhookURL(webhookURL string) error {
	parsedURL, err := url.Parse(webhookURL)
	if err != nil {
		return newValidationError("webhook_url", "invalid URL format")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return newValidationError("webhook_url", "only http and https schemes are allowed")
	}

	if parsedURL.Host == "" {
		return newValidationError("webhook_url", "host is required")
	}

	return nil
}

func (s *CronService) getWebhookMethod(method string) string {
	if method == "" {
		return "POST"
	}

	validMethods := map[string]bool{
		"GET":    true,
		"POST":   true,
		"PUT":    true,
		"DELETE": true,
		"PATCH":  true,
	}

	if validMethods[method] {
		return method
	}

	return "POST"
}

func (s *CronService) getTimeoutSeconds(timeoutSeconds int) int {
	if timeoutSeconds < 1 {
		return 30
	}
	if timeoutSeconds > 300 {
		return 300
	}
	return timeoutSeconds
}

func (s *CronService) getRetryDelaySeconds(delaySeconds int) int {
	if delaySeconds < 1 {
		return 60
	}
	if delaySeconds > 3600 {
		return 3600
	}
	return delaySeconds
}

func boolPtr(b bool) *bool {
	return &b
}
