package service

import (
	"bytes"
	"codepix/bank-api/account"
	"codepix/bank-api/account/interactor"
	"codepix/bank-api/account/repository"
	"codepix/bank-api/account/service/proto"
	"codepix/bank-api/adapters/rpc"
	"codepix/bank-api/bank/auth"
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	Interactor interactor.Interactor
	Repository repository.Repository
	proto.UnimplementedAccountServiceServer
}

var _ proto.AccountServiceServer = Service{}

func (s Service) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterReply, error) {
	bankID := auth.GetBankID(ctx)
	input := registerInput(req, bankID)
	output, err := s.Interactor.Register(input)
	return registerReply(output), rpc.MapError(ctx, err)
}

func registerInput(req *proto.RegisterRequest, bankID uuid.UUID) interactor.RegisterInput {
	return interactor.RegisterInput{
		Number:    req.Number,
		OwnerName: req.OwnerName,
		BankID:    bankID,
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
	bankID := auth.GetBankID(ctx)
	ID, _ := uuid.FromBytes(req.Id)

	account, IDs, err := s.Repository.Find(ID)
	if err == nil && !bytes.Equal(IDs.BankID[:], bankID[:]) {
		return nil, status.Error(codes.PermissionDenied, "")
	}
	return findReply(account), rpc.MapError(ctx, err)
}

func findReply(account *account.Account) *proto.FindReply {
	if account == nil {
		return nil
	}
	return &proto.FindReply{
		Number:    account.Number,
		OwnerName: account.OwnerName,
	}
}
