package apis

import (
	"github.com/labstack/echo/v5"
	recordmodels "github.com/pafthang/servicebase/services/record/models"
)

const contextAuthRecordKey = "authRecord"

func currentProjectUserID(c echo.Context) (string, bool) {
	record, _ := c.Get(contextAuthRecordKey).(*recordmodels.Record)
	if record == nil || record.Id == "" {
		return "", false
	}
	return record.Id, true
}
