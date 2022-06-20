package pixkeytest

import (
	"codepix/bank-api/adapters/databaseclient"
	"codepix/bank-api/adapters/validator"
	"codepix/bank-api/bankapitest"
	"codepix/bank-api/pixkey"
	"codepix/bank-api/pixkey/repository"
	"codepix/bank-api/pixkey/repository/database"
	"codepix/bank-api/pixkey/service"
	proto "codepix/bank-api/proto/codepix/pixkey"
	"strings"

	"github.com/google/uuid"
)

func ValidPixKey() pixkey.PixKey {
	uniqueKey := uuid.NewString() + "@domain.com"
	return pixkey.PixKey{
		Type: pixkey.EmailKey,
		Key:  uniqueKey,
	}
}
func InvalidPixKey() pixkey.PixKey {
	return pixkey.PixKey{
		Type: 0,
		Key:  strings.Repeat("A", 101),
	}
}

type Creator struct {
	PixKeyIDs func(pixkey.PixKey) repository.IDs
}

func Repo() (repository.Repository, Creator) {
	client, err := databaseclient.Open(bankapitest.Config, bankapitest.Logger)
	if err != nil {
		panic(err)
	}
	err = client.AutoMigrate(
		&database.PixKey{},
	)
	if err != nil {
		panic(err)
	}
	repo := &database.Database{Database: client}
	return repo, Creator{
		PixKeyIDs(repo),
	}
}

func PixKeyIDs(repo repository.Repository) func(pixkey.PixKey) repository.IDs {
	return func(pk pixkey.PixKey) repository.IDs {
		accountID, bankID := uuid.New(), uuid.New()
		ID, _ := repo.Add(pk, accountID, bankID)
		return repository.IDs{
			PixKeyID:  *ID,
			AccountID: accountID,
			BankID:    bankID,
		}
	}
}

func Service() (proto.ServiceClient, repository.Repository, Creator) {
	validator, err := validator.New()
	if err != nil {
		panic(err)
	}
	server, client, serve := bankapitest.Server(validator)
	repo, creator := Repo()

	err = service.Register(server, validator, repo)
	if err != nil {
		panic(err)
	}
	serve()
	return proto.NewServiceClient(client), repo, creator
}
func ServiceWithMocks() (proto.ServiceClient, *MockRepo) {
	validator, err := validator.New()
	if err != nil {
		panic(err)
	}
	server, client, serve := bankapitest.Server(validator)
	repo := new(MockRepo)

	err = service.Register(server, validator, repo)
	if err != nil {
		panic(err)
	}
	serve()
	return proto.NewServiceClient(client), repo
}
