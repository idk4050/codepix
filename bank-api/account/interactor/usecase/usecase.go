package usecase

import (
	"codepix/bank-api/account"
	"codepix/bank-api/account/interactor"
	"codepix/bank-api/account/repository"
)

type Usecase struct {
	Repository repository.Repository
}

var _ interactor.Interactor = Usecase{}

func (uc Usecase) Register(input interactor.RegisterInput) (*interactor.RegisterOutput, error) {
	account := account.Account{
		Number:    input.Number,
		OwnerName: input.OwnerName,
	}
	ID, err := uc.Repository.Add(account, input.BankID)
	if err != nil {
		return nil, err
	}
	output := &interactor.RegisterOutput{
		Account: account,
		ID:      *ID,
	}
	return output, nil
}
