package stream

import (
	"codepix/bank-api/adapters/eventbus"
	"codepix/bank-api/bank/auth"
	proto "codepix/bank-api/proto/codepix/transaction/read"
	"codepix/bank-api/transaction"
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/looplab/eventhorizon"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Stream struct {
	Logger    logr.Logger
	BusReader *eventbus.Reader
	proto.UnimplementedStreamServer
}

var _ proto.StreamServer = Stream{}

func (s Stream) Consume(ctx context.Context,
	sendEvents func([]eventhorizon.Event) error,
	receiveAck func() (*proto.Ack, error),
	eventType eventhorizon.EventType,
	streamName string,
	group string,
) error {
	randomID := uuid.New().String()
	consumer := randomID[:8]

	subKvs := []any{
		"type", eventType,
		"group", group,
		"consumer", consumer,
	}
	err := s.BusReader.CreateGroup(ctx, streamName, group)
	if err != nil {
		s.Logger.Error(err, "fail: create consumer group", subKvs...)
		return err
	}
	s.Logger.Info("consumer group created", subKvs...)

	for {
		events, messageIDs, err := s.BusReader.Consume(ctx, streamName, group, consumer)
		if err != nil {
			s.Logger.Error(err, "fail: consume events", subKvs...)
			return err
		}
		if len(events) == 0 {
			continue
		}
		eventIDs := []uuid.UUID{}
		for _, event := range events {
			eventIDs = append(eventIDs, event.AggregateID())
		}
		sendKvs := append(subKvs,
			"events", eventIDs,
			"messages", messageIDs,
		)
		err = sendEvents(events)
		if err != nil {
			s.Logger.Error(err, "fail: send events", sendKvs...)
			return err
		}
		s.Logger.Info("events sent", sendKvs...)

		ack, err := receiveAck()
		if err != nil {
			s.Logger.Error(err, "fail: receive ack", sendKvs...)
			return err
		}
		if len(ack.Nacks) != len(messageIDs) {
			err := fmt.Errorf("expected %d nacks, received %d", len(messageIDs), len(ack.Nacks))
			s.Logger.Error(err, "invalid ack", sendKvs...)
			return err
		}
		goodEvents := []uuid.UUID{}
		goodMessages := []string{}
		badEvents := []uuid.UUID{}
		badMessages := []string{}
		for i, nack := range ack.Nacks {
			if nack {
				badEvents = append(badEvents, eventIDs[i])
				badMessages = append(badMessages, messageIDs[i])
			} else {
				goodEvents = append(goodEvents, eventIDs[i])
				goodMessages = append(goodMessages, messageIDs[i])
			}
		}
		ackKvs := append(subKvs,
			"nacks", fmt.Sprintf("%d/%d", len(badMessages), len(messageIDs)),
			"good-events", goodEvents, "bad-events", badEvents,
			"good-messages", goodMessages, "bad-messages", badMessages,
		)
		s.Logger.Info("ack received", ackKvs...)

		if len(goodMessages) == 0 {
			continue
		}
		err = s.BusReader.Ack(ctx, streamName, group, goodMessages)
		if err != nil {
			s.Logger.Error(err, "fail: ack events", ackKvs...)
			return err
		}
		s.Logger.Info("events acked", ackKvs...)
	}
}

func (s Stream) Started(stream proto.Stream_StartedServer) error {
	sender := func(events []eventhorizon.Event) error {
		ps := []*proto.StartedTransaction{}
		for _, event := range events {
			p := startedMapper(event)
			ps = append(ps, p)
		}
		return stream.Send(&proto.StartedTransactions{
			Events: ps,
		})
	}
	bankID := auth.GetBankID(stream.Context())
	return s.Consume(stream.Context(),
		sender,
		stream.Recv,
		transaction.StartedEvent,
		transaction.StartedStream(bankID),
		bankID.String(),
	)
}
func startedMapper(event eventhorizon.Event) *proto.StartedTransaction {
	ID := event.AggregateID()
	started := event.Data().(*transaction.TransactionStarted)
	return &proto.StartedTransaction{
		Id:           ID[:],
		Timestamp:    timestamppb.New(event.Timestamp()),
		Sender:       started.Sender[:],
		SenderBank:   started.SenderBank[:],
		Receiver:     started.Receiver[:],
		ReceiverBank: started.ReceiverBank[:],
		Amount:       started.Amount,
		Description:  started.Description,
	}
}

func (s Stream) Confirmed(stream proto.Stream_ConfirmedServer) error {
	sender := func(events []eventhorizon.Event) error {
		ps := []*proto.ConfirmedTransaction{}
		for _, event := range events {
			p := confirmedMapper(event)
			ps = append(ps, p)
		}
		return stream.Send(&proto.ConfirmedTransactions{
			Events: ps,
		})
	}
	bankID := auth.GetBankID(stream.Context())
	return s.Consume(stream.Context(),
		sender,
		stream.Recv,
		transaction.ConfirmedEvent,
		transaction.ConfirmedStream(bankID),
		bankID.String(),
	)
}
func confirmedMapper(event eventhorizon.Event) *proto.ConfirmedTransaction {
	ID := event.AggregateID()
	return &proto.ConfirmedTransaction{
		Id:        ID[:],
		Timestamp: timestamppb.New(event.Timestamp()),
	}
}

func (s Stream) Completed(stream proto.Stream_CompletedServer) error {
	sender := func(events []eventhorizon.Event) error {
		ps := []*proto.CompletedTransaction{}
		for _, event := range events {
			p := completedMapper(event)
			ps = append(ps, p)
		}
		return stream.Send(&proto.CompletedTransactions{
			Events: ps,
		})
	}
	bankID := auth.GetBankID(stream.Context())
	return s.Consume(stream.Context(),
		sender,
		stream.Recv,
		transaction.CompletedEvent,
		transaction.CompletedStream(bankID),
		bankID.String(),
	)
}
func completedMapper(event eventhorizon.Event) *proto.CompletedTransaction {
	ID := event.AggregateID()
	return &proto.CompletedTransaction{
		Id:        ID[:],
		Timestamp: timestamppb.New(event.Timestamp()),
	}
}

func (s Stream) Failed(stream proto.Stream_FailedServer) error {
	sender := func(events []eventhorizon.Event) error {
		ps := []*proto.FailedTransaction{}
		for _, event := range events {
			p := failedMapper(event)
			ps = append(ps, p)
		}
		return stream.Send(&proto.FailedTransactions{
			Events: ps,
		})
	}
	bankID := auth.GetBankID(stream.Context())
	return s.Consume(stream.Context(),
		sender,
		stream.Recv,
		transaction.FailedEvent,
		transaction.FailedStream(bankID),
		bankID.String(),
	)
}
func failedMapper(event eventhorizon.Event) *proto.FailedTransaction {
	ID := event.AggregateID()
	failed := event.Data().(*transaction.TransactionFailed)
	return &proto.FailedTransaction{
		Id:        ID[:],
		Timestamp: timestamppb.New(event.Timestamp()),
		Reason:    failed.Reason,
	}
}
