package service

import (
	"bytes"
	"codepix/bank-api/adapters/validator"
	"codepix/bank-api/lib/validation"
	pixkeyrepository "codepix/bank-api/pixkey/repository"
	proto "codepix/bank-api/proto/codepix/transaction/write"
	"codepix/bank-api/transaction/write"

	"github.com/looplab/eventhorizon"
	"google.golang.org/grpc"
)

func Register(server *grpc.Server, val *validation.Validator,
	commandHandler eventhorizon.CommandHandler, pixKeyRepository pixkeyrepository.Repository,
) error {
	err := validator.LoadTranslationFile(val, bytes.NewReader(write.Translations),
		proto.StartRequest{},
		proto.FailRequest{},
	)
	if err != nil {
		return err
	}
	service := &Service{
		CommandHandler:   commandHandler,
		PixKeyRepository: pixKeyRepository,
	}
	proto.RegisterServiceServer(server, service)
	return nil
}
