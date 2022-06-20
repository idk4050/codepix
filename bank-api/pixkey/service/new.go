package service

import (
	"bytes"
	accountrepository "codepix/bank-api/account/repository"
	"codepix/bank-api/adapters/validator"
	"codepix/bank-api/lib/validation"
	"codepix/bank-api/pixkey"
	"codepix/bank-api/pixkey/interactor"
	"codepix/bank-api/pixkey/repository"
	"codepix/bank-api/pixkey/service/proto"
	_ "embed"

	"github.com/cuducos/go-cpf"
	"google.golang.org/grpc"
)

//go:embed translations.json
var translations []byte

func Register(server *grpc.Server, val *validation.Validator,
	interactor interactor.Interactor, repository repository.Repository,
	accountRepository accountrepository.Repository,
) error {
	err := SetupValidator(val)
	if err != nil {
		return err
	}
	service := &Service{
		Interactor:        interactor,
		Repository:        repository,
		AccountRepository: accountRepository,
	}
	proto.RegisterPixKeyServiceServer(server, service)
	return nil
}

func SetupValidator(val *validation.Validator) error {
	err := validator.LoadTranslationFile(val, bytes.NewReader(translations), proto.RegisterRequest{})
	if err != nil {
		return err
	}
	err = validation.AddStructValidations(val,
		validation.StructValidation[proto.RegisterRequest]{
			Field: "Key",
			Tag:   "cpf_key",
			IsValid: func(request *proto.RegisterRequest) bool {
				return request.Type != proto.Type(pixkey.CPFKey) ||
					cpf.IsValid(request.Key)
			},
		},
		validation.StructValidation[proto.RegisterRequest]{
			Field: "Key",
			Tag:   "phone_key",
			IsValid: func(request *proto.RegisterRequest) bool {
				return request.Type != proto.Type(pixkey.PhoneKey) ||
					validation.IsValid(val, request.Key, "e164")
			},
		},
		validation.StructValidation[proto.RegisterRequest]{
			Field: "Key",
			Tag:   "email_key",
			IsValid: func(request *proto.RegisterRequest) bool {
				return request.Type != proto.Type(pixkey.EmailKey) ||
					validation.IsValid(val, request.Key, "email")
			},
		},
	)
	return err
}
