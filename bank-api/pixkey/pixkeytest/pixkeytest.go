package pixkeytest

import (
	"codepix/bank-api/account"
	"codepix/bank-api/account/accounttest"
	accountrepository "codepix/bank-api/account/repository"
	accountdatabase "codepix/bank-api/account/repository/database"
	"codepix/bank-api/adapters/databaseclient"
	"codepix/bank-api/adapters/validator"
	"codepix/bank-api/bankapitest"
	"codepix/bank-api/pixkey"
	"codepix/bank-api/pixkey/interactor"
	"codepix/bank-api/pixkey/interactor/usecase"
	"codepix/bank-api/pixkey/repository"
	"codepix/bank-api/pixkey/repository/database"
	"codepix/bank-api/pixkey/service"
	"codepix/bank-api/pixkey/service/proto"

	"github.com/google/uuid"
)

func ValidPixKey() pixkey.PixKey {
	uniqueKey := uuid.NewString() + "@domain.com"
	return pixkey.PixKey{Type: pixkey.EmailKey, Key: uniqueKey}
}
func InvalidPixKey() pixkey.PixKey {
	return pixkey.PixKey{Type: 0, Key: ""}
}

type Creator struct {
	PixKeyIDs  func(pixkey.PixKey) repository.IDs
	AccountIDs func(account.Account) accountrepository.IDs
}

func Repo() (repository.Repository, Creator) {
	client, err := databaseclient.Open(bankapitest.Config, bankapitest.Logger)
	if err != nil {
		panic(err)
	}
	err = client.AutoMigrate(
		&accountdatabase.Account{},
		&database.PixKey{},
	)
	if err != nil {
		panic(err)
	}
	repo := &database.Database{DB: client}
	accountRepo := &accountdatabase.Database{DB: client}
	return repo, Creator{
		PixKeyIDs(repo, accountRepo),
		accounttest.AccountIDs(accountRepo),
	}
}

func PixKeyIDs(repo repository.Repository, accountRepo accountrepository.Repository,
) func(pixkey.PixKey) repository.IDs {
	return func(pk pixkey.PixKey) repository.IDs {
		account := accounttest.ValidAccount()
		accountIDs := accounttest.AccountIDs(accountRepo)(account)
		ID, _ := repo.Add(pk, accountIDs.AccountID)
		return repository.IDs{
			PixKeyID:  *ID,
			AccountID: accountIDs.AccountID,
			BankID:    accountIDs.BankID,
		}
	}
}

func Interactor() (interactor.Interactor, repository.Repository, Creator) {
	repo, creator := Repo()
	client := repo.(*database.Database).DB
	accountRepo := &accountdatabase.Database{DB: client}
	return &usecase.Usecase{
		Repository:        repo,
		AccountRepository: accountRepo,
	}, repo, creator
}
func InteractorWithMocks() (interactor.Interactor, *MockRepo) {
	repo := new(MockRepo)
	accountRepo := new(accounttest.MockRepo)
	return &usecase.Usecase{
		Repository:        repo,
		AccountRepository: accountRepo,
	}, repo
}

func Service() (proto.PixKeyServiceClient, repository.Repository, Creator) {
	validator, err := validator.New()
	if err != nil {
		panic(err)
	}
	server, client, serve := bankapitest.Server(validator)
	interactor, repo, creator := Interactor()
	database := repo.(*database.Database).DB
	accountRepo := &accountdatabase.Database{DB: database}

	err = service.Register(server, validator, interactor, repo, accountRepo)
	if err != nil {
		panic(err)
	}
	serve()
	return proto.NewPixKeyServiceClient(client), repo, creator
}
func ServiceWithMocks() (proto.PixKeyServiceClient, *MockInteractor, *MockRepo, *accounttest.MockRepo) {
	validator, err := validator.New()
	if err != nil {
		panic(err)
	}
	server, client, serve := bankapitest.Server(validator)
	interactor := new(MockInteractor)
	repo := new(MockRepo)
	accountRepo := new(accounttest.MockRepo)

	err = service.Register(server, validator, interactor, repo, accountRepo)
	if err != nil {
		panic(err)
	}
	serve()
	return proto.NewPixKeyServiceClient(client), interactor, repo, accountRepo
}
