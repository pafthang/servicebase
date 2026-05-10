package cron

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/pafthang/servicebase/core"
)

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
	Err     error  `json:"-"`
}

func (e *APIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func defaultLogger(app core.App) *slog.Logger {
	if app != nil && app.Logger() != nil {
		return app.Logger()
	}
	return slog.Default()
}

func newValidationError(field, reason string) *APIError {
	return &APIError{
		Code:    "VALIDATION_ERROR",
		Message: fmt.Sprintf("Validation failed for field '%s': %s", field, reason),
		Status:  http.StatusBadRequest,
	}
}

func newNotFoundError(message string) *APIError {
	return &APIError{
		Code:    "NOT_FOUND",
		Message: message,
		Status:  http.StatusNotFound,
	}
}

var errForbidden = &APIError{
	Code:    "FORBIDDEN",
	Message: "Access denied",
	Status:  http.StatusForbidden,
}
