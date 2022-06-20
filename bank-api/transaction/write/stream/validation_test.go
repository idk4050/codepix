package stream_test

import (
	"codepix/bank-api/adapters/modifier"
	"codepix/bank-api/adapters/validator"
	"codepix/bank-api/lib/validation"
	"codepix/bank-api/pixkey/pixkeytest"
	proto "codepix/bank-api/proto/codepix/transaction/write"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var Validator, _ = validator.New()

var ValidPixKey = pixkeytest.ValidPixKey

func TestStartValidCases(t *testing.T) {
	sender := uuid.New()
	receiver := ValidPixKey().Key

	A100x := strings.Repeat("A", 100)

	startCases := []*proto.StartRequest{
		{SenderId: sender[:], ReceiverKey: receiver, Amount: 100, Description: "Test description"},
		{SenderId: sender[:], ReceiverKey: receiver, Amount: 100, Description: ""},
		{SenderId: sender[:], ReceiverKey: receiver, Amount: 100, Description: " Test description 	"},
		{SenderId: sender[:], ReceiverKey: receiver, Amount: 100, Description: A100x},
	}
	for i, tc := range startCases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			trimmedDescription := strings.TrimSpace(tc.Description)

			modifier.Mold(tc)
			err := validation.Validate(Validator, tc)

			assert.NoError(t, err)
			assert.Equal(t, trimmedDescription, tc.Description)
		})
	}
}

func TestStartInvalidCases(t *testing.T) {
	sender := uuid.New()
	receiver := ValidPixKey().Key

	A101x := strings.Repeat("A", 101)

	startCases := []*proto.StartRequest{
		{SenderId: nil, ReceiverKey: receiver, Amount: 100, Description: "Test description"},
		{SenderId: []byte{}, ReceiverKey: receiver, Amount: 100, Description: "Test description"},
		{SenderId: sender[:15], ReceiverKey: receiver, Amount: 100, Description: "Test description"},
		{SenderId: append(sender[:], byte(0)), ReceiverKey: receiver, Amount: 100, Description: "Test description"},
		{SenderId: sender[:], ReceiverKey: "", Amount: 100, Description: "Test description"},
		{SenderId: sender[:], ReceiverKey: receiver, Amount: 0, Description: "Test description"},
		{SenderId: sender[:], ReceiverKey: receiver, Amount: 100, Description: A101x},
	}
	for i, tc := range startCases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			trimmedDescription := strings.TrimSpace(tc.Description)

			modifier.Mold(tc)
			err := validation.Validate(Validator, tc)

			assert.IsType(t, &validation.Error{}, err)
			assert.Equal(t, trimmedDescription, tc.Description)
		})
	}
}
