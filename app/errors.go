package app

import "net/http"

type ApiError struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    map[string]any `json:"data"`
}

func (e *ApiError) Error() string { return e.Message }

func NewApiError(status int, message string, data any) *ApiError {
	if message == "" {
		message = http.StatusText(status)
	}

	apiErr := &ApiError{
		Code:    status,
		Message: message,
		Data:    map[string]any{},
	}

	if err, ok := data.(error); ok && err != nil {
		apiErr.Data["raw"] = err.Error()
	}

	return apiErr
}

func NewBadRequestError(message string, data any) error {
	return NewApiError(http.StatusBadRequest, message, data)
}

func NewUnauthorizedError(message string, data any) error {
	return NewApiError(http.StatusUnauthorized, message, data)
}

func NewForbiddenError(message string, data any) error {
	return NewApiError(http.StatusForbidden, message, data)
}

func NewNotFoundError(message string, data any) error {
	return NewApiError(http.StatusNotFound, message, data)
}
