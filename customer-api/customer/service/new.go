package service

import (
	"codepix/customer-api/adapters/httputils"
	"codepix/customer-api/config"
	"codepix/customer-api/customer/auth"
	"codepix/customer-api/customer/repository"
	"codepix/customer-api/lib/validation"
	userauth "codepix/customer-api/user/auth"

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
	customer := anyCustomer.Append(auth.ClaimedAndParamIDsMatch)

	handle("GET", "/customer/:customer-id", customer.Append(
		httputils.WithParams(validator, Find{}),
	).ThenFunc(service.Find))

	return nil
}
