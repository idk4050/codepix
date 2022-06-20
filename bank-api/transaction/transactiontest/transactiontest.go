package transactiontest

import (
	"codepix/bank-api/pixkey"
	"codepix/bank-api/pixkey/pixkeytest"
	pixkeyrepository "codepix/bank-api/pixkey/repository"
	proto "codepix/bank-api/proto/codepix/transaction/write"
	"codepix/bank-api/transaction"
	readrepository "codepix/bank-api/transaction/read/repository"
	"context"
	"reflect"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/looplab/eventhorizon"
)

var Copy = func(ag *transaction.Aggregate) *transaction.Aggregate {
	aggregate := transaction.New(ag.EntityID())
	ctx := context.Background()
	for _, event := range ag.UncommittedEvents() {
		aggregate.AppendEvent(event.EventType(), event.Data(), event.Timestamp())
		aggregate.ApplyEvent(ctx, event)
	}
	return aggregate
}

var Comparer = cmp.Comparer(func(x, y *transaction.Aggregate) bool {
	xID, yID := x.EntityID(), y.EntityID()
	xEvents, yEvents := x.UncommittedEvents(), y.UncommittedEvents()
	return x.AggregateType() == y.AggregateType() &&
		x.AggregateVersion() == y.AggregateVersion() &&
		xID == yID &&
		cmp.Equal(x.Transaction, y.Transaction) &&
		EqualEvents(xEvents, yEvents)
})

var EqualEvents = func(xs, ys []eventhorizon.Event) bool {
	if len(xs) != len(ys) {
		return false
	}
	for i, x := range xs {
		y := ys[i]
		xID, yID := x.AggregateID(), y.AggregateID()

		if !(x.AggregateType() == y.AggregateType() &&
			x.EventType() == y.EventType() &&
			x.Version() == y.Version() &&
			xID == yID &&
			reflect.DeepEqual(x.Data(), y.Data()) &&
			reflect.DeepEqual(x.Metadata(), y.Metadata())) {
			return false
		}
	}
	return true
}

var ExceptStatus = cmp.FilterPath(func(p cmp.Path) bool {
	return p.String() == "Transaction.Status"
}, cmp.Ignore())

var ExceptID = cmp.FilterPath(func(p cmp.Path) bool {
	return p.String() == "ID"
}, cmp.Ignore())

func ValidStartCommand(ID uuid.UUID) transaction.Start {
	return transaction.Start{
		ID:           ID,
		BankID:       uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		Sender:       uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		SenderBank:   uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		Receiver:     uuid.MustParse("33333333-3333-3333-3333-333333333333"),
		ReceiverBank: uuid.MustParse("44444444-4444-4444-4444-444444444444"),
		Amount:       100,
		Description:  "test",
	}
}
func ValidConfirmCommand(ID uuid.UUID) transaction.Confirm {
	return transaction.Confirm{
		ID:     ID,
		BankID: uuid.MustParse("44444444-4444-4444-4444-444444444444"),
	}
}
func ValidCompleteCommand(ID uuid.UUID) transaction.Complete {
	return transaction.Complete{
		ID:     ID,
		BankID: uuid.MustParse("22222222-2222-2222-2222-222222222222"),
	}
}
func ValidFailCommand(ID uuid.UUID) transaction.Fail {
	return transaction.Fail{
		ID:     ID,
		BankID: uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		Reason: "not enough balance",
	}
}

func StartedTransaction(ID uuid.UUID) *transaction.Aggregate {
	ag := transaction.New(ID)
	cmd := ValidStartCommand(ag.EntityID())
	ctx := context.Background()

	ag.HandleCommand(ctx, cmd)
	lastEvent := ag.UncommittedEvents()[len(ag.UncommittedEvents())-1]
	ag.ApplyEvent(ctx, lastEvent)
	return ag
}
func ConfirmedTransaction(ID uuid.UUID) *transaction.Aggregate {
	ag := StartedTransaction(ID)
	cmd := ValidConfirmCommand(ag.EntityID())
	ctx := context.Background()

	ag.HandleCommand(ctx, cmd)
	lastEvent := ag.UncommittedEvents()[len(ag.UncommittedEvents())-1]
	ag.ApplyEvent(ctx, lastEvent)
	return ag
}
func CompletedTransaction(ID uuid.UUID) *transaction.Aggregate {
	ag := ConfirmedTransaction(ID)
	cmd := ValidCompleteCommand(ag.EntityID())
	ctx := context.Background()

	ag.HandleCommand(ctx, cmd)
	lastEvent := ag.UncommittedEvents()[len(ag.UncommittedEvents())-1]
	ag.ApplyEvent(ctx, lastEvent)
	return ag
}
func FailedTransaction(ID uuid.UUID) *transaction.Aggregate {
	ag := ConfirmedTransaction(ID)
	cmd := ValidFailCommand(ag.EntityID())
	ctx := context.Background()

	ag.HandleCommand(ctx, cmd)
	lastEvent := ag.UncommittedEvents()[len(ag.UncommittedEvents())-1]
	ag.ApplyEvent(ctx, lastEvent)
	return ag
}

