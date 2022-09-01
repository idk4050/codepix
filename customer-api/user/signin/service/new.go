package service

import (
	"codepix/customer-api/adapters/httputils"
	"codepix/customer-api/config"
	customerauth "codepix/customer-api/customer/auth"
	customerrepository "codepix/customer-api/customer/repository"
	"codepix/customer-api/lib/validation"
	userauth "codepix/customer-api/user/auth"
	userrepository "codepix/customer-api/user/repository"
	"codepix/customer-api/user/signin/interactor"
	"codepix/customer-api/user/signin/repository"

	"github.com/justinas/alice"
)

func Register(
	config config.Config,
	chain alice.Chain,
	handle httputils.RouterHandler,
	validator *validation.Validator,
	interactor interactor.Interactor,
	repository repository.Repository,
	userRepository userrepository.Repository,
	customerRepository customerrepository.Repository,
) error {
	service := &Service{
		Interactor: interactor,
	}

	handle("POST", "/signin-request", chain.Append(
		httputils.WithBody(validator, Start{}),
	).ThenFunc(service.Start))

	handle("DELETE", "/signin-request/:token", chain.Append(
		httputils.WithParams(validator, Finish{}),
		userauth.AddClaims(userRepository, repository),
		customerauth.AddClaims(customerRepository),
		service.Finish,
	).ThenFunc(userauth.CreateToken(config)))

	return nil
}
