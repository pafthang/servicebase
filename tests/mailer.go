package tests

import (
	"sync"

	"github.com/pafthang/servicebase/tools/mailer"
)

var _ mailer.Mailer = (*TestMailer)(nil)

// TestMailer is a mock `mailer.Mailer` implementation.
type TestMailer struct {
	mux sync.Mutex

	TotalSend   int
	LastMessage mailer.Message

	SentMessages []mailer.Message
}

// Reset clears any previously test collected data.
func (tm *TestMailer) Reset() {
	tm.mux.Lock()
	defer tm.mux.Unlock()

	tm.TotalSend = 0
	tm.LastMessage = mailer.Message{}
	tm.SentMessages = nil
}

// Send implements `mailer.Mailer` interface.
func (tm *TestMailer) Send(m *mailer.Message) error {
	tm.mux.Lock()
	defer tm.mux.Unlock()

	tm.TotalSend++
	tm.LastMessage = *m
	tm.SentMessages = append(tm.SentMessages, tm.LastMessage)

	return nil
}
