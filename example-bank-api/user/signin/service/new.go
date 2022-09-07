package service

import (
	"codepix/example-bank-api/adapters/httputils"
	"codepix/example-bank-api/adapters/messagequeue"
	"codepix/example-bank-api/config"
	customerauth "codepix/example-bank-api/customer/auth"
	customerrepository "codepix/example-bank-api/customer/repository"
	"codepix/example-bank-api/lib/validation"
	userauth "codepix/example-bank-api/user/auth"
	userrepository "codepix/example-bank-api/user/repository"
	"codepix/example-bank-api/user/signin/repository"

	"github.com/justinas/alice"
)

func Register(
	config config.Config,
	chain alice.Chain,
	handle httputils.RouterHandler,
	validator *validation.Validator,
	messageQueue *messagequeue.MessageQueue,
	repository repository.Repository,
	userRepository userrepository.Repository,
	customerRepository customerrepository.Repository,
) error {
	service := &Service{
		MessageQueue:   messageQueue,
		Repository:     repository,
		UserRepository: userRepository,
	}
	handle("POST", "/signin-request", chain.Append(
		httputils.WithBody(validator, Start{}),
	).ThenFunc(service.Start))

	handle("POST", "/signin-request/:token", chain.Append(
		httputils.WithParams(validator, Finish{}),
		userauth.AddClaims(userRepository, repository, "token"),
		customerauth.AddClaims(customerRepository),
		service.Finish,
	).ThenFunc(userauth.CreateToken(config)))
	return nil
}
