package service_test

import (
	"codepix/bank-api/adapters/modifier"
	"codepix/bank-api/adapters/validator"
	"codepix/bank-api/lib/validation"
	"codepix/bank-api/pixkey"
	"codepix/bank-api/pixkey/service"
	proto "codepix/bank-api/proto/codepix/pixkey"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var Validator = func() *validation.Validator {
	val, _ := validator.New()
	service.SetupValidator(val)
	return val
}()

func TestRegisterValidCases(t *testing.T) {
	ID := uuid.New()
	aID := ID[:]

	validCases := []*proto.RegisterRequest{
		{Type: proto.Type(pixkey.CPFKey), Key: " 28556370071", AccountId: aID},
		{Type: proto.Type(pixkey.CPFKey), Key: "28556370071 ", AccountId: aID},
		{Type: proto.Type(pixkey.CPFKey), Key: "28556370071", AccountId: aID},
		{Type: proto.Type(pixkey.CPFKey), Key: "285.563.700-71", AccountId: aID},

		{Type: proto.Type(pixkey.PhoneKey), Key: " +12025550164", AccountId: aID},
		{Type: proto.Type(pixkey.PhoneKey), Key: "+12025550164 ", AccountId: aID},
		{Type: proto.Type(pixkey.PhoneKey), Key: "+12025550164", AccountId: aID},
		{Type: proto.Type(pixkey.PhoneKey), Key: "+120255501641234", AccountId: aID},
		{Type: proto.Type(pixkey.PhoneKey), Key: "+551199887766", AccountId: aID},
		{Type: proto.Type(pixkey.PhoneKey), Key: "+5511999887766", AccountId: aID},

		{Type: proto.Type(pixkey.EmailKey), Key: " name@domain.com", AccountId: aID},
		{Type: proto.Type(pixkey.EmailKey), Key: "name@domain.com ", AccountId: aID},
		{Type: proto.Type(pixkey.EmailKey), Key: "name@domain.com", AccountId: aID},
		{Type: proto.Type(pixkey.EmailKey), Key: "name@subdomain.domain.com", AccountId: aID},
		{Type: proto.Type(pixkey.EmailKey), Key: strings.Repeat("a", 89) + "@domain.com", AccountId: aID},
	}
	for i, tc := range validCases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			trimmedKey := strings.TrimSpace(tc.Key)

			modifier.Mold(tc)
			err := validation.Validate(Validator, tc)

			assert.NoError(t, err)
			assert.Equal(t, trimmedKey, tc.Key)
		})
	}
}

func TestRegisterInvalidCases(t *testing.T) {
	ID := uuid.New()
	aID := ID[:]

	testCases := []*proto.RegisterRequest{
		{},
		{Type: proto.Type(0), Key: "name@domain.com", AccountId: aID},
		{Type: proto.Type(4), Key: "name@domain.com", AccountId: aID},

		{Type: proto.Type(pixkey.CPFKey), Key: "", AccountId: aID},
		{Type: proto.Type(pixkey.CPFKey), Key: "   ", AccountId: aID},
		{Type: proto.Type(pixkey.CPFKey), Key: "123", AccountId: aID},
		{Type: proto.Type(pixkey.CPFKey), Key: "285.563", AccountId: aID},
		{Type: proto.Type(pixkey.CPFKey), Key: "285.563.700-712", AccountId: aID},
		{Type: proto.Type(pixkey.CPFKey), Key: "285563700712", AccountId: aID},
		{Type: proto.Type(pixkey.CPFKey), Key: "111.111.111-11", AccountId: aID},
		{Type: proto.Type(pixkey.CPFKey), Key: "123.456.769/01", AccountId: aID},
		{Type: proto.Type(pixkey.CPFKey), Key: "ABC.DEF.GHI-JK", AccountId: aID},

		{Type: proto.Type(pixkey.PhoneKey), Key: "", AccountId: aID},
		{Type: proto.Type(pixkey.PhoneKey), Key: "   ", AccountId: aID},
		{Type: proto.Type(pixkey.PhoneKey), Key: "+1" + strings.Repeat("2", 99), AccountId: aID},
		{Type: proto.Type(pixkey.PhoneKey), Key: "123", AccountId: aID},
		{Type: proto.Type(pixkey.PhoneKey), Key: "ABCDEFGHIJK", AccountId: aID},
		{Type: proto.Type(pixkey.PhoneKey), Key: "+ABCDEFGHIJK", AccountId: aID},
		{Type: proto.Type(pixkey.PhoneKey), Key: "2025550164", AccountId: aID},
		{Type: proto.Type(pixkey.PhoneKey), Key: "+1202-555-0164", AccountId: aID},
		{Type: proto.Type(pixkey.PhoneKey), Key: "+1-202-555-0164", AccountId: aID},
		{Type: proto.Type(pixkey.PhoneKey), Key: "+1202555016412345", AccountId: aID},

		{Type: proto.Type(pixkey.EmailKey), Key: "", AccountId: aID},
		{Type: proto.Type(pixkey.EmailKey), Key: "   ", AccountId: aID},
		{Type: proto.Type(pixkey.EmailKey), Key: strings.Repeat("a", 90) + "@domain.com", AccountId: aID},
		{Type: proto.Type(pixkey.EmailKey), Key: "name", AccountId: aID},

		{Type: proto.Type(pixkey.EmailKey), Key: "name@domain.com", AccountId: nil},
		{Type: proto.Type(pixkey.EmailKey), Key: "name@domain.com", AccountId: []byte{}},
		{Type: proto.Type(pixkey.EmailKey), Key: "name@domain.com", AccountId: aID[:15]},
		{Type: proto.Type(pixkey.EmailKey), Key: "name@domain.com", AccountId: append(aID, byte(0))},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			trimmedKey := strings.TrimSpace(tc.Key)

			modifier.Mold(tc)
			err := validation.Validate(Validator, tc)

			assert.IsType(t, &validation.Error{}, err)
			assert.Equal(t, trimmedKey, tc.Key)
		})
	}
}
