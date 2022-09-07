package service

import (
	"codepix/example-bank-api/adapters/httputils"
	"codepix/example-bank-api/config"
	"codepix/example-bank-api/customer/auth"
	"codepix/example-bank-api/customer/repository"
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
	anyCustomer := anyUser.Append(auth.ValidateClaims)
	customer := anyCustomer.Append(auth.ClaimedAndParamIDsMatch("customer-id"))

	handle("GET", "/customer/:customer-id", customer.Append(
		httputils.WithParams(validator, Find{}),
	).ThenFunc(service.Find))
	return nil
}
