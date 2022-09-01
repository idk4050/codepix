package service

import (
	"codepix/customer-api/adapters/httputils"
	"codepix/customer-api/config"
	customerauth "codepix/customer-api/customer/auth"
	apikeyrepository "codepix/customer-api/customer/bank/apikey/repository"
	"codepix/customer-api/customer/bank/interactor"
	"codepix/customer-api/customer/bank/repository"
	"codepix/customer-api/lib/validation"
	userauth "codepix/customer-api/user/auth"

	"github.com/justinas/alice"
)

func RegisterHandler(
	config config.Config,
	chain alice.Chain,
	handle httputils.RouterHandler,
	validator *validation.Validator,
	interactor interactor.Interactor,
	repository repository.Repository,
	apiKeyRepository apikeyrepository.Repository,
) error {
	service := &Service{
		Interactor:       interactor,
		Repository:       repository,
		APIKeyRepository: apiKeyRepository,
	}
	anyUser := chain.Append(userauth.ValidateToken(config))
	anyCustomer := anyUser.Append(customerauth.ValidateClaims)
	bankOwner := anyCustomer.Append(customerauth.CustomerOwnsParamBankID(repository))

	handle("POST", "/bank", anyCustomer.Append(
		httputils.WithBody(validator, Register{}),
	).ThenFunc(service.Register))

	handle("GET", "/bank/:bank-id", bankOwner.Append(
		httputils.WithParams(validator, Find{}),
	).ThenFunc(service.Find))

	handle("DELETE", "/bank/:bank-id", bankOwner.Append(
		httputils.WithParams(validator, Remove{}),
	).ThenFunc(service.Remove))

	handle("GET", "/bank/:bank-id/apikeys", bankOwner.Append(
		httputils.WithParams(validator, ListAPIKeys{}),
	).ThenFunc(service.ListAPIKeys))

	return nil
}
