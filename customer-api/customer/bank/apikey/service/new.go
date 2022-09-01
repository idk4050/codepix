package service

import (
	"codepix/customer-api/adapters/httputils"
	"codepix/customer-api/config"
	customerauth "codepix/customer-api/customer/auth"
	"codepix/customer-api/customer/bank/apikey/interactor"
	"codepix/customer-api/customer/bank/apikey/repository"
	bankrepository "codepix/customer-api/customer/bank/repository"
	"codepix/customer-api/lib/validation"
	userauth "codepix/customer-api/user/auth"

	"github.com/justinas/alice"
)

func Register(
	config config.Config,
	chain alice.Chain,
	handle httputils.RouterHandler,
	validator *validation.Validator,
	interactor interactor.Interactor,
	repository repository.Repository,
	bankRepository bankrepository.Repository,
) error {
	service := &Service{
		Repository: repository,
		Interactor: interactor,
	}
	anyUser := chain.Append(userauth.ValidateToken(config))
	anyCustomer := anyUser.Append(customerauth.ValidateClaims)
	bankOwner := anyCustomer.Append(customerauth.CustomerOwnsParamBankID(bankRepository))

	handle("POST", "/bank/:bank-id/apikeys", bankOwner.Append(
		httputils.WithBody(validator, Create{}),
		httputils.WithParams(validator, CreateParams{}),
	).ThenFunc(service.Create))

	handle("DELETE", "/bank/:bank-id/apikeys/:apikey-id", bankOwner.Append(
		httputils.WithParams(validator, Remove{}),
	).ThenFunc(service.Remove))

	return nil
}
