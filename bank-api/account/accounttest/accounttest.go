package accounttest

import (
	"codepix/bank-api/account"
	"codepix/bank-api/account/interactor"
	"codepix/bank-api/account/interactor/usecase"
	"codepix/bank-api/account/repository"
	"codepix/bank-api/account/repository/database"
	"codepix/bank-api/account/service"
	"codepix/bank-api/account/service/proto"
	"codepix/bank-api/adapters/databaseclient"
	"codepix/bank-api/adapters/validator"
	"codepix/bank-api/bankapitest"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

func ValidAccount() account.Account {
	number := fmt.Sprint(random.Uint32())
	ownerName := fmt.Sprint("Account Owner ", number)
	return account.Account{Number: number, OwnerName: ownerName}
}
func InvalidAccount() account.Account {
	return account.Account{Number: "", OwnerName: ""}
}

type Creator struct {
	AccountIDs func(account.Account) repository.IDs
}

func Repo() (repository.Repository, Creator) {
	client, err := databaseclient.Open(bankapitest.Config, bankapitest.Logger)
	if err != nil {
		panic(err)
	}
	err = client.AutoMigrate(
		&database.Account{},
	)
	if err != nil {
		panic(err)
	}
	repo := &database.Database{DB: client}
	return repo, Creator{
		AccountIDs(repo),
	}
}

func AccountIDs(repo repository.Repository) func(account.Account) repository.IDs {
	return func(account account.Account) repository.IDs {
		bankID := uuid.New()
		accountID, _ := repo.Add(account, bankID)
		return repository.IDs{
			AccountID: *accountID,
			BankID:    bankID,
		}
	}
}

func Interactor() (interactor.Interactor, repository.Repository) {
	repo, _ := Repo()
	return &usecase.Usecase{
		Repository: repo,
	}, repo
}
func InteractorWithMocks() (interactor.Interactor, *MockRepo) {
	repo := new(MockRepo)
	return &usecase.Usecase{
		Repository: repo,
	}, repo
}

func Service() (proto.AccountServiceClient, repository.Repository) {
	validator, err := validator.New()
	if err != nil {
		panic(err)
	}
	server, client, serve := bankapitest.Server(validator)
	interactor, repo := Interactor()

	err = service.Register(server, validator, interactor, repo)
	if err != nil {
		panic(err)
	}
	serve()
	return proto.NewAccountServiceClient(client), repo
}
func ServiceWithMocks() (proto.AccountServiceClient, *MockInteractor, *MockRepo) {
	validator, err := validator.New()
	if err != nil {
		panic(err)
	}
	server, client, serve := bankapitest.Server(validator)
	interactor := new(MockInteractor)
	repo := new(MockRepo)

	err = service.Register(server, validator, interactor, repo)
	if err != nil {
		panic(err)
	}
	serve()
	return proto.NewAccountServiceClient(client), interactor, repo
}
