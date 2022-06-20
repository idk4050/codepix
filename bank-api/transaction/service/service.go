package service

import (
	"bytes"
	accountrepository "codepix/bank-api/account/repository"
	"codepix/bank-api/adapters/rpc"
	"codepix/bank-api/bank/auth"
	pixkeyrepository "codepix/bank-api/pixkey/repository"
	"codepix/bank-api/transaction"
	"codepix/bank-api/transaction/readrepository"
	"codepix/bank-api/transaction/service/proto"
	"context"

	"github.com/google/uuid"
	"github.com/looplab/eventhorizon"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	CommandHandler    eventhorizon.CommandHandler
	Repository        readrepository.ReadRepository
	AccountRepository accountrepository.Repository
	PixKeyRepository  pixkeyrepository.Repository
	proto.UnimplementedTransactionServiceServer
}

var _ proto.TransactionServiceServer = Service{}

func (s Service) Start(ctx context.Context, req *proto.StartRequest) (*proto.StartReply, error) {
	bankID := auth.GetBankID(ctx)

	_, senderIDs, err := s.AccountRepository.FindByNumber(req.SenderAccountNumber)
	if err != nil {
		return nil, rpc.MapError(ctx, err)
	}
	if !bytes.Equal(senderIDs.BankID[:], bankID[:]) {
		return nil, status.Error(codes.PermissionDenied, "")
	}
	_, receiverIDs, err := s.PixKeyRepository.FindByKey(req.ReceiverKey)
	if err != nil {
		return nil, rpc.MapError(ctx, err)
	}

	ID := uuid.New()
	command := startCommand(req, ID, *senderIDs, *receiverIDs)
	err = s.CommandHandler.HandleCommand(ctx, command)
	return startReply(ID, err), rpc.MapError(ctx, err)
}

func startCommand(req *proto.StartRequest, ID uuid.UUID,
	senderIDs accountrepository.IDs, receiverIDs pixkeyrepository.IDs) transaction.Start {
	return transaction.Start{
		ID:           ID,
		Sender:       senderIDs.AccountID,
		SenderBank:   senderIDs.BankID,
		Receiver:     receiverIDs.PixKeyID,
		ReceiverBank: receiverIDs.BankID,
		Amount:       req.Amount,
		Description:  req.Description,
	}
}
func startReply(ID uuid.UUID, err error) *proto.StartReply {
	if err != nil {
		return nil
	}
	return &proto.StartReply{
		Id: ID[:],
	}
}

func (s Service) Find(ctx context.Context, req *proto.FindRequest) (*proto.FindReply, error) {
	bankID := auth.GetBankID(ctx)
	ID, _ := uuid.FromBytes(req.Id)

	transaction, err := s.Repository.Find(ID)
	if !(bytes.Equal(transaction.SenderBank[:], bankID[:]) ||
		bytes.Equal(transaction.ReceiverBank[:], bankID[:])) {
		return nil, status.Error(codes.PermissionDenied, "")
	}
	return findReply(transaction), rpc.MapError(ctx, err)
}

func findReply(transaction *readrepository.Transaction) *proto.FindReply {
	if transaction == nil {
		return nil
	}
	return &proto.FindReply{
		CreatedAt:        timestamppb.New(transaction.CreatedAt),
		UpdatedAt:        timestamppb.New(transaction.UpdatedAt),
		Sender:           transaction.Sender[:],
		Receiver:         transaction.Receiver[:],
		Amount:           transaction.Amount,
		Description:      transaction.Description,
		Status:           proto.Status(transaction.Status),
		ReasonForFailing: transaction.ReasonForFailing,
	}
}
