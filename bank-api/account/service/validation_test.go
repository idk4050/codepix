package service_test

import (
	"codepix/bank-api/account/service/proto"
	"codepix/bank-api/adapters/modifier"
	"codepix/bank-api/adapters/validator"
	"codepix/bank-api/lib/validation"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var Validator, _ = validator.New()

func TestRegisterValidCases(t *testing.T) {
	A100x := strings.Repeat("A", 100)

	validCases := []*proto.RegisterRequest{
		{Number: "1", OwnerName: "Owner"},
		{Number: "1-A", OwnerName: "Owner"},
		{Number: " 1 ", OwnerName: "Owner"},
		{Number: A100x, OwnerName: "Owner"},
		{Number: "1", OwnerName: " Owner "},
		{Number: "1", OwnerName: "Owner Name"},
		{Number: "1", OwnerName: A100x},
	}
	for i, tc := range validCases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			trimmedNumber := strings.TrimSpace(tc.Number)
			trimmedOwnerName := strings.TrimSpace(tc.OwnerName)

			modifier.Mold(tc)
			err := validation.Validate(Validator, tc)

			assert.NoError(t, err)
			assert.Equal(t, trimmedNumber, tc.Number)
			assert.Equal(t, trimmedOwnerName, tc.OwnerName)
		})
	}
}

func TestRegisterInvalidCases(t *testing.T) {
	A101x := strings.Repeat("A", 101)

	invalidCases := []*proto.RegisterRequest{
		{Number: "", OwnerName: "Owner"},
		{Number: A101x, OwnerName: "Owner"},
		{Number: "1", OwnerName: ""},
		{Number: "1", OwnerName: " 	"},
		{Number: "1", OwnerName: A101x},
	}
	for i, tc := range invalidCases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			trimmedNumber := strings.TrimSpace(tc.Number)
			trimmedOwnerName := strings.TrimSpace(tc.OwnerName)

			modifier.Mold(tc)
			err := validation.Validate(Validator, tc)

			assert.IsType(t, &validation.Error{}, err)
			assert.Equal(t, trimmedNumber, tc.Number)
			assert.Equal(t, trimmedOwnerName, tc.OwnerName)
		})
	}
}
