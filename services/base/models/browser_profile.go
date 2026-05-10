package models

import (
	"github.com/pafthang/servicebase/tools/types"
)

type BrowserProfile struct {
	BaseModel
	User      string         `db:"user" json:"user"`
	Name      string         `db:"name" json:"name"`
	UserAgent string         `db:"user_agent" json:"user_agent"`
	Platform  string         `db:"platform" json:"platform"`
	IsActive  bool           `db:"is_active" json:"is_active"`
	LastUsed  types.DateTime `db:"last_used" json:"last_used"`
}

func (m *BrowserProfile) TableName() string {
	return "browser_profiles"
}
