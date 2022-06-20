package service

import (
	"codepix/bank-api/adapters/rpc"
	"codepix/bank-api/bank/auth"
	"codepix/bank-api/pixkey"
	"codepix/bank-api/pixkey/repository"
	proto "codepix/bank-api/proto/codepix/pixkey"
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	Repository repository.Repository
	proto.UnimplementedServiceServer
}

var _ proto.ServiceServer = Service{}

func (s Service) Register(ctx context.Context, req *proto.RegisterRequest,
) (*proto.RegisterReply, error) {
	bankID := auth.GetBankID(ctx)
	accountID, _ := uuid.FromBytes(req.AccountId)

	ID, err := s.Repository.Add(newPixKey(req), accountID, bankID)
	return registerReply(ID), rpc.MapError(ctx, err)
}

func newPixKey(req *proto.RegisterRequest) pixkey.PixKey {
	return pixkey.PixKey{
		Type: pixkey.Type(req.Type),
		Key:  req.Key,
	}
}
func registerReply(ID *uuid.UUID) *proto.RegisterReply {
	if ID == nil {
		return nil
	}
	return &proto.RegisterReply{
		Id: ID[:],
	}
}

func (s Service) Find(ctx context.Context, req *proto.FindRequest) (*proto.FindReply, error) {
	bankID := auth.GetBankID(ctx)
	ID, _ := uuid.FromBytes(req.Id)

	pixKey, IDs, err := s.Repository.Find(ID)
	if err == nil && IDs.BankID != bankID {
		return nil, status.Error(codes.PermissionDenied, "")
	}
	return findReply(pixKey, IDs), rpc.MapError(ctx, err)
}

func findReply(pixKey *pixkey.PixKey, IDs *repository.IDs) *proto.FindReply {
	if pixKey == nil {
		return nil
	}
	return &proto.FindReply{
		Id:        IDs.PixKeyID[:],
		Type:      proto.Type(pixKey.Type),
		Key:       pixKey.Key,
		AccountId: IDs.AccountID[:],
	}
}

func (s Service) List(ctx context.Context, req *proto.ListRequest) (*proto.ListReply, error) {
	bankID := auth.GetBankID(ctx)
	accountID, _ := uuid.FromBytes(req.AccountId)

	options := repository.ListOptions{
		AccountID: accountID,
		BankID:    bankID,
	}
	pixKeys, err := s.Repository.List(options)
	return ListReply(pixKeys), rpc.MapError(ctx, err)
}

func ListReply(pixKeys []repository.ListItem) *proto.ListReply {
	if pixKeys == nil {
		return nil
	}
	items := []*proto.ListItem{}
	for _, pixKey := range pixKeys {
		ID := pixKey.ID
		items = append(items, &proto.ListItem{
			Id:   ID[:],
			Type: proto.Type(pixKey.Type),
			Key:  pixKey.Key,
		})
	}
	return &proto.ListReply{
		Items: items,
	}
}
