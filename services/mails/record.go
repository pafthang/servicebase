package mails

import (
	"html/template"
	"net/mail"

	"github.com/pafthang/servicebase/core"
	"github.com/pafthang/servicebase/services/mails/templates"
	recordmodels "github.com/pafthang/servicebase/services/record/models"
	settingsmodels "github.com/pafthang/servicebase/services/settings/models"
	"github.com/pafthang/servicebase/tools/mailer"
	"github.com/pafthang/servicebase/tools/tokens"
)

// SendRecordPasswordLoginAlert sends an OAuth2 password login alert to the specified auth record.
func SendRecordPasswordLoginAlert(app core.App, authRecord *recordmodels.Record, providerNames ...string) error {
	return New(app).SendRecordPasswordLoginAlert(authRecord, providerNames...)
}

// SendRecordPasswordLoginAlert sends an OAuth2 password login alert to the specified auth record.
func (s *Service) SendRecordPasswordLoginAlert(authRecord *recordmodels.Record, providerNames ...string) error {
	params := struct {
		AppName       string
		AppUrl        string
		Record        *recordmodels.Record
		ProviderNames []string
	}{
		AppName:       s.app.Settings().Meta.AppName,
		AppUrl:        s.app.Settings().Meta.AppUrl,
		Record:        authRecord,
		ProviderNames: providerNames,
	}

	body, renderErr := resolveTemplateContent(params, templates.Layout, templates.PasswordLoginAlertBody)
	if renderErr != nil {
		return renderErr
	}

	message := &mailer.Message{
		From: mail.Address{
			Name:    s.app.Settings().Meta.SenderName,
			Address: s.app.Settings().Meta.SenderAddress,
		},
		To:      []mail.Address{{Address: authRecord.Email()}},
		Subject: "Password login alert",
		HTML:    body,
	}

	return s.app.NewMailClient().Send(message)
}

// SendRecordPasswordReset sends a password reset request email to the specified user.
func SendRecordPasswordReset(app core.App, authRecord *recordmodels.Record) error {
	return New(app).SendRecordPasswordReset(authRecord)
}

// SendRecordPasswordReset sends a password reset request email to the specified user.
func (s *Service) SendRecordPasswordReset(authRecord *recordmodels.Record) error {
	token, tokenErr := tokens.NewRecordResetPasswordToken(s.app, authRecord)
	if tokenErr != nil {
		return tokenErr
	}

	mailClient := s.app.NewMailClient()

	subject, body, err := resolveEmailTemplate(s.app, token, s.app.Settings().Meta.ResetPasswordTemplate)
	if err != nil {
		return err
	}

	message := &mailer.Message{
		From: mail.Address{
			Name:    s.app.Settings().Meta.SenderName,
			Address: s.app.Settings().Meta.SenderAddress,
		},
		To:      []mail.Address{{Address: authRecord.Email()}},
		Subject: subject,
		HTML:    body,
	}

	event := &core.MailerRecordEvent{
		BaseCollectionEvent: core.BaseCollectionEvent{Collection: authRecord.Collection()},
		MailClient:          mailClient,
		Message:             message,
		Record:              authRecord,
		Meta:                map[string]any{"token": token},
	}

	return s.app.OnMailerBeforeRecordResetPasswordSend().Trigger(event, func(e *core.MailerRecordEvent) error {
		if err := e.MailClient.Send(e.Message); err != nil {
			return err
		}

		return s.app.OnMailerAfterRecordResetPasswordSend().Trigger(e)
	})
}

// SendRecordVerification sends a verification request email to the specified user.
func SendRecordVerification(app core.App, authRecord *recordmodels.Record) error {
	return New(app).SendRecordVerification(authRecord)
}

// SendRecordVerification sends a verification request email to the specified user.
func (s *Service) SendRecordVerification(authRecord *recordmodels.Record) error {
	token, tokenErr := tokens.NewRecordVerifyToken(s.app, authRecord)
	if tokenErr != nil {
		return tokenErr
	}

	mailClient := s.app.NewMailClient()

	subject, body, err := resolveEmailTemplate(s.app, token, s.app.Settings().Meta.VerificationTemplate)
	if err != nil {
		return err
	}

	message := &mailer.Message{
		From: mail.Address{
			Name:    s.app.Settings().Meta.SenderName,
			Address: s.app.Settings().Meta.SenderAddress,
		},
		To:      []mail.Address{{Address: authRecord.Email()}},
		Subject: subject,
		HTML:    body,
	}

	event := &core.MailerRecordEvent{
		BaseCollectionEvent: core.BaseCollectionEvent{Collection: authRecord.Collection()},
		MailClient:          mailClient,
		Message:             message,
		Record:              authRecord,
		Meta:                map[string]any{"token": token},
	}

	return s.app.OnMailerBeforeRecordVerificationSend().Trigger(event, func(e *core.MailerRecordEvent) error {
		if err := e.MailClient.Send(e.Message); err != nil {
			return err
		}

		return s.app.OnMailerAfterRecordVerificationSend().Trigger(e)
	})
}

// SendRecordChangeEmail sends a change email confirmation email to the specified user.
func SendRecordChangeEmail(app core.App, record *recordmodels.Record, newEmail string) error {
	return New(app).SendRecordChangeEmail(record, newEmail)
}

// SendRecordChangeEmail sends a change email confirmation email to the specified user.
func (s *Service) SendRecordChangeEmail(record *recordmodels.Record, newEmail string) error {
	token, tokenErr := tokens.NewRecordChangeEmailToken(s.app, record, newEmail)
	if tokenErr != nil {
		return tokenErr
	}

	mailClient := s.app.NewMailClient()

	subject, body, err := resolveEmailTemplate(s.app, token, s.app.Settings().Meta.ConfirmEmailChangeTemplate)
	if err != nil {
		return err
	}

	message := &mailer.Message{
		From: mail.Address{
			Name:    s.app.Settings().Meta.SenderName,
			Address: s.app.Settings().Meta.SenderAddress,
		},
		To:      []mail.Address{{Address: newEmail}},
		Subject: subject,
		HTML:    body,
	}

	event := &core.MailerRecordEvent{
		BaseCollectionEvent: core.BaseCollectionEvent{Collection: record.Collection()},
		MailClient:          mailClient,
		Message:             message,
		Record:              record,
		Meta: map[string]any{
			"token":    token,
			"newEmail": newEmail,
		},
	}

	return s.app.OnMailerBeforeRecordChangeEmailSend().Trigger(event, func(e *core.MailerRecordEvent) error {
		if err := e.MailClient.Send(e.Message); err != nil {
			return err
		}

		return s.app.OnMailerAfterRecordChangeEmailSend().Trigger(e)
	})
}

func resolveEmailTemplate(
	app core.App,
	token string,
	emailTemplate settingsmodels.EmailTemplate,
) (subject string, body string, err error) {
	subject, rawBody, _ := emailTemplate.Resolve(
		app.Settings().Meta.AppName,
		app.Settings().Meta.AppUrl,
		token,
	)

	params := struct {
		HtmlContent template.HTML
	}{
		HtmlContent: template.HTML(rawBody),
	}

	body, err = resolveTemplateContent(params, templates.Layout, templates.HtmlBody)
	if err != nil {
		return "", "", err
	}

	return subject, body, nil
}