type TearDown = func()

type SenderIDs struct {
	AccountID uuid.UUID
	BankID    uuid.UUID
}

type txCreator = func() (
	ID uuid.UUID, sender SenderIDs, receiver pixkeyrepository.IDs,
)

type Creator struct {
	ReceiverIDs  func(pixkey.PixKey) pixkeyrepository.IDs
	StartedID    func(SenderIDs, pixkeyrepository.IDs) uuid.UUID
	StartedIDs   txCreator
	ConfirmedIDs txCreator
	CompletedIDs txCreator
	FailedIDs    txCreator
}

func StartedID(commandHandler eventhorizon.CommandHandler) func(SenderIDs, pixkeyrepository.IDs) uuid.UUID {
	return func(senderIDs SenderIDs, receiverIDs pixkeyrepository.IDs) uuid.UUID {
		ID := uuid.New()
		start := ValidStartCommand(ID)
		start.BankID = senderIDs.BankID
		start.Sender = senderIDs.AccountID
		start.SenderBank = senderIDs.BankID
		start.Receiver = receiverIDs.AccountID
		start.ReceiverBank = receiverIDs.BankID

		commandHandler.HandleCommand(context.Background(), start)
		return ID
	}
}

func StartedIDs(commandHandler eventhorizon.CommandHandler,
	receiverIDs func(pixkey.PixKey) pixkeyrepository.IDs,
) txCreator {
	return func() (uuid.UUID, SenderIDs, pixkeyrepository.IDs) {
		senderIDs := SenderIDs{
			AccountID: uuid.New(),
			BankID:    uuid.New(),
		}
		receiverIDs := receiverIDs(pixkeytest.ValidPixKey())

		ID := uuid.New()
		start := ValidStartCommand(ID)
		start.BankID = senderIDs.BankID
		start.Sender = senderIDs.AccountID
		start.SenderBank = senderIDs.BankID
		start.Receiver = receiverIDs.AccountID
		start.ReceiverBank = receiverIDs.BankID

		commandHandler.HandleCommand(context.Background(), start)
		return ID, senderIDs, receiverIDs
	}
}

func ConfirmedIDs(commandHandler eventhorizon.CommandHandler, startedIDs txCreator) txCreator {
	return func() (uuid.UUID, SenderIDs, pixkeyrepository.IDs) {
		ID, senderIDs, receiverIDs := startedIDs()
		confirm := ValidConfirmCommand(ID)
		confirm.BankID = receiverIDs.BankID
		commandHandler.HandleCommand(context.Background(), confirm)
		return ID, senderIDs, receiverIDs
	}
}

func CompletedIDs(commandHandler eventhorizon.CommandHandler, confirmedIDs txCreator) txCreator {
	return func() (uuid.UUID, SenderIDs, pixkeyrepository.IDs) {
		ID, senderIDs, receiverIDs := confirmedIDs()
		complete := ValidCompleteCommand(ID)
		complete.BankID = senderIDs.BankID
		commandHandler.HandleCommand(context.Background(), complete)
		return ID, senderIDs, receiverIDs
	}
}

func FailedIDs(commandHandler eventhorizon.CommandHandler, confirmedIDs txCreator) txCreator {
	return func() (uuid.UUID, SenderIDs, pixkeyrepository.IDs) {
		ID, senderIDs, receiverIDs := confirmedIDs()
		fail := ValidFailCommand(ID)
		fail.BankID = senderIDs.BankID
		commandHandler.HandleCommand(context.Background(), fail)
		return ID, senderIDs, receiverIDs
	}
}

func ValidTransaction() *readrepository.Transaction {
	cmd := ValidStartCommand(uuid.New())
	now := time.Now()
	return &readrepository.Transaction{
		ID:           cmd.ID,
		Sender:       cmd.Sender,
		SenderBank:   cmd.SenderBank,
		Receiver:     cmd.Receiver,
		ReceiverBank: cmd.ReceiverBank,

		CreatedAt:        now,
		UpdatedAt:        now,
		Amount:           cmd.Amount,
		Description:      cmd.Description,
		Status:           transaction.Started,
		ReasonForFailing: "",
	}
}

func ValidStartRequest() *proto.StartRequest {
	validCommand := ValidStartCommand(uuid.New())
	return &proto.StartRequest{
		SenderId:    validCommand.Sender[:],
		ReceiverKey: pixkeytest.ValidPixKey().Key,
		Amount:      validCommand.Amount,
		Description: validCommand.Description,
	}
}
func InvalidStartRequest() *proto.StartRequest {
	return &proto.StartRequest{
		Description: strings.Repeat("A", 101),
	}
}

func InvalidConfirmRequest() *proto.ConfirmRequest {
	return &proto.ConfirmRequest{}
}

func InvalidCompleteRequest() *proto.CompleteRequest {
	return &proto.CompleteRequest{}
}

func InvalidFailRequest() *proto.FailRequest {
	return &proto.FailRequest{
		Reason: strings.Repeat("A", 101),
	}
}
