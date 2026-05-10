package models

import (
	basemodels "github.com/pafthang/servicebase/services/base/models"
	"github.com/pafthang/servicebase/tools/types"
)

var _ basemodels.Model = (*Log)(nil)

type Log struct {
	basemodels.BaseModel

	Data    types.JsonMap `db:"data" json:"data"`
	Message string        `db:"message" json:"message"`
	Level   int           `db:"level" json:"level"`
}

func (m *Log) TableName() string {
	return "_logs"
}
