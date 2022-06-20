package transaction

import (
	"github.com/google/uuid"
	eh "github.com/looplab/eventhorizon"
)

type Event interface {
	Apply(tx *Transaction)
	Type() eh.EventType
}

type TransactionStarted struct {
	Sender       uuid.UUID `json:"sender" bson:"sender"`
	SenderBank   uuid.UUID `json:"sender_bank" bson:"sender_bank"`
	Receiver     uuid.UUID `json:"receiver" bson:"receiver"`
	ReceiverBank uuid.UUID `json:"receiver_bank" bson:"receiver_bank"`
	Amount       Amount    `json:"amount" bson:"amount"`
	Description  string    `json:"description" bson:"description"`
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
	SenderBank   uuid.UUID `json:"sender_bank" bson:"sender_bank"`
	ReceiverBank uuid.UUID `json:"receiver_bank" bson:"receiver_bank"`
}

func (e TransactionConfirmed) Apply(tx *Transaction) {
	tx.Status = Confirmed
}

type TransactionCompleted struct {
	SenderBank   uuid.UUID `json:"sender_bank" bson:"sender_bank"`
	ReceiverBank uuid.UUID `json:"receiver_bank" bson:"receiver_bank"`
}

func (e TransactionCompleted) Apply(tx *Transaction) {
	tx.Status = Completed
}

type TransactionFailed struct {
	SenderBank   uuid.UUID `json:"sender_bank" bson:"sender_bank"`
	ReceiverBank uuid.UUID `json:"receiver_bank" bson:"receiver_bank"`
	Reason       string    `json:"reason" bson:"reason"`
}

func (e TransactionFailed) Apply(tx *Transaction) {
	tx.Status = Failed
}

const (
	StartedEvent   = eh.EventType(AggregateType + "_started")
	ConfirmedEvent = eh.EventType(AggregateType + "_confirmed")
	CompletedEvent = eh.EventType(AggregateType + "_completed")
	FailedEvent    = eh.EventType(AggregateType + "_failed")
)

func init() {
	eh.RegisterEventData(StartedEvent, func() eh.EventData { return &TransactionStarted{} })
	eh.RegisterEventData(ConfirmedEvent, func() eh.EventData { return &TransactionConfirmed{} })
	eh.RegisterEventData(CompletedEvent, func() eh.EventData { return &TransactionCompleted{} })
	eh.RegisterEventData(FailedEvent, func() eh.EventData { return &TransactionFailed{} })
}

func (TransactionStarted) Type() eh.EventType   { return StartedEvent }
func (TransactionConfirmed) Type() eh.EventType { return ConfirmedEvent }
func (TransactionCompleted) Type() eh.EventType { return CompletedEvent }
func (TransactionFailed) Type() eh.EventType    { return FailedEvent }
