package service

import (
	"codepix/example-bank-api/adapters/httputils"
	"codepix/example-bank-api/adapters/pix"
	"codepix/example-bank-api/config"
	"codepix/example-bank-api/customer/account/repository"
	customerauth "codepix/example-bank-api/customer/auth"
	"codepix/example-bank-api/lib/validation"
	readproto "codepix/example-bank-api/proto/codepix/transaction/read"
	writeproto "codepix/example-bank-api/proto/codepix/transaction/write"
	userauth "codepix/example-bank-api/user/auth"

	"github.com/justinas/alice"
)

func Register(
	config config.Config,
	chain alice.Chain,
	handle httputils.RouterHandler,
	validator *validation.Validator,
	accountRepository repository.Repository,
	pixAPIClient *pix.Client,
) error {
	service := &Service{
		WriteServiceClient: writeproto.NewServiceClient(pixAPIClient.Conn),
		ReadServiceClient:  readproto.NewServiceClient(pixAPIClient.Conn),
	}
	anyUser := chain.Append(userauth.ValidateToken(config))
	anyCustomer := anyUser.Append(customerauth.ValidateClaims)
	accountOwner := anyCustomer.Append(customerauth.CustomerOwnsParamAccountID(
		accountRepository, "account-id"))

	handle("POST", "/account/:account-id/pix/transactions", accountOwner.Append(
		httputils.WithParams(validator, SendParams{}),
		httputils.WithBody(validator, Send{}),
	).ThenFunc(service.Send))

	handle("GET", "/account/:account-id/pix/transactions/:pix-transaction-id", accountOwner.Append(
		httputils.WithParams(validator, Find{}),
	).ThenFunc(service.Find))

	handle("GET", "/account/:account-id/pix/transactions", accountOwner.Append(
		httputils.WithParams(validator, List{}),
	).ThenFunc(service.List))
	return nil
}
