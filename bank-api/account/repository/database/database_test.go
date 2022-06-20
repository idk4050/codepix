package database_test

import (
	"errors"
	"testing"

	"codepix/bank-api/account/accounttest"
	"codepix/bank-api/account/repository"
	"codepix/bank-api/account/repository/database"
	"codepix/bank-api/lib/repositories"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var ValidAccount = accounttest.ValidAccount
var Repo = accounttest.Repo

func TestAdd(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	repo, _ := Repo()

	account := ValidAccount()
	bankID := uuid.New()

	ID, err := repo.Add(account, bankID)
	assert.NotNil(t, ID)
	assert.NoError(t, err)

	persisted, IDs, err := repo.Find(*ID)
	assert.NoError(t, err)
	assert.Empty(t, cmp.Diff(account, *persisted))
	assert.Empty(t, cmp.Diff(repository.IDs{AccountID: *ID, BankID: bankID}, *IDs))

	ID, err = repo.Add(account, bankID)
	assert.Nil(t, ID)
	assert.IsType(t, &repositories.AlreadyExistsError{}, err)

	repo.(*database.Database).AddError(errors.New("an error"))
	ID, err = repo.Add(account, bankID)
	assert.Nil(t, ID)
	assert.IsType(t, &repositories.InternalError{}, err)
}

func TestFind(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	repo, creator := Repo()

	account := ValidAccount()
	accountIDs := creator.AccountIDs(account)

	persisted, IDs, err := repo.Find(accountIDs.AccountID)
	assert.NoError(t, err)
	assert.Empty(t, cmp.Diff(account, *persisted))
	assert.Empty(t, cmp.Diff(accountIDs, *IDs))

	missingID := uuid.New()
	missing, IDs, err := repo.Find(missingID)
	assert.Nil(t, missing)
	assert.Nil(t, IDs)
	assert.IsType(t, &repositories.NotFoundError{}, err)

	repo.(*database.Database).AddError(errors.New("an error"))
	missing, IDs, err = repo.Find(accountIDs.AccountID)
	assert.Nil(t, missing)
	assert.Nil(t, IDs)
	assert.IsType(t, &repositories.InternalError{}, err)
}

func TestFindByNumber(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	repo, creator := Repo()

	account := ValidAccount()
	accountIDs := creator.AccountIDs(account)

	persisted, IDs, err := repo.FindByNumber(account.Number)
	assert.NoError(t, err)
	assert.Empty(t, cmp.Diff(account, *persisted))
	assert.Empty(t, cmp.Diff(accountIDs, *IDs))

	missingNumber := "123"
	missing, IDs, err := repo.FindByNumber(missingNumber)
	assert.Nil(t, missing)
	assert.Nil(t, IDs)
	assert.IsType(t, &repositories.NotFoundError{}, err)

	missing, IDs, err = repo.FindByNumber("")
	assert.Nil(t, missing)
	assert.Nil(t, IDs)
	assert.IsType(t, &repositories.NotFoundError{}, err)

	repo.(*database.Database).AddError(errors.New("an error"))
	missing, IDs, err = repo.FindByNumber(account.Number)
	assert.Nil(t, missing)
	assert.Nil(t, IDs)
	assert.IsType(t, &repositories.InternalError{}, err)
}

func TestExistsWithBankID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	repo, creator := Repo()

	account := ValidAccount()
	accountIDs := creator.AccountIDs(account)

	err := repo.ExistsWithBankID(accountIDs.AccountID, accountIDs.BankID)
	assert.NoError(t, err)

	err = repo.ExistsWithBankID(uuid.New(), accountIDs.BankID)
	assert.IsType(t, &repositories.NotFoundError{}, err)

	err = repo.ExistsWithBankID(accountIDs.AccountID, uuid.New())
	assert.IsType(t, &repositories.NotFoundError{}, err)

	repo.(*database.Database).AddError(errors.New("an error"))
	err = repo.ExistsWithBankID(accountIDs.AccountID, accountIDs.BankID)
	assert.IsType(t, &repositories.InternalError{}, err)
}
