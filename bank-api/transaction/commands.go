package transaction

import (
	"errors"

	eh "github.com/looplab/eventhorizon"
	"github.com/looplab/eventhorizon/uuid"
)

var (
	ErrAlreadyStarted            = errors.New("transaction already started")
	ErrCannotConfirmIfNotStarted = errors.New(
		"cannot confirm transaction if not started")
	ErrCannotFailIfNotStartedOrConfirmed = errors.New(
		"cannot fail transaction if not started or confirmed")
	ErrCannotCompleteIfNotConfirmed = errors.New(
		"cannot complete transaction if not confirmed")
)

type Command interface {
	ToEvent(ag Aggregate) (Event, eh.EventType, error)
}

type Start struct {
	ID           uuid.UUID
	Sender       uuid.UUID
	SenderBank   uuid.UUID
	Receiver     uuid.UUID
	ReceiverBank uuid.UUID
	Amount       Amount
	Description  string
}

func (c Start) ToEvent(ag Aggregate) (Event, eh.EventType, error) {
	if ag.Transaction.Status != 0 {
		return nil, "", ErrAlreadyStarted
	}
	return TransactionStarted{
		Sender:       c.Sender,
		SenderBank:   c.SenderBank,
		Receiver:     c.Receiver,
		ReceiverBank: c.ReceiverBank,
		Amount:       c.Amount,
		Description:  c.Description,
	}, StartedEvent, nil
}

type Confirm struct {
	ID         uuid.UUID
	SenderBank uuid.UUID
}

func (c Confirm) ToEvent(ag Aggregate) (Event, eh.EventType, error) {
	if ag.Transaction.Status != Started {
		return nil, "", ErrCannotConfirmIfNotStarted
	}
	return TransactionConfirmed{
		SenderBank: c.SenderBank,
	}, ConfirmedEvent, nil
}

type Complete struct {
	ID uuid.UUID
}

func (c Complete) ToEvent(ag Aggregate) (Event, eh.EventType, error) {
	if ag.Transaction.Status != Confirmed {
		return nil, "", ErrCannotCompleteIfNotConfirmed
	}
	return TransactionCompleted{}, CompletedEvent, nil
}

type Fail struct {
	ID     uuid.UUID
	Reason string
}

func (c Fail) ToEvent(ag Aggregate) (Event, eh.EventType, error) {
	if !(ag.Transaction.Status == Started || ag.Transaction.Status == Confirmed) {
		return nil, "", ErrCannotFailIfNotStartedOrConfirmed
	}
	return TransactionFailed{
		Reason: c.Reason,
	}, FailedEvent, nil
}

const (
	StartCommand    = eh.CommandType(AggregateType + ":start")
	ConfirmCommand  = eh.CommandType(AggregateType + ":confirm")
	CompleteCommand = eh.CommandType(AggregateType + ":complete")
	FailCommand     = eh.CommandType(AggregateType + ":fail")
)

func init() {
	eh.RegisterCommand(func() eh.Command { return Start{} })
	eh.RegisterCommand(func() eh.Command { return Confirm{} })
	eh.RegisterCommand(func() eh.Command { return Complete{} })
	eh.RegisterCommand(func() eh.Command { return Fail{} })
}

func (c Start) AggregateID() uuid.UUID          { return c.ID }
func (c Start) AggregateType() eh.AggregateType { return AggregateType }
func (c Start) CommandType() eh.CommandType     { return StartCommand }

func (c Confirm) AggregateID() uuid.UUID          { return c.ID }
func (c Confirm) AggregateType() eh.AggregateType { return AggregateType }
func (c Confirm) CommandType() eh.CommandType     { return ConfirmCommand }

func (c Complete) AggregateID() uuid.UUID          { return c.ID }
func (c Complete) AggregateType() eh.AggregateType { return AggregateType }
func (c Complete) CommandType() eh.CommandType     { return CompleteCommand }

func (c Fail) AggregateID() uuid.UUID          { return c.ID }
func (c Fail) AggregateType() eh.AggregateType { return AggregateType }
func (c Fail) CommandType() eh.CommandType     { return FailCommand }

var (
	_ Command    = Start{}
	_ Command    = Confirm{}
	_ Command    = Complete{}
	_ Command    = Fail{}
	_ eh.Command = Start{}
	_ eh.Command = Confirm{}
	_ eh.Command = Complete{}
	_ eh.Command = Fail{}
)
