package apis

import (
	"errors"

	"github.com/labstack/echo/v5"
)

func httpError(status int, message string, data any) error {
	err := echo.NewHTTPError(status, message)
	if internal := asError(data); internal != nil {
		err.Internal = internal
	}
	return err
}

func asError(data any) error {
	switch v := data.(type) {
	case nil:
		return nil
	case error:
		return v
	case string:
		if v == "" {
			return nil
		}
		return errors.New(v)
	default:
		return nil
	}
}
