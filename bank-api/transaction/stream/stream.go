package stream

import (
	"codepix/bank-api/account"
	accountrepository "codepix/bank-api/account/repository"
	"codepix/bank-api/adapters/rpc"
	"codepix/bank-api/bank/auth"
	"codepix/bank-api/transaction"
	"codepix/bank-api/transaction/stream/proto"
	"context"

	"github.com/google/uuid"
	"github.com/looplab/eventhorizon"
	"google.golang.org/grpc/status"
)

type Stream struct {
	StartedBus        StartedBus
	ConfirmedBus      ConfirmedBus
	CommandHandler    eventhorizon.CommandHandler
	AccountRepository accountrepository.Repository
	proto.UnimplementedTransactionStreamServer
}

func (s *Stream) StartedTransactions(stream proto.TransactionStream_StartedServer) error {
	ctx := stream.Context()
	receiverBankID := auth.GetBankID(ctx)

	handler := func(event transaction.TransactionStarted, ID uuid.UUID) error {
		receiver, _, err := s.AccountRepository.Find(event.Receiver)
		if err != nil {
			return err
		}

		out := Started(event, ID, receiver)
		outErr := stream.Send(&out)
		if outErr != nil {
			return outErr
		}
		in, inErr := stream.Recv()
		if inErr != nil {
			return inErr
		}

		var cmdErr error
		if in.Fail != nil {
			command := Fail(in.Fail)
			cmdErr = s.CommandHandler.HandleCommand(ctx, command)
		} else {
			command := Confirm(in.Confirm, event.SenderBank)
			cmdErr = s.CommandHandler.HandleCommand(ctx, command)
		}
		if cmdErr != nil {
			out = proto.StartedOut{
				Error: Error(ID, cmdErr),
			}
			outErr = stream.Send(&out)
			if outErr != nil {
				return outErr
			}
			return cmdErr
		}
		return nil
	}
	return s.StartedBus.AddReceiverHandler(receiverBankID, handler)
}

func Started(event transaction.TransactionStarted, ID uuid.UUID, receiver *account.Account,
) proto.StartedOut {
	return proto.StartedOut{
		Started: &proto.TransactionStarted{
			Id:                    ID[:],
			ReceiverAccountNumber: receiver.Number,
			Amount:                event.Amount,
			Description:           event.Description,
		},
	}
}
func Confirm(in *proto.ConfirmTransaction, senderBank uuid.UUID) transaction.Confirm {
	ID, _ := uuid.FromBytes(in.Id)
	return transaction.Confirm{
		ID:         ID,
		SenderBank: senderBank,
	}
}

func (s *Stream) ConfirmedTransactions(stream proto.TransactionStream_ConfirmedServer) error {
	ctx := stream.Context()
	senderBankID := auth.GetBankID(ctx)

	handler := func(event transaction.TransactionConfirmed, ID uuid.UUID) error {
		out := Confirmed(event, ID)
		outErr := stream.Send(&out)
		if outErr != nil {
			return outErr
		}
		in, inErr := stream.Recv()
		if inErr != nil {
			return inErr
		}

		var cmdErr error
		if in.Fail != nil {
			command := Fail(in.Fail)
			cmdErr = s.CommandHandler.HandleCommand(ctx, command)
		} else {
			command := Complete(in.Complete)
			cmdErr = s.CommandHandler.HandleCommand(ctx, command)
		}
		if cmdErr != nil {
			out = proto.ConfirmedOut{
				Error: Error(ID, cmdErr),
			}
			outErr = stream.Send(&out)
			if outErr != nil {
				return outErr
			}
			return cmdErr
		}
		return nil
	}
	return s.ConfirmedBus.AddSenderHandler(senderBankID, handler)
}

func Confirmed(event transaction.TransactionConfirmed, ID uuid.UUID) proto.ConfirmedOut {
	return proto.ConfirmedOut{
		Confirmed: &proto.TransactionConfirmed{
			Id: ID[:],
		},
	}
}
func Complete(in *proto.CompleteTransaction) transaction.Complete {
	ID, _ := uuid.FromBytes(in.Id)
	return transaction.Complete{
		ID: ID,
	}
}

func Error(ID uuid.UUID, err error) *proto.TransactionError {
	statusError := rpc.MapError(context.Background(), err)
	status, _ := status.FromError(statusError)
	return &proto.TransactionError{
		Id:      ID[:],
		Code:    int32(status.Code()),
		Message: status.Message(),
	}
}
func Fail(in *proto.FailTransaction) transaction.Fail {
	ID, _ := uuid.FromBytes(in.Id)
	return transaction.Fail{
		ID:     ID,
		Reason: in.Reason,
	}
}
