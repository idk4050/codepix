package database_test

import (
	"errors"
	"testing"

	"codepix/bank-api/lib/repositories"
	"codepix/bank-api/pixkey"
	"codepix/bank-api/pixkey/pixkeytest"
	"codepix/bank-api/pixkey/repository"
	"codepix/bank-api/pixkey/repository/database"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var ValidPixKey = pixkeytest.ValidPixKey
var Repo = pixkeytest.Repo

func TestAdd(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	repo, _ := Repo()

	pixKey := ValidPixKey()
	accountID, bankID := uuid.New(), uuid.New()

	ID, err := repo.Add(pixKey, accountID, bankID)
	assert.NotNil(t, ID)
	assert.NoError(t, err)

	persisted, IDs, err := repo.FindByKey(pixKey.Key)
	assert.NoError(t, err)
	assert.Empty(t, cmp.Diff(pixKey, *persisted))
	assert.Empty(t, cmp.Diff(repository.IDs{
		PixKeyID:  *ID,
		AccountID: accountID,
		BankID:    bankID,
	}, *IDs))

	persisted, IDs, err = repo.Find(*ID)
	assert.NoError(t, err)
	assert.Empty(t, cmp.Diff(pixKey, *persisted))
	assert.Empty(t, cmp.Diff(repository.IDs{
		PixKeyID:  *ID,
		AccountID: accountID,
		BankID:    bankID,
	}, *IDs))

	ID, err = repo.Add(pixKey, accountID, bankID)
	assert.Nil(t, ID)
	assert.IsType(t, &repositories.AlreadyExistsError{}, err)

	ID, err = repo.Add(pixKey, uuid.New(), uuid.New())
	assert.Nil(t, ID)
	assert.IsType(t, &repositories.AlreadyExistsError{}, err)

	repo.(*database.Database).AddError(errors.New("an error"))
	ID, err = repo.Add(pixKey, accountID, bankID)
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

func TestList(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	repo, _ := Repo()

	accountID, bankID := uuid.New(), uuid.New()

	nPixKeys := 10
	pixKeys := []pixkey.PixKey{}
	IDs := []uuid.UUID{}
	for i := 0; i < nPixKeys; i++ {
		pixKey := ValidPixKey()
		pixKeys = append(pixKeys, pixKey)
		ID, _ := repo.Add(pixKey, accountID, bankID)
		IDs = append(IDs, *ID)
	}
	for i := 0; i < nPixKeys; i++ {
		repo.Add(ValidPixKey(), uuid.New(), uuid.New())
	}
	options := repository.ListOptions{
		AccountID: accountID,
		BankID:    bankID,
	}

	persisted, err := repo.List(options)
	assert.NoError(t, err)
	assert.Len(t, persisted, nPixKeys)
	for i := 0; i < nPixKeys; i++ {
		expected := repository.ListItem{
			ID:   IDs[i],
			Type: pixKeys[i].Type,
			Key:  pixKeys[i].Key,
		}
		assert.Empty(t, cmp.Diff(persisted[i], expected))
	}

	missingAccountID := uuid.New()
	missing, err := repo.List(repository.ListOptions{
		AccountID: missingAccountID,
		BankID:    bankID,
	})
	assert.NoError(t, err)
	assert.NotNil(t, missing)
	assert.Empty(t, missing)

	missingBankID := uuid.New()
	missing, err = repo.List(repository.ListOptions{
		AccountID: accountID,
		BankID:    missingBankID,
	})
	assert.NoError(t, err)
	assert.NotNil(t, missing)
	assert.Empty(t, missing)

	repo.(*database.Database).AddError(errors.New("an error"))
	missing, err = repo.List(options)
	assert.Nil(t, missing)
	assert.IsType(t, &repositories.InternalError{}, err)
}
