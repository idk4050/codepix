package service

import (
	"bytes"
	"codepix/bank-api/account/interactor"
	"codepix/bank-api/account/repository"
	"codepix/bank-api/account/service/proto"
	"codepix/bank-api/adapters/validator"
	"codepix/bank-api/lib/validation"
	_ "embed"

	"google.golang.org/grpc"
)

//go:embed translations.json
var translations []byte

func Register(server *grpc.Server, val *validation.Validator,
	interactor interactor.Interactor, repository repository.Repository,
) error {
	err := validator.LoadTranslationFile(val, bytes.NewReader(translations), proto.RegisterRequest{})
	if err != nil {
		return err
	}
	service := &Service{
		Interactor: interactor,
		Repository: repository,
	}
	proto.RegisterAccountServiceServer(server, service)
	return nil
}
