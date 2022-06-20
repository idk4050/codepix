package usecase_test

import (
	"codepix/bank-api/account/accounttest"
	"codepix/bank-api/account/interactor"
	"codepix/bank-api/account/repository"
	"codepix/bank-api/lib/repositories"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ValidAccount = accounttest.ValidAccount
var InvalidAccount = accounttest.InvalidAccount
var Interactor = accounttest.Interactor
var InteractorWithMocks = accounttest.InteractorWithMocks

func TestRegister(t *testing.T) {
	type input = interactor.RegisterInput
	type output = interactor.RegisterOutput
	type findBank = []interface{}
	type add = []interface{}

	type testCase struct {
		description string
		input       input
		add         add
		err         error
		output      *output
	}

	ID := uuid.New()
	valid := ValidAccount()
	register := interactor.RegisterInput{
		Number:    valid.Number,
		OwnerName: valid.OwnerName,
		BankID:    uuid.New(),
	}

	testCases := []testCase{
		{
			"valid",
			register,
			add{&ID, nil},
			nil,
			&output{Account: valid, ID: ID},
		},
		{
			"fail to register",
			register,
			add{nil, &repositories.InternalError{}},
			&repositories.InternalError{},
			nil,
		},
		{
			"already exists",
			register,
			add{nil, &repositories.AlreadyExistsError{}},
			&repositories.AlreadyExistsError{},
			nil,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprint(i, "_", tc.description), func(t *testing.T) {
			uc, repo := InteractorWithMocks()

			repo.On("Add", valid, tc.input.BankID).Return(tc.add...)

			output, err := uc.Register(tc.input)

			if tc.err == nil {
				assert.NotNil(t, output)
				assert.NoError(t, err)
				assert.Empty(t, cmp.Diff(tc.output, output))
			} else {
				assert.Nil(t, output)
				assert.IsType(t, tc.err, err)
			}
		})
	}
}

func TestRegisterIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	type input = interactor.RegisterInput
	type output = interactor.RegisterOutput

	type testCase struct {
		description  string
		input        func() input
		err          error
		persistCheck func(t *testing.T, input input, output *output)
	}

	uc, repo := Interactor()

	validInput := func() input {
		account := ValidAccount()
		bankID := uuid.New()
		return input{Number: account.Number, OwnerName: account.OwnerName, BankID: bankID}
	}
	existingAccount := func() input {
		account := ValidAccount()
		bankID := uuid.New()
		repo.Add(account, bankID)
		return input{Number: account.Number, OwnerName: account.OwnerName, BankID: bankID}
	}

	noPersistCheck := func(t *testing.T, input input, output *output) {}
	shouldPersist := func(t *testing.T, input input, output *output) {
		require.NotNil(t, output)
		require.NotNil(t, output.ID)

		persisted, persistedIDs, err := repo.Find(output.ID)
		require.NoError(t, err)
		assert.Empty(t, cmp.Diff(output.Account, *persisted))
		assert.Empty(t, cmp.Diff(repository.IDs{AccountID: output.ID, BankID: input.BankID}, *persistedIDs))
	}

	testCases := []testCase{
		{
			"valid",
			validInput,
			nil,
			shouldPersist,
		},
		{
			"already exists",
			existingAccount,
			&repositories.AlreadyExistsError{},
			noPersistCheck,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprint(i, "_", tc.description), func(t *testing.T) {
			input := tc.input()

			output, err := uc.Register(input)

			if tc.err == nil {
				assert.NotNil(t, output)
				assert.NoError(t, err)
			} else {
				assert.Nil(t, output)
				assert.IsType(t, tc.err, err)
			}
			tc.persistCheck(t, input, output)
		})
	}
}
