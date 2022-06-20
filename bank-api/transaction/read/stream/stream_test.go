package stream_test

import (
	"codepix/bank-api/bankapitest"
	proto "codepix/bank-api/proto/codepix/transaction/read"
	"codepix/bank-api/transaction"
	"codepix/bank-api/transaction/transactiontest"
	"context"
	"errors"
	"strconv"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/looplab/eventhorizon"
	"github.com/stretchr/testify/assert"
)

var Stream = transactiontest.ReadStream

var ValidStartCommand = func(ID, bankID uuid.UUID) transaction.Start {
	cmd := transactiontest.ValidStartCommand(ID)
	cmd.ReceiverBank = bankID
	return cmd
}

var busTimeout = bankapitest.Config.Transaction.BusBlockDuration * 2
var busInterval = busTimeout / 5

func TestConsume(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	stream, makeCtx, commandHandler, tearDown := Stream()
	defer tearDown()

	SingleEvent := func(t *testing.T) {
		ID := uuid.New()
		bankID := uuid.New()

		start := ValidStartCommand(ID, bankID)
		go commandHandler.HandleCommand(context.Background(), start)

		rec := make(chan []eventhorizon.Event)

		go stream.Consume(makeCtx(bankID),
			func(events []eventhorizon.Event) error {
				rec <- events
				return nil
			},
			func() (*proto.Ack, error) {
				return &proto.Ack{Nacks: []bool{false}}, nil
			},
			transaction.StartedEvent,
			transaction.StartedStream(bankID),
			bankID.String(),
		)
		received := <-rec
		assert.Len(t, received, 1)

		assert.Never(t, func() bool { return len(rec) > 0 }, busTimeout, busInterval)
	}
	SingleEventNack := func(t *testing.T) {
		ID := uuid.New()
		bankID := uuid.New()

		start := ValidStartCommand(ID, bankID)
		go commandHandler.HandleCommand(context.Background(), start)

		rec := make(chan []eventhorizon.Event, 1)
		nacks := make(chan []bool, 1)

		for i := 0; i < 2; i++ {
			go stream.Consume(makeCtx(bankID),
				func(events []eventhorizon.Event) error {
					rec <- events
					return nil
				},
				func() (*proto.Ack, error) {
					return &proto.Ack{Nacks: <-nacks}, nil
				},
				transaction.StartedEvent,
				transaction.StartedStream(bankID),
				bankID.String(),
			)
		}
		received := <-rec
		assert.Len(t, received, 1)
		nacks <- []bool{true}

		received = <-rec
		assert.Len(t, received, 1)
		nacks <- []bool{false}

		assert.Never(t, func() bool { return len(rec) > 0 }, busTimeout, busInterval)
	}
	MultipleEvents := func(nEvents int, bankID uuid.UUID, wg *sync.WaitGroup) func(t *testing.T) {
		return func(t *testing.T) {
			start := ValidStartCommand(uuid.New(), bankID)
			for i := 0; i < nEvents; i++ {
				copy := start
				copy.ID = uuid.New()
				commandHandler.HandleCommand(context.Background(), copy)
			}
			ch := make(chan []eventhorizon.Event, 1)

			go stream.Consume(makeCtx(bankID),
				func(events []eventhorizon.Event) error {
					for range events {
						wg.Done()
					}
					ch <- events
					return nil
				},
				func() (*proto.Ack, error) {
					events := <-ch
					return &proto.Ack{Nacks: make([]bool, len(events))}, nil
				},
				transaction.StartedEvent,
				transaction.StartedStream(bankID),
				bankID.String(),
			)
			wg.Wait()
		}
	}
	MultipleBanks := func(t *testing.T) {
		nBanks := 10
		nEventsPerBank := 10

		wg := &sync.WaitGroup{}
		wg.Add(nBanks * nEventsPerBank)
		for i := 0; i < nBanks; i++ {
			go MultipleEvents(nEventsPerBank, uuid.New(), wg)(t)
		}
		wg.Wait()
	}
	MultipleConsumers := func(t *testing.T) {
		nConsumers := 10
		nEventsPerConsumer := 10
		bankID := uuid.New()

		wg := &sync.WaitGroup{}
		wg.Add(nConsumers * nEventsPerConsumer)
		for i := 0; i < nConsumers; i++ {
			go MultipleEvents(nEventsPerConsumer, bankID, wg)(t)
		}
		wg.Wait()
	}
	FailToConsume := func(t *testing.T) {
		ID := uuid.New()
		bankID := uuid.New()

		start := ValidStartCommand(ID, bankID)
		go commandHandler.HandleCommand(context.Background(), start)

		ctx, cancel := context.WithCancel(makeCtx(bankID))
		cancel()
		err := stream.Consume(ctx,
			func(events []eventhorizon.Event) error {
				return nil
			},
			func() (*proto.Ack, error) {
				return nil, nil
			},
			transaction.StartedEvent,
			transaction.StartedStream(bankID),
			bankID.String(),
		)
		assert.ErrorContains(t, err, "create consumer group: context canceled")
	}
	FailToSend := func(t *testing.T) {
		ID := uuid.New()
		bankID := uuid.New()

		start := ValidStartCommand(ID, bankID)
		go commandHandler.HandleCommand(context.Background(), start)

		err := stream.Consume(makeCtx(bankID),
			func(events []eventhorizon.Event) error {
				return errors.New("failed to send events")
			},
			func() (*proto.Ack, error) {
				return nil, nil
			},
			transaction.StartedEvent,
			transaction.StartedStream(bankID),
			bankID.String(),
		)
		assert.ErrorContains(t, err, "failed to send events")
	}
	FailToReceiveNacks := func(t *testing.T) {
		ID := uuid.New()
		bankID := uuid.New()

		start := ValidStartCommand(ID, bankID)
		go commandHandler.HandleCommand(context.Background(), start)

		err := stream.Consume(makeCtx(bankID),
			func(events []eventhorizon.Event) error {
				return nil
			},
			func() (*proto.Ack, error) {
				return nil, errors.New("failed to receive ack")
			},
			transaction.StartedEvent,
			transaction.StartedStream(bankID),
			bankID.String(),
		)
		assert.ErrorContains(t, err, "failed to receive ack")
	}
	InvalidNacks := func(t *testing.T) {
		ID := uuid.New()
		bankID := uuid.New()

		start := ValidStartCommand(ID, bankID)
		go commandHandler.HandleCommand(context.Background(), start)

		err := stream.Consume(makeCtx(bankID),
			func(events []eventhorizon.Event) error {
				return nil
			},
			func() (*proto.Ack, error) {
				return &proto.Ack{Nacks: []bool{}}, nil
			},
			transaction.StartedEvent,
			transaction.StartedStream(bankID),
			bankID.String(),
		)
		assert.ErrorContains(t, err, "expected 1 nacks, received 0")

		err = stream.Consume(makeCtx(bankID),
			func(events []eventhorizon.Event) error {
				return nil
			},
			func() (*proto.Ack, error) {
				return &proto.Ack{Nacks: []bool{false, false}}, nil
			},
			transaction.StartedEvent,
			transaction.StartedStream(bankID),
			bankID.String(),
		)
		assert.ErrorContains(t, err, "expected 1 nacks, received 2")
	}
	FailToAck := func(t *testing.T) {
		ID := uuid.New()
		bankID := uuid.New()

		start := ValidStartCommand(ID, bankID)
		go commandHandler.HandleCommand(context.Background(), start)

		ctx, cancel := context.WithCancel(makeCtx(bankID))
		err := stream.Consume(ctx,
			func(events []eventhorizon.Event) error {
				return nil
			},
			func() (*proto.Ack, error) {
				cancel()
				return &proto.Ack{Nacks: []bool{false}}, nil
			},
			transaction.StartedEvent,
			transaction.StartedStream(bankID),
			bankID.String(),
		)
		assert.ErrorContains(t, err, "ack messages: context canceled")
	}
	tests := []func(t *testing.T){
		SingleEvent,
		SingleEventNack,
		MultipleEvents(100, uuid.New(), func() *sync.WaitGroup {
			wg := &sync.WaitGroup{}
			wg.Add(100)
			return wg
		}()),
		MultipleBanks,
		MultipleConsumers,
		FailToConsume,
		FailToSend,
		FailToReceiveNacks,
		InvalidNacks,
		FailToAck,
	}
	for i, test := range tests {
		t.Run(strconv.Itoa(i), test)
	}
}
