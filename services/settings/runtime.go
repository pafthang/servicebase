package settings

import (
	baseforms "github.com/pafthang/servicebase/services/base/forms"
	settingsforms "github.com/pafthang/servicebase/services/settings/forms"
	settingsmodels "github.com/pafthang/servicebase/services/settings/models"
)

func (s *Service) RedactClone() (*settingsmodels.Settings, error) {
	return s.App().Settings().RedactClone()
}

func (s *Service) NewUpsertForm() *settingsforms.SettingsUpsert {
	return settingsforms.NewSettingsUpsert(s.App())
}

func (s *Service) SubmitUpsert(
	form *settingsforms.SettingsUpsert,
	interceptors ...baseforms.InterceptorFunc[*settingsmodels.Settings],
) error {
	return form.Submit(interceptors...)
}

func (s *Service) NewTestS3Form() *settingsforms.TestS3Filesystem {
	return settingsforms.NewTestS3Filesystem(s.App())
}

func (s *Service) TestS3(form *settingsforms.TestS3Filesystem) error {
	return form.Submit()
}

func (s *Service) NewTestEmailForm() *settingsforms.TestEmailSend {
	return settingsforms.NewTestEmailSend(s.App())
}

func (s *Service) TestEmail(form *settingsforms.TestEmailSend) error {
	return form.Submit()
}
