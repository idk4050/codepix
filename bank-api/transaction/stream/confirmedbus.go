package stream

import (
	"bytes"
	"codepix/bank-api/transaction"
	"context"

	"github.com/google/uuid"
	"github.com/looplab/eventhorizon"
)

type ConfirmedBus struct {
	Outbox eventhorizon.Outbox
}

type ConfirmedHandler func(transaction.TransactionConfirmed, uuid.UUID) error

func (b ConfirmedBus) AddSenderHandler(bankID uuid.UUID, handler ConfirmedHandler) error {
	var handlerWrapper eventhorizon.EventHandlerFunc = func(
		ctx context.Context, event eventhorizon.Event,
	) error {
		confirmed := event.Data().(transaction.TransactionConfirmed)
		return handler(confirmed, event.AggregateID())
	}
	return b.Outbox.AddHandler(
		context.Background(),
		SenderMatcher{bankID},
		handlerWrapper,
	)
}

type SenderMatcher struct {
	BankID uuid.UUID
}

func (m SenderMatcher) Match(event eventhorizon.Event) bool {
	confirmed, ok := event.Data().(transaction.TransactionConfirmed)
	if !ok {
		return false
	}
	return bytes.Equal(m.BankID[:], confirmed.SenderBank[:])
}
