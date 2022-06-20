package service

import (
	"codepix/bank-api/adapters/rpc"
	"codepix/bank-api/bank/auth"
	proto "codepix/bank-api/proto/codepix/transaction/read"
	"codepix/bank-api/transaction/read/repository"
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	Repository repository.Repository
	proto.UnimplementedServiceServer
}

var _ proto.ServiceServer = Service{}

func (s Service) Find(ctx context.Context, req *proto.FindRequest) (*proto.FindReply, error) {
	bankID := auth.GetBankID(ctx)
	ID, _ := uuid.FromBytes(req.Id)

	transaction, err := s.Repository.Find(ctx, ID)
	if err == nil {
		allowed := transaction.SenderBank == bankID || transaction.ReceiverBank == bankID
		if !allowed {
			return nil, status.Error(codes.PermissionDenied, "")
		}
	}
	return findReply(transaction), rpc.MapError(ctx, err)
}

func findReply(transaction *repository.Transaction) *proto.FindReply {
	if transaction == nil {
		return nil
	}
	return &proto.FindReply{
		Id:           transaction.ID[:],
		Sender:       transaction.Sender[:],
		SenderBank:   transaction.SenderBank[:],
		Receiver:     transaction.Receiver[:],
		ReceiverBank: transaction.ReceiverBank[:],

		CreatedAt:        timestamppb.New(transaction.CreatedAt),
		UpdatedAt:        timestamppb.New(transaction.UpdatedAt),
		Amount:           transaction.Amount,
		Description:      transaction.Description,
		Status:           proto.Status(transaction.Status),
		ReasonForFailing: transaction.ReasonForFailing,
	}
}

func (s Service) List(ctx context.Context, req *proto.ListRequest) (*proto.ListReply, error) {
	bankID := auth.GetBankID(ctx)
	senderID, _ := uuid.FromBytes(req.SenderId)
	receiverID, _ := uuid.FromBytes(req.ReceiverId)

	options := repository.ListOptions{
		CreatedAfter: req.CreatedAfter.AsTime(),
		SenderID:     senderID,
		ReceiverID:   receiverID,
		Limit:        req.Limit,
		Skip:         req.Skip,
	}
	transactions, err := s.Repository.List(ctx, options)
	if err == nil {
		for _, transaction := range transactions {
			allowed := transaction.SenderBank == bankID || transaction.ReceiverBank == bankID
			if !allowed {
				return nil, status.Error(codes.PermissionDenied, "")
			}
		}
	}
	return listReply(transactions), rpc.MapError(ctx, err)
}

func listReply(transactions []repository.ListItem) *proto.ListReply {
	if transactions == nil {
		return nil
	}
	items := []*proto.ListItem{}
	for _, transaction := range transactions {
		items = append(items, listItemReply(transaction))
	}
	return &proto.ListReply{
		Items: items,
	}
}

func listItemReply(transaction repository.ListItem) *proto.ListItem {
	return &proto.ListItem{
		Id:           transaction.ID[:],
		Sender:       transaction.Sender[:],
		SenderBank:   transaction.SenderBank[:],
		Receiver:     transaction.Receiver[:],
		ReceiverBank: transaction.ReceiverBank[:],

		CreatedAt:        timestamppb.New(transaction.CreatedAt),
		UpdatedAt:        timestamppb.New(transaction.UpdatedAt),
		Amount:           transaction.Amount,
		Description:      transaction.Description,
		Status:           proto.Status(transaction.Status),
		ReasonForFailing: transaction.ReasonForFailing,
	}
}
