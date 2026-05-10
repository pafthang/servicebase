package models

import (
	"github.com/pafthang/servicebase/tools/types"
)

var _ Model = (*TrackFocus)(nil)

type TrackFocus struct {
	BaseModel

	User      string                  `db:"user" json:"user"`
	Device    string                  `db:"device" json:"device"`
	Tags      types.JsonArray[string] `db:"tags" json:"tags"`
	Metadata  string                  `db:"metadata" json:"metadata"`
	BeginDate types.DateTime          `db:"begin_date" json:"begin_date"`
	EndDate   types.DateTime          `db:"end_date" json:"end_date"`
}

func (m *TrackFocus) TableName() string {
	return "track_focus"
}
