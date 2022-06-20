package usecase

import (
	accountrepository "codepix/bank-api/account/repository"
	"codepix/bank-api/pixkey"
	"codepix/bank-api/pixkey/interactor"
	"codepix/bank-api/pixkey/repository"
)

type Usecase struct {
	Repository        repository.Repository
	AccountRepository accountrepository.Repository
}

var _ interactor.Interactor = Usecase{}

func (uc Usecase) Register(input interactor.RegisterInput) (*interactor.RegisterOutput, error) {
	pixKey := pixkey.PixKey{
		Type: input.Type,
		Key:  input.Key,
	}
	ID, err := uc.Repository.Add(pixKey, input.AccountID)
	if err != nil {
		return nil, err
	}
	output := &interactor.RegisterOutput{
		PixKey: pixKey,
		ID:     *ID,
	}
	return output, nil
}
