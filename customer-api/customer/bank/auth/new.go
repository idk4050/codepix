package auth

import (
	"codepix/customer-api/adapters/httputils"
	"codepix/customer-api/config"
	apikeyrepository "codepix/customer-api/customer/bank/apikey/repository"
	"codepix/customer-api/lib/validation"

	"github.com/justinas/alice"
)

func New(
	config config.Config,
	chain alice.Chain,
	handle httputils.RouterHandler,
	validator *validation.Validator,
	apiKeyRepository apikeyrepository.Repository,
) error {
	handle("POST", "/bank-auth", chain.Append(
		httputils.WithBody(validator, Authenticate{}),
		AddClaims(apiKeyRepository),
	).ThenFunc(CreateToken(config)))

	return nil
}
