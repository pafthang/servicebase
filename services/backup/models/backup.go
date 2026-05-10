package models

import "github.com/pafthang/servicebase/tools/types"

type BackupFileInfo struct {
	Key      string         `json:"key"`
	Size     int64          `json:"size"`
	Modified types.DateTime `json:"modified"`
}
