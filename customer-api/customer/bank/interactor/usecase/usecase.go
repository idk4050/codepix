package usecase

import (
	"codepix/customer-api/customer/bank"
	"codepix/customer-api/customer/bank/interactor"
	"codepix/customer-api/customer/bank/repository"
	customerrepository "codepix/customer-api/customer/repository"
)

type Usecase struct {
	Repository         repository.Repository
	CustomerRepository customerrepository.Repository
}

var _ interactor.Interactor = Usecase{}

func (uc Usecase) Register(input interactor.RegisterInput) (*interactor.RegisterOutput, error) {
	customer, err := uc.CustomerRepository.Find(input.CustomerID)
	if err != nil {
		return nil, err
	}
	bank, err := bank.New(input.Code, input.Name, *customer)
	if err != nil {
		return nil, err
	}
	ID, err := uc.Repository.Add(*bank, input.CustomerID)
	if err != nil {
		return nil, err
	}
	output := &interactor.RegisterOutput{
		Bank: *bank,
		ID:   *ID,
	}
	return output, nil
}
