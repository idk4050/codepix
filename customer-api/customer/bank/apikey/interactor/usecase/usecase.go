package usecase

import (
	"codepix/customer-api/customer/bank/apikey"
	"codepix/customer-api/customer/bank/apikey/interactor"
	"codepix/customer-api/customer/bank/apikey/repository"
	"codepix/customer-api/lib/repositories"
	"errors"
)

type Usecase struct {
	Repository repository.Repository
}

var _ interactor.Interactor = Usecase{}

func (uc Usecase) Create(input interactor.CreateInput) (*interactor.CreateOutput, error) {
	for {
		apiKey, err := apikey.New(input.Name)
		if err != nil {
			return nil, err
		}
		ID, err := uc.Repository.Add(*apiKey, input.BankID)
		if err == nil {
			output := &interactor.CreateOutput{
				APIKey: *apiKey,
				ID:     *ID,
			}
			return output, nil
		}
		var alreadyExists *repositories.AlreadyExistsError
		if !errors.As(err, &alreadyExists) {
			return nil, err
		}
	}
}
