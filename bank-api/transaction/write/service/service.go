package service

import (
	"codepix/bank-api/adapters/rpc"
	"codepix/bank-api/bank/auth"
	pixkeyrepository "codepix/bank-api/pixkey/repository"
	proto "codepix/bank-api/proto/codepix/transaction/write"
	"codepix/bank-api/transaction"
	"context"

	"github.com/google/uuid"
	"github.com/looplab/eventhorizon"
)

type Service struct {
	CommandHandler   eventhorizon.CommandHandler
	PixKeyRepository pixkeyrepository.Repository
	proto.UnimplementedServiceServer
}

var _ proto.ServiceServer = Service{}

func (s Service) Start(ctx context.Context, req *proto.StartRequest) (*proto.Started, error) {
	ID := uuid.New()
	bankID := auth.GetBankID(ctx)
	senderID, _ := uuid.FromBytes(req.SenderId)
	_, receiverIDs, err := s.PixKeyRepository.FindByKey(req.ReceiverKey)
	if err != nil {
		return nil, rpc.MapError(ctx, err)
	}
	command := startCommand(req, ID, bankID, senderID, *receiverIDs)
	err = s.CommandHandler.HandleCommand(ctx, command)
	return startReply(ID), rpc.MapError(ctx, err)
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
func startReply(ID uuid.UUID) *proto.Started {
	return &proto.Started{
		Id: ID[:],
	}
}
