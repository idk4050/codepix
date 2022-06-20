package transaction

import (
	"codepix/bank-api/lib/aggregates"

	eh "github.com/looplab/eventhorizon"
	"github.com/looplab/eventhorizon/uuid"
)

var (
	ErrAlreadyStarted = &aggregates.StatusMismatchError{
		"transaction already started",
	}
	ErrCannotConfirmIfNotStarted = &aggregates.StatusMismatchError{
		"cannot confirm transaction if not started",
	}
	ErrCannotCompleteIfNotConfirmed = &aggregates.StatusMismatchError{
		"cannot complete transaction if not confirmed",
	}
	ErrCannotFailIfNotStartedOrConfirmed = &aggregates.StatusMismatchError{
		"cannot fail transaction if not started or confirmed",
	}

	ErrCannotStartIfNotTheSender = &aggregates.PermissionError{
		"cannot start transaction if not the sender",
	}
	ErrCannotConfirmIfNotTheReceiver = &aggregates.PermissionError{
		"cannot confirm transaction if not the receiver",
	}
	ErrCannotCompleteIfNotTheSender = &aggregates.PermissionError{
		"cannot complete transaction if not the sender",
	}
	ErrCannotFailIfNotSenderOrReceiver = &aggregates.PermissionError{
		"cannot fail transaction if not the sender or receiver",
	}
)

type Command interface {
	ToEvent(ag Aggregate) (Event, error)
}

type Start struct {
	ID           uuid.UUID
	BankID       uuid.UUID
	Sender       uuid.UUID
	SenderBank   uuid.UUID
	Receiver     uuid.UUID
	ReceiverBank uuid.UUID
	Amount       Amount
	Description  string `eh:"optional"`
}

func (c Start) ToEvent(ag Aggregate) (Event, error) {
	if ag.Transaction.Status != 0 {
		return nil, ErrAlreadyStarted
	}
	if c.BankID != c.SenderBank {
		return nil, ErrCannotStartIfNotTheSender
	}
	return TransactionStarted{
		Sender:       c.Sender,
		SenderBank:   c.SenderBank,
		Receiver:     c.Receiver,
		ReceiverBank: c.ReceiverBank,
		Amount:       c.Amount,
		Description:  c.Description,
	}, nil
}

type Confirm struct {
	ID     uuid.UUID
	BankID uuid.UUID
}

func (c Confirm) ToEvent(ag Aggregate) (Event, error) {
	if ag.Transaction.Status != Started {
		return nil, ErrCannotConfirmIfNotStarted
	}
	if c.BankID != ag.Transaction.ReceiverBank {
		return nil, ErrCannotConfirmIfNotTheReceiver
	}
	return TransactionConfirmed{
		SenderBank:   ag.Transaction.SenderBank,
		ReceiverBank: ag.Transaction.ReceiverBank,
	}, nil
}

type Complete struct {
	ID     uuid.UUID
	BankID uuid.UUID
}

func (c Complete) ToEvent(ag Aggregate) (Event, error) {
	if ag.Transaction.Status != Confirmed {
		return nil, ErrCannotCompleteIfNotConfirmed
	}
	if c.BankID != ag.Transaction.SenderBank {
		return nil, ErrCannotCompleteIfNotTheSender
	}
	return TransactionCompleted{
		SenderBank:   ag.Transaction.SenderBank,
		ReceiverBank: ag.Transaction.ReceiverBank,
	}, nil
}

type Fail struct {
	ID     uuid.UUID
	BankID uuid.UUID
	Reason string `eh:"optional"`
}

func (c Fail) ToEvent(ag Aggregate) (Event, error) {
	if !(ag.Transaction.Status == Started || ag.Transaction.Status == Confirmed) {
		return nil, ErrCannotFailIfNotStartedOrConfirmed
	}
	if !(c.BankID == ag.Transaction.SenderBank || c.BankID == ag.Transaction.ReceiverBank) {
		return nil, ErrCannotFailIfNotSenderOrReceiver
	}
	return TransactionFailed{
		SenderBank:   ag.Transaction.SenderBank,
		ReceiverBank: ag.Transaction.ReceiverBank,
		Reason:       c.Reason,
	}, nil
}

const (
	StartCommand    = eh.CommandType(AggregateType + "_start")
	ConfirmCommand  = eh.CommandType(AggregateType + "_confirm")
	CompleteCommand = eh.CommandType(AggregateType + "_complete")
	FailCommand     = eh.CommandType(AggregateType + "_fail")
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
