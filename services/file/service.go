package file

import (
	"errors"
	"strings"

	"github.com/pafthang/servicebase/core"
	servicebase "github.com/pafthang/servicebase/services/base"
	basemodels "github.com/pafthang/servicebase/services/base/models"
	recordmodels "github.com/pafthang/servicebase/services/record/models"
	"github.com/pafthang/servicebase/tools/security"
	"github.com/pafthang/servicebase/tools/tokens"

	"github.com/spf13/cast"
)

var Descriptor = servicebase.Descriptor{
	Name:    "file",
	Purpose: "Owns file token issuance, protected download resolution, zip/export helpers, quota and file schema.",
	Dependencies: []string{
		"core.App",
		"tokens",
		"security",
		"services/file/models",
	},
	Operations: []string{
		"NewFileToken",
		"FindAuthRecordByFileToken",
		"StreamSelectionAsZipWithApp",
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

func (s *Service) NewFileToken(model basemodels.Model) (string, error) {
	if record, ok := model.(*recordmodels.Record); ok && record != nil {
		return tokens.NewRecordFileToken(s.App(), record)
	}

	return "", errors.New("unsupported file token model")
}

func (s *Service) FindAuthRecordByFileToken(fileToken string) (*recordmodels.Record, error) {
	fileToken = strings.TrimSpace(fileToken)
	if fileToken == "" {
		return nil, errors.New("missing file token")
	}

	claims, _ := security.ParseUnverifiedJWT(strings.TrimSpace(fileToken))
	tokenType := cast.ToString(claims["type"])

	switch tokenType {
	case tokens.TypeAuthRecord:
		record, err := s.Dao().FindUserRecordByToken(
			fileToken,
			s.App().Settings().RecordFileToken.Secret,
		)
		if err == nil && record != nil {
			return record, nil
		}
	}

	return nil, errors.New("missing or invalid file token")
}
