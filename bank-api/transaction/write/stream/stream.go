package stream

import (
	"codepix/bank-api/adapters/rpc"
	"codepix/bank-api/bank/auth"
	pixkeyrepository "codepix/bank-api/pixkey/repository"
	proto "codepix/bank-api/proto/codepix/transaction/write"
	"codepix/bank-api/transaction"
	"context"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/looplab/eventhorizon"
	"google.golang.org/grpc/status"
	protobuf "google.golang.org/protobuf/proto"
)

type Stream struct {
	Logger           logr.Logger
	CommandHandler   eventhorizon.CommandHandler
	PixKeyRepository pixkeyrepository.Repository
	proto.UnimplementedStreamServer
}

var _ proto.StreamServer = Stream{}

func (s Stream) Write(ctx context.Context,
	writer string,
	commandType eventhorizon.CommandType,
	recv func() (protobuf.Message, error),
	send func(protobuf.Message) error,
	wrapError func(error) protobuf.Message,
	commandReply func(req protobuf.Message) (eventhorizon.Command, protobuf.Message, error),
) error {
	baseKvs := []any{
		"type", commandType,
		"writer", writer,
	}
	s.Logger.Info("writer connected", baseKvs...)
	for {
		select {
		case <-ctx.Done():
			err := ctx.Err()
			s.Logger.Error(err, "writer disconnected", baseKvs...)
			return err
		default:
		}
		req, err := recv()
		if err != nil {
			s.Logger.Error(err, "fail: receive command", baseKvs...)
			err := send(wrapError(rpc.MapError(ctx, err)))
			s.Logger.Error(err, "fail: send error", baseKvs...)
			continue
		}
		s.Logger.Info("command received", baseKvs...)

		command, reply, err := commandReply(req)
		if err != nil {
			s.Logger.Error(err, "fail: create command", baseKvs...)
			err := send(wrapError(rpc.MapError(ctx, err)))
			s.Logger.Error(err, "fail: send error", baseKvs...)
			continue
		}
		kvs := append(baseKvs, "tx", command.AggregateID())
		err = s.CommandHandler.HandleCommand(ctx, command)
		if err != nil {
			s.Logger.Error(err, "fail: handle command", kvs...)
			err := send(wrapError(rpc.MapError(ctx, err)))
			s.Logger.Error(err, "fail: send error", kvs...)
			continue
		}
		err = send(reply)
		if err != nil {
			s.Logger.Error(err, "fail: send reply", kvs...)
			continue
		}
		s.Logger.Info("reply sent", kvs...)
	}
}

func (s Stream) Start(stream proto.Stream_StartServer) error {
	ctx := stream.Context()
	bankID := auth.GetBankID(ctx)

	return s.Write(ctx,
		bankID.String(),
		transaction.StartCommand,
		func() (protobuf.Message, error) {
			return stream.Recv()
		},
		func(m protobuf.Message) error {
			return stream.SendMsg(m)
		},
		func(err error) protobuf.Message {
			return &proto.StartReply{
				Message: &proto.StartReply_Error{
					Error: status.Convert(err).Proto(),
				},
			}
		},
		func(m protobuf.Message) (eventhorizon.Command, protobuf.Message, error) {
			req := m.(*proto.StartRequest)
			senderID, _ := uuid.FromBytes(req.SenderId)
			_, receiverIDs, err := s.PixKeyRepository.FindByKey(req.ReceiverKey)
			if err != nil {
				return nil, nil, err
			}
			ID := uuid.New()
			command := startCommand(req, ID, bankID, senderID, *receiverIDs)
			reply := startReply(ID)
			return command, reply, nil
		},
	)
}

func startCommand(req *proto.StartRequest, ID, bankID, senderID uuid.UUID,
	receiverIDs pixkeyrepository.IDs) transaction.Start {
	return transaction.Start{
		ID:           ID,
		BankID:       bankID,
		Sender:       senderID,
		SenderBank:   bankID,
		Receiver:     receiverIDs.AccountID,
		ReceiverBank: receiverIDs.BankID,
		Amount:       req.Amount,
		Description:  req.Description,
	}
}
func startReply(ID uuid.UUID) *proto.StartReply {
	return &proto.StartReply{
		Message: &proto.StartReply_Started{
			Started: &proto.Started{
				Id: ID[:],
			},
		},
	}
}

