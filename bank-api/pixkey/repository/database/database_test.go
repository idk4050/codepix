package database_test

import (
	"errors"
	"testing"

	"codepix/bank-api/account/accounttest"
	"codepix/bank-api/lib/repositories"
	"codepix/bank-api/pixkey/pixkeytest"
	"codepix/bank-api/pixkey/repository"
	"codepix/bank-api/pixkey/repository/database"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var ValidAccount = accounttest.ValidAccount
var ValidPixKey = pixkeytest.ValidPixKey
var Repo = pixkeytest.Repo

func TestAdd(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	repo, creator := Repo()

	pixKey := ValidPixKey()
	accountIDs := creator.AccountIDs(ValidAccount())

	ID, err := repo.Add(pixKey, accountIDs.AccountID)
	assert.NotNil(t, ID)
	assert.NoError(t, err)

	persisted, IDs, err := repo.FindByKey(pixKey.Key)
	assert.NoError(t, err)
	assert.Empty(t, cmp.Diff(pixKey, *persisted))
	assert.Empty(t, cmp.Diff(repository.IDs{
		PixKeyID:  *ID,
		AccountID: accountIDs.AccountID,
		BankID:    accountIDs.BankID,
	}, *IDs))

	persisted, IDs, err = repo.Find(*ID)
	assert.NoError(t, err)
	assert.Empty(t, cmp.Diff(pixKey, *persisted))
	assert.Empty(t, cmp.Diff(repository.IDs{
		PixKeyID:  *ID,
		AccountID: accountIDs.AccountID,
		BankID:    accountIDs.BankID,
	}, *IDs))

	ID, err = repo.Add(pixKey, accountIDs.AccountID)
	assert.Nil(t, ID)
	assert.IsType(t, &repositories.AlreadyExistsError{}, err)

	ID, err = repo.Add(pixKey, uuid.New())
	assert.Nil(t, ID)
	assert.IsType(t, &repositories.AlreadyExistsError{}, err)

	repo.(*database.Database).AddError(errors.New("an error"))
	ID, err = repo.Add(pixKey, accountIDs.AccountID)
	assert.Nil(t, ID)
	assert.IsType(t, &repositories.InternalError{}, err)
}

func TestFind(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	repo, creator := Repo()

	pixKey := ValidPixKey()
	pixKeyIDs := creator.PixKeyIDs(pixKey)

	persisted, IDs, err := repo.Find(pixKeyIDs.PixKeyID)
	assert.NoError(t, err)
	assert.Empty(t, cmp.Diff(pixKey, *persisted))
	assert.Empty(t, cmp.Diff(pixKeyIDs, *IDs))

	missingID := uuid.New()
	missing, IDs, err := repo.Find(missingID)
	assert.Nil(t, missing)
	assert.Nil(t, IDs)
	assert.IsType(t, &repositories.NotFoundError{}, err)

	repo.(*database.Database).AddError(errors.New("an error"))
	missing, IDs, err = repo.Find(pixKeyIDs.PixKeyID)
	assert.Nil(t, missing)
	assert.Nil(t, IDs)
	assert.IsType(t, &repositories.InternalError{}, err)
}

func TestFindByKey(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	repo, creator := Repo()

	pixKey := ValidPixKey()
	pixKeyIDs := creator.PixKeyIDs(pixKey)

	persisted, IDs, err := repo.FindByKey(pixKey.Key)
	assert.NoError(t, err)
	assert.Empty(t, cmp.Diff(pixKey, *persisted))
	assert.Empty(t, cmp.Diff(pixKeyIDs, *IDs))

	missingKey := "123"
	missing, IDs, err := repo.FindByKey(missingKey)
	assert.Nil(t, missing)
	assert.Nil(t, IDs)
	assert.IsType(t, &repositories.NotFoundError{}, err)

	missing, IDs, err = repo.FindByKey("")
	assert.Nil(t, missing)
	assert.Nil(t, IDs)
	assert.IsType(t, &repositories.NotFoundError{}, err)

	repo.(*database.Database).AddError(errors.New("an error"))
	missing, IDs, err = repo.FindByKey(pixKey.Key)
	assert.Nil(t, missing)
	assert.Nil(t, IDs)
	assert.IsType(t, &repositories.InternalError{}, err)
}
