package emailtest

import (
	"context"
	"sync"

	"nebula-exchange/backend/internal/notify/email"
)

type RecordingSender struct {
	mutex        sync.Mutex
	sentMessages []email.Message
	failure      error
}

func NewRecordingSender() *RecordingSender {
	return &RecordingSender{}
}

func (sender *RecordingSender) FailWith(failure error) {
	sender.mutex.Lock()
	defer sender.mutex.Unlock()
	sender.failure = failure
}

func (sender *RecordingSender) Send(_ context.Context, message email.Message) error {
	sender.mutex.Lock()
	defer sender.mutex.Unlock()
	if sender.failure != nil {
		return sender.failure
	}
	sender.sentMessages = append(sender.sentMessages, message)
	return nil
}

func (sender *RecordingSender) SentMessages() []email.Message {
	sender.mutex.Lock()
	defer sender.mutex.Unlock()
	return append([]email.Message(nil), sender.sentMessages...)
}
