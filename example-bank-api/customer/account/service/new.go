package service

import (
	"codepix/example-bank-api/adapters/httputils"
	"codepix/example-bank-api/config"
	"codepix/example-bank-api/customer/account/repository"
	customerauth "codepix/example-bank-api/customer/auth"
	"codepix/example-bank-api/lib/validation"
	userauth "codepix/example-bank-api/user/auth"

	"github.com/justinas/alice"
)

func Register(
	config config.Config,
	chain alice.Chain,
	handle httputils.RouterHandler,
	validator *validation.Validator,
	repository repository.Repository,
) error {
	service := &Service{
		Repository: repository,
	}
	anyUser := chain.Append(userauth.ValidateToken(config))
	anyCustomer := anyUser.Append(customerauth.ValidateClaims)
	accountOwner := anyCustomer.Append(customerauth.CustomerOwnsParamAccountID(repository, "account-id"))
	customer := anyCustomer.Append(customerauth.ClaimedAndParamIDsMatch("customer-id"))

	handle("POST", "/account", anyCustomer.ThenFunc(service.Register))

	handle("GET", "/account/:account-id", accountOwner.Append(
		httputils.WithParams(validator, Find{}),
	).ThenFunc(service.Find))

	handle("DELETE", "/account/:account-id", accountOwner.Append(
		httputils.WithParams(validator, Remove{}),
	).ThenFunc(service.Remove))

	handle("GET", "/customer/:customer-id/accounts", customer.Append(
		httputils.WithParams(validator, List{}),
	).ThenFunc(service.List))
	return nil
}
