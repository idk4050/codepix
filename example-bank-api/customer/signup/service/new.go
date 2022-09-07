package service

import (
	"codepix/example-bank-api/adapters/httputils"
	"codepix/example-bank-api/adapters/messagequeue"
	"codepix/example-bank-api/customer/signup/repository"
	"codepix/example-bank-api/lib/validation"

	"github.com/justinas/alice"
)

func Register(
	chain alice.Chain,
	handle httputils.RouterHandler,
	validator *validation.Validator,
	messageQueue *messagequeue.MessageQueue,
	repository repository.Repository,
) error {
	service := &Service{
		MessageQueue: messageQueue,
		Repository:   repository,
	}
	handle("POST", "/signup-request", chain.Append(
		httputils.WithBody(validator, Start{}),
	).ThenFunc(service.Start))

	handle("POST", "/signup-request/:token", chain.Append(
		httputils.WithParams(validator, Finish{}),
	).ThenFunc(service.Finish))
	return nil
}
