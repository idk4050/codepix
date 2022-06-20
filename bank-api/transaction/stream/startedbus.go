package stream

import (
	"bytes"
	"codepix/bank-api/transaction"
	"context"

	"github.com/google/uuid"
	"github.com/looplab/eventhorizon"
)

type StartedBus struct {
	Outbox eventhorizon.Outbox
}

type StartedHandler func(transaction.TransactionStarted, uuid.UUID) error

func (b StartedBus) AddReceiverHandler(bankID uuid.UUID, handler StartedHandler) error {
	var handlerWrapper eventhorizon.EventHandlerFunc = func(
		ctx context.Context, event eventhorizon.Event,
	) error {
		started := event.Data().(transaction.TransactionStarted)
		return handler(started, event.AggregateID())
	}
	return b.Outbox.AddHandler(
		context.Background(),
		ReceiverMatcher{bankID},
		handlerWrapper,
	)
}

type ReceiverMatcher struct {
	BankID uuid.UUID
}

func (m ReceiverMatcher) Match(event eventhorizon.Event) bool {
	started, ok := event.Data().(transaction.TransactionStarted)
	if !ok {
		return false
	}
	return bytes.Equal(m.BankID[:], started.ReceiverBank[:])
}
