package mails

import (
	"bytes"
	"html/template"
	texttpl "text/template"

	"github.com/pafthang/servicebase/tools/mailer"
)

type Message struct {
	*mailer.Message
}

func NewMessage() *Message {
	return &Message{
		Message: &mailer.Message{},
	}
}

func (m *Message) RenderHTML(raw string, data any) error {
	tpl, err := template.New("html").Parse(raw)
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	if err := tpl.Execute(&buf, data); err != nil {
		return err
	}

	m.HTML = buf.String()

	return nil
}

func (m *Message) RenderText(raw string, data any) error {
	tpl, err := texttpl.New("text").Parse(raw)
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	if err := tpl.Execute(&buf, data); err != nil {
		return err
	}

	m.Text = buf.String()

	return nil
}
