package usecase_test

import (
	"codepix/bank-api/account/accounttest"
	"codepix/bank-api/lib/repositories"
	"codepix/bank-api/pixkey/interactor"
	"codepix/bank-api/pixkey/pixkeytest"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ValidAccount = accounttest.ValidAccount
var ValidPixKey = pixkeytest.ValidPixKey
var Interactor = pixkeytest.Interactor
var InteractorWithMocks = pixkeytest.InteractorWithMocks

func TestRegister(t *testing.T) {
	type input = interactor.RegisterInput
	type output = interactor.RegisterOutput
	type findAccount = []interface{}
	type add = []interface{}

	type testCase struct {
		description string
		input       input
		findAccount findAccount
		add         add
		err         error
		output      *output
	}

	ID := uuid.New()
	valid := ValidPixKey()
	register := interactor.RegisterInput{
		Type:      valid.Type,
		Key:       valid.Key,
		AccountID: uuid.New(),
	}

	testCases := []testCase{
		{
			"valid",
			register,
			findAccount{nil, nil, nil},
			add{&ID, nil},
			nil,
			&output{PixKey: valid, ID: ID},
		},
		{
			"valid with failure to register",
			register,
			findAccount{nil, nil, nil},
			add{nil, &repositories.InternalError{}},
			&repositories.InternalError{},
			nil,
		},
		{
			"already exists",
			register,
			findAccount{nil, nil, nil},
			add{nil, &repositories.AlreadyExistsError{}},
			&repositories.AlreadyExistsError{},
			nil,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprint(i, "_", tc.description), func(t *testing.T) {
			uc, repo := InteractorWithMocks()

			repo.On("Add", valid, tc.input.AccountID).Return(tc.add...)

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

	uc, repo, creator := Interactor()

	validInput := func() input {
		pk := ValidPixKey()
		accountIDs := creator.AccountIDs(ValidAccount())
		return input{Type: pk.Type, Key: pk.Key, AccountID: accountIDs.AccountID}
	}
	accountAndKey := func() input {
		pk := ValidPixKey()
		pixKeyIDs := creator.PixKeyIDs(pk)
		return input{Type: pk.Type, Key: pk.Key, AccountID: pixKeyIDs.AccountID}
	}

	noPersistCheck := func(t *testing.T, input input, output *output) {}
	shouldPersist := func(t *testing.T, input input, output *output) {
		require.NotNil(t, output)
		require.NotNil(t, output.ID)

		persisted, persistedIDs, err := repo.FindByKey(input.Key)
		require.NoError(t, err)
		assert.Empty(t, cmp.Diff(output.PixKey, *persisted))
		assert.Equal(t, output.ID, persistedIDs.PixKeyID)
		assert.Equal(t, input.AccountID, persistedIDs.AccountID)
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
			accountAndKey,
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
