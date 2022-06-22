package publishers

import (
	"codepix/customer-api/lib/outboxes"
	"encoding/json"
)

type Publisher interface {
	Publish(message outboxes.Message) error
}

type PublishError struct {
	Err       error
	Message   outboxes.Message
	Namespace string
}

func NewPublishError(err error, message outboxes.Message, namespace string) *PublishError {
	return &PublishError{err, message, namespace}
}

func (e *PublishError) Error() string {
	return e.Err.Error()
}

type SkipMessage struct{}

func (*SkipMessage) Error() string { return "message skipped" }

func Payload[T any](message outboxes.Message) (T, error) {
	var payload T
	err := json.Unmarshal(message.Payload, &payload)
	if err != nil {
		return *new(T), err
	}
	return payload, err
}