func (s Stream) Confirm(stream proto.Stream_ConfirmServer) error {
	ctx := stream.Context()
	bankID := auth.GetBankID(ctx)

	return s.Write(ctx,
		bankID.String(),
		transaction.ConfirmCommand,
		func() (protobuf.Message, error) {
			return stream.Recv()
		},
		func(m protobuf.Message) error {
			return stream.SendMsg(m)
		},
		func(err error) protobuf.Message {
			return &proto.ConfirmReply{
				Message: &proto.ConfirmReply_Error{
					Error: status.Convert(err).Proto(),
				},
			}
		},
		func(m protobuf.Message) (eventhorizon.Command, protobuf.Message, error) {
			req := m.(*proto.ConfirmRequest)
			command := confirmCommand(req, bankID)
			reply := confirmReply()
			return command, reply, nil
		},
	)
}

func confirmCommand(req *proto.ConfirmRequest, bankID uuid.UUID) transaction.Confirm {
	ID, _ := uuid.FromBytes(req.Id)
	return transaction.Confirm{
		ID:     ID,
		BankID: bankID,
	}
}
func confirmReply() *proto.ConfirmReply {
	return &proto.ConfirmReply{
		Message: &proto.ConfirmReply_Confirmed{
			Confirmed: &proto.Confirmed{},
		},
	}
}

func (s Stream) Complete(stream proto.Stream_CompleteServer) error {
	ctx := stream.Context()
	bankID := auth.GetBankID(ctx)

	return s.Write(ctx,
		bankID.String(),
		transaction.CompleteCommand,
		func() (protobuf.Message, error) {
			return stream.Recv()
		},
		func(m protobuf.Message) error {
			return stream.SendMsg(m)
		},
		func(err error) protobuf.Message {
			return &proto.CompleteReply{
				Message: &proto.CompleteReply_Error{
					Error: status.Convert(err).Proto(),
				},
			}
		},
		func(m protobuf.Message) (eventhorizon.Command, protobuf.Message, error) {
			req := m.(*proto.CompleteRequest)
			command := completeCommand(req, bankID)
			reply := completeReply()
			return command, reply, nil
		},
	)
}

func completeCommand(req *proto.CompleteRequest, bankID uuid.UUID) transaction.Complete {
	ID, _ := uuid.FromBytes(req.Id)
	return transaction.Complete{
		ID:     ID,
		BankID: bankID,
	}
}
func completeReply() *proto.CompleteReply {
	return &proto.CompleteReply{
		Message: &proto.CompleteReply_Completed{
			Completed: &proto.Completed{},
		},
	}
}

func (s Stream) Fail(stream proto.Stream_FailServer) error {
	ctx := stream.Context()
	bankID := auth.GetBankID(ctx)

	return s.Write(ctx,
		bankID.String(),
		transaction.FailCommand,
		func() (protobuf.Message, error) {
			return stream.Recv()
		},
		func(m protobuf.Message) error {
			return stream.SendMsg(m)
		},
		func(err error) protobuf.Message {
			return &proto.FailReply{
				Message: &proto.FailReply_Error{
					Error: status.Convert(err).Proto(),
				},
			}
		},
		func(m protobuf.Message) (eventhorizon.Command, protobuf.Message, error) {
			req := m.(*proto.FailRequest)
			command := failCommand(req, bankID)
			reply := failReply()
			return command, reply, nil
		},
	)
}

func failCommand(req *proto.FailRequest, bankID uuid.UUID) transaction.Fail {
	ID, _ := uuid.FromBytes(req.Id)
	return transaction.Fail{
		ID:     ID,
		BankID: bankID,
		Reason: req.Reason,
	}
}
func failReply() *proto.FailReply {
	return &proto.FailReply{
		Message: &proto.FailReply_Failed{
			Failed: &proto.Failed{},
		},
	}
}
