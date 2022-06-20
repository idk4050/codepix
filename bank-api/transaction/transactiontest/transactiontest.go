package transactiontest

import (
	"bytes"
	"codepix/bank-api/adapters/commandbus"
	"codepix/bank-api/adapters/eventstore"
	"codepix/bank-api/adapters/projectionclient"
	"codepix/bank-api/bankapitest"
	"codepix/bank-api/transaction"
	"codepix/bank-api/transaction/commandhandler"
	"codepix/bank-api/transaction/readrepository"
	"codepix/bank-api/transaction/readrepository/projection"
	"context"
	"reflect"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/looplab/eventhorizon"
	"github.com/looplab/eventhorizon/commandhandler/bus"
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
		bytes.Equal(xID[:], yID[:]) &&
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
			bytes.Equal(xID[:], yID[:]) &&
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

func ValidStartCommand(ID uuid.UUID) transaction.Start {
	return transaction.Start{
		ID:           ID,
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
		ID:         ID,
		SenderBank: uuid.MustParse("22222222-2222-2222-2222-222222222222"),
	}
}
func ValidCompleteCommand(ID uuid.UUID) transaction.Complete {
	return transaction.Complete{
		ID: ID,
	}
}
func ValidFailCommand(ID uuid.UUID) transaction.Fail {
	return transaction.Fail{
		ID:     ID,
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

type StoreTearDown func()

func CommandHandler() (eventhorizon.CommandHandler, eventhorizon.Outbox, StoreTearDown) {
	eventStore, err := eventstore.Open(bankapitest.Config, bankapitest.Logger)
	if err != nil {
		panic(err)
	}
	tearDown := func() {
		err := eventStore.OnDisconnect()
		if err != nil {
			panic(err)
		}
	}
	commandBusHandler := bus.NewCommandHandler()
	commandBus := eventhorizon.UseCommandHandlerMiddleware(commandBusHandler,
		commandbus.Logger(bankapitest.Logger),
	)
	err = commandhandler.Setup(eventStore, commandBusHandler)
	if err != nil {
		panic(err)
	}
	return commandBus, eventStore.Outbox, tearDown
}

type ReadRepoTearDown func()

func ReadRepo() (readrepository.ReadRepository, eventhorizon.CommandHandler, ReadRepoTearDown) {
	commandHandler, outbox, storeTearDown := CommandHandler()

	client, err := projectionclient.Open(bankapitest.Config, bankapitest.Logger, outbox)
	if err != nil {
		panic(err)
	}
	projection, err := projection.New(client)
	if err != nil {
		panic(err)
	}
	outbox.Start()

	tearDown := func() {
		err := client.OnDisconnect()
		if err != nil {
			panic(err)
		}
		storeTearDown()
	}
	return projection, commandHandler, tearDown
}
