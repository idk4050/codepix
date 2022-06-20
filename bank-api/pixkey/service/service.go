package service

import (
	accountrepository "codepix/bank-api/account/repository"
	"codepix/bank-api/adapters/rpc"
	"codepix/bank-api/bank/auth"
	"codepix/bank-api/lib/repositories"
	"codepix/bank-api/pixkey"
	"codepix/bank-api/pixkey/interactor"
	"codepix/bank-api/pixkey/repository"
	"codepix/bank-api/pixkey/service/proto"
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
)

type Service struct {
	Interactor        interactor.Interactor
	Repository        repository.Repository
	AccountRepository accountrepository.Repository
	proto.UnimplementedPixKeyServiceServer
}

var _ proto.PixKeyServiceServer = Service{}

func (s Service) Register(ctx context.Context, req *proto.RegisterRequest,
) (*proto.RegisterReply, error) {
	accountID, _ := uuid.FromBytes(req.AccountId)
	bankID := auth.GetBankID(ctx)

	err := s.AccountRepository.ExistsWithBankID(accountID, bankID)
	if err != nil {
		return nil, rpc.MapError(ctx, err, rpc.Mapping{
			&repositories.NotFoundError{}: codes.PermissionDenied,
		})
	}

	input := registerInput(req, accountID)
	output, err := s.Interactor.Register(input)
	return registerReply(output), rpc.MapError(ctx, err)
}

func registerInput(req *proto.RegisterRequest, accountID uuid.UUID) interactor.RegisterInput {
	return interactor.RegisterInput{
		Type:      pixkey.Type(req.Type),
		Key:       req.Key,
		AccountID: accountID,
	}
}
func registerReply(output *interactor.RegisterOutput) *proto.RegisterReply {
	if output == nil {
		return nil
	}
	return &proto.RegisterReply{
		Id: output.ID[:],
	}
}

func (s Service) Find(ctx context.Context, req *proto.FindRequest) (*proto.FindReply, error) {
	ID, _ := uuid.FromBytes(req.Id)
	pixKey, _, err := s.Repository.Find(ID)
	return findReply(pixKey), rpc.MapError(ctx, err)
}

func findReply(pixKey *pixkey.PixKey) *proto.FindReply {
	if pixKey == nil {
		return nil
	}
	return &proto.FindReply{
		Type: proto.Type(pixKey.Type),
		Key:  pixKey.Key,
	}
}
