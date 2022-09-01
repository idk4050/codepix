package service

import (
	"codepix/customer-api/adapters/httputils"
	"codepix/customer-api/customer/signup/interactor"
	"codepix/customer-api/lib/validation"

	"github.com/justinas/alice"
)

func Register(
	interactor interactor.Interactor,
	chain alice.Chain,
	handle httputils.RouterHandler,
	validator *validation.Validator,
) error {
	service := &Service{
		Interactor: interactor,
	}
	handle("POST", "/signup-request", chain.Append(
		httputils.WithBody(validator, Start{}),
	).ThenFunc(service.Start))

	handle("DELETE", "/signup-request/:token", chain.Append(
		httputils.WithParams(validator, Finish{}),
	).ThenFunc(service.Finish))

	return nil
}
