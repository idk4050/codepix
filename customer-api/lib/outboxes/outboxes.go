package outboxes

import (
	"context"
	"fmt"
	"time"
)

type Outbox interface {
	AutoMigrate() error
	Start(context.Context)
	Write(tx interface{}, message NewMessage) error
}

type Namespace = string

type NewMessage interface {
	Namespace() Namespace
	Type() string
}

type Message struct {
	ID      string
	Type    string
	Payload []byte
}

type UnknownMessageTypeError struct {
	Message Message
	When    time.Time
}

func (e *UnknownMessageTypeError) Error() string {
	return fmt.Sprintf("message with unknown type %T received", e.Message.Type)
}
func NewUnknownMessageTypeError(message Message) *UnknownMessageTypeError {
	return &UnknownMessageTypeError{message, time.Now()}
}
