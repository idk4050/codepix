package service

import (
	proto "codepix/bank-api/proto/codepix/transaction/read"
	"codepix/bank-api/transaction/read/repository"

	"google.golang.org/grpc"
)

func Register(server *grpc.Server, repository repository.Repository) error {
	service := &Service{Repository: repository}
	proto.RegisterServiceServer(server, service)
	return nil
}
