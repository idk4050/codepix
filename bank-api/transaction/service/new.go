package service

import (
	"bytes"
	accountrepository "codepix/bank-api/account/repository"
	"codepix/bank-api/adapters/validator"
	"codepix/bank-api/lib/validation"
	pixkeyrepository "codepix/bank-api/pixkey/repository"
	"codepix/bank-api/transaction/readrepository"
	"codepix/bank-api/transaction/service/proto"
	_ "embed"

	"github.com/looplab/eventhorizon"
	"google.golang.org/grpc"
)

//go:embed translations.json
var translations []byte

func Register(server *grpc.Server, val *validation.Validator, commandHandler eventhorizon.CommandHandler, repository readrepository.ReadRepository, accountRepository accountrepository.Repository,
	pixKeyRepository pixkeyrepository.Repository,
) error {
	err := validator.LoadTranslationFile(val, bytes.NewReader(translations), proto.StartRequest{})
	if err != nil {
		return err
	}
	service := &Service{
		CommandHandler:    commandHandler,
		Repository:        repository,
		AccountRepository: accountRepository,
		PixKeyRepository:  pixKeyRepository,
	}
	proto.RegisterTransactionServiceServer(server, service)
	return nil
}
