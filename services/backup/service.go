package backup

import (
	"github.com/pafthang/servicebase/core"
	servicebase "github.com/pafthang/servicebase/services/base"
)

var Descriptor = servicebase.Descriptor{
	Name:    "backup",
	Purpose: "Wraps backup listing, creation, upload, restore and deletion workflows.",
	Dependencies: []string{
		"core.App",
		"services/backup/apis",
		"services/backup/forms",
		"services/backup/models",
		"services/base/forms",
		"file service",
		"team service",
		"filesystem",
	},
	RuntimeState: []string{
		"observes active backup/restore state via app store",
		"uses configured backups filesystem for archive storage",
	},
	Operations: []string{
		"Bind HTTP routes via services/backup/apis",
		"HasActiveBackup",
		"List",
		"NewCreateForm",
		"SubmitCreate",
		"NewUploadForm",
		"SubmitUpload",
		"CanDelete",
		"CanDownload",
		"BackupExists",
		"RestoreAsync",
		"Delete",
	},
}

type Service struct {
	servicebase.Service
}

func New(app core.App) *Service {
	return &Service{
		Service: servicebase.NewWithApp(app),
	}
}
