package transaction

import (
	"github.com/google/uuid"
	eh "github.com/looplab/eventhorizon"
)

type Event interface {
	Apply(tx *Transaction)
}

type TransactionStarted struct {
	Sender       uuid.UUID
	SenderBank   uuid.UUID
	Receiver     uuid.UUID
	ReceiverBank uuid.UUID
	Amount       Amount
	Description  string
}

func (e TransactionStarted) Apply(tx *Transaction) {
	tx.Sender = e.Sender
	tx.SenderBank = e.SenderBank
	tx.Receiver = e.Receiver
	tx.ReceiverBank = e.ReceiverBank
	tx.Amount = e.Amount
	tx.Description = e.Description
	tx.Status = Started
}

type TransactionConfirmed struct {
	SenderBank uuid.UUID
}

func (e TransactionConfirmed) Apply(tx *Transaction) {
	tx.Status = Confirmed
}

type TransactionCompleted struct {
}

func (e TransactionCompleted) Apply(tx *Transaction) {
	tx.Status = Completed
}

type TransactionFailed struct {
	Reason string
}

func (e TransactionFailed) Apply(tx *Transaction) {
	tx.Status = Failed
}

const (
	StartedEvent   = eh.EventType(AggregateType + ":started")
	ConfirmedEvent = eh.EventType(AggregateType + ":confirmed")
	CompletedEvent = eh.EventType(AggregateType + ":completed")
	FailedEvent    = eh.EventType(AggregateType + ":failed")
)

func init() {
	eh.RegisterEventData(StartedEvent, func() eh.EventData { return &TransactionStarted{} })
	eh.RegisterEventData(ConfirmedEvent, func() eh.EventData { return &TransactionConfirmed{} })
	eh.RegisterEventData(CompletedEvent, func() eh.EventData { return &TransactionCompleted{} })
	eh.RegisterEventData(FailedEvent, func() eh.EventData { return &TransactionFailed{} })
}
