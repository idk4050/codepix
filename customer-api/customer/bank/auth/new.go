package auth

import (
	"codepix/customer-api/adapters/httputils"
	"codepix/customer-api/config"
	apikeyrepository "codepix/customer-api/customer/bank/apikey/repository"

	"github.com/justinas/alice"
)

func New(
	config config.Config,
	chain alice.Chain,
	handle httputils.RouterHandler,
	apiKeyRepository apikeyrepository.Repository,
) error {
	handle("POST", "/bank-auth", chain.Append(
		AddClaims(apiKeyRepository),
	).ThenFunc(CreateToken(config)))
	return nil
}
