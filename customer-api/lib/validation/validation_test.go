package validation_test

import (
	"codepix/customer-api/lib/validation"
	"fmt"
	"testing"

	en_US_locale "github.com/go-playground/locales/en_US"
	ut "github.com/go-playground/universal-translator"
	en_US_translation "github.com/go-playground/validator/v10/translations/en"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type SimpleStruct struct {
	Field        string
	AnotherField string `validate:"field_tag"`
}

type User struct {
	Username         string `validate:"required"`
	TagNotTranslated string `validate:"tnt"`
	Details          *Details
}
type Details struct {
	FirstName string      `validate:"required"`
	LastName  string      `validate:"required"`
	Addresses *[]*Address `validate:"required,dive,required"`
}
type Address struct {
	FullAddress string `validate:"required"`
	Field       string `validate:"required"`
}
type AddressesMap struct {
	AddressesMap *map[string]*Address `validate:"required,dive,required"`
}

type User2 struct {
	Username string `validate:"test"`
	Details  *Details2
}
type Details2 struct {
	FirstName string       `validate:"test"`
	LastName  string       `validate:"test"`
	Addresses *[]*Address2 `validate:"dive,test"`
}
type Address2 struct {
	FullAddress string `validate:"test"`
	Field       string `validate:"test"`
}

const EN_US = "en_US"

func TestValidation(t *testing.T) {
	var validator = validation.New(nil)

	err := validation.AddValidation(validator, "field_tag", func(field string) bool {
		return field == "test1"
	})
	assert.NoError(t, err)
	assert.True(t, validation.IsValid(validator, "test1", "field_tag"))
	assert.False(t, validation.IsValid(validator, "123", "field_tag"))

	assert.PanicsWithValue(t,
		"Undefined validation function 'struct_tag' on field ''",
		func() {
			validation.IsValid(validator, "test2", "struct_tag")
		})

	validation.AddStructValidations(validator, validation.StructValidation[SimpleStruct]{
		Field:   "Field",
		Tag:     "struct_tag",
		IsValid: func(s *SimpleStruct) bool { return s.Field == "test2" },
	})
	assert.NoError(t, validation.Validate(validator,
		SimpleStruct{Field: "test2", AnotherField: "test1"}))
	assert.Error(t, validation.Validate(validator,
		&SimpleStruct{Field: "test1", AnotherField: "test1"}))

	assert.PanicsWithValue(t,
		"Undefined validation function 'struct_tag' on field ''",
		func() {
			validation.IsValid(validator,
				SimpleStruct{Field: "test2", AnotherField: "test1"}, "struct_tag")
		})
	assert.PanicsWithValue(t,
		"Undefined validation function 'struct_tag' on field ''",
		func() {
			validation.IsValid(validator, "test2", "struct_tag")
		})

	assert.ErrorContains(t,
		validation.AddStructValidations(validator, validation.StructValidation[SimpleStruct]{
			Field:   "Abc",
			Tag:     "struct_tag",
			IsValid: func(s *SimpleStruct) bool { return s.Field == "test2" },
		}),
		"add struct validation: field 'Abc' not found",
	)

	err = validation.Validate(validator, SimpleStruct{})
	assert.IsType(t, &validation.Error{}, err)

	assert.ErrorContains(t, validation.Validate(validator, 123), "validator: (nil int)")
}

func TestErrorMessages(t *testing.T) {
	var TestValidator = validation.New(nil)

	err := validation.AddValidation(TestValidator, "tnt", func(fieldValue any) bool { return fieldValue != "" })
	require.NoError(t, err)

	testCases := []struct {
		err         error
		errorMap    validation.ErrorMap
		errorString string
	}{
		{
			validation.Validate(TestValidator, User{
				"", "", &Details{"", "", &[]*Address{{"", ""}}},
			}),
			validation.ErrorMap{
				"details.addresses[0].field":        "validation failed on details.addresses[0].field (required)",
				"details.addresses[0].full_address": "validation failed on details.addresses[0].full_address (required)",
				"details.first_name":                "validation failed on details.first_name (required)",
				"details.last_name":                 "validation failed on details.last_name (required)",
				"username":                          "validation failed on username (required)",
				"tag_not_translated":                "validation failed on tag_not_translated (tnt)",
			},
			"validation failed on username (required)",
		},
		{
			validation.Validate(TestValidator, Address{"", ""}),
			validation.ErrorMap{
				"field":        "validation failed on field (required)",
				"full_address": "validation failed on full_address (required)",
			},
			"validation failed on full_address (required)",
		},
		{
			validation.Validate(TestValidator, AddressesMap{&map[string]*Address{"a1": {"", ""}}}),
			validation.ErrorMap{
				"addresses_map[a1].field":        "validation failed on addresses_map[a1].field (required)",
				"addresses_map[a1].full_address": "validation failed on addresses_map[a1].full_address (required)",
			},
			"validation failed on addresses_map[a1].full_address (required)",
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			require.IsType(t, &validation.Error{}, tc.err)

			errorMap := tc.err.(*validation.Error).ToErrorMap()
			errorString := tc.err.(*validation.Error).Error()
			assert.Equal(t, tc.errorMap, *errorMap)
			assert.Equal(t, tc.errorString, errorString)
		})
	}
}

func TestTranslatedErrorMessages(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	en_US_translator := en_US_locale.New()
	fallback_translator := en_US_translator

	var universalTranslator = ut.New(
		fallback_translator,
		en_US_translator,
	)
	var validator = validation.New(universalTranslator)

	validation.RegisterTranslations(validator,
		en_US_translation.RegisterDefaultTranslations, EN_US)

	err := validation.AddValidation(validator, "tnt", func(fieldValue any) bool { return fieldValue != "" })
	require.NoError(t, err)

	err = validator.AddErrorMessageTranslation("123", "required", "Test: {0}")
	require.Error(t, err)
	require.Equal(t, err.Error(), "translator with locale 123 not found")

	err = validator.AddErrorMessageTranslation(EN_US, "required", "Test: {0}")
	require.NoError(t, err)

	validator.AddErrorMessageTranslation(EN_US, "test", "Test: {0}")

	err = validation.AddValidation(validator, "test", func(fieldValue any) bool { return false })
	require.NoError(t, err)

	testCases := []struct {
		err      error
		errorMap validation.ErrorMap
	}{
		{
			validation.Validate(validator, User2{
				"1", &Details2{"1", "1", &[]*Address2{{"1", "1"}}},
			}),
			validation.ErrorMap{
				"details.addresses[0].full_address": "Test: details.addresses[0].full_address",
				"details.addresses[0].field":        "Test: details.addresses[0].field",
				"details.first_name":                "Test: details.first_name",
				"details.last_name":                 "Test: details.last_name",
				"username":                          "Test: username",
			},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			require.IsType(t, &validation.Error{}, tc.err)

			errorMap := validator.MapErrors(tc.err.(*validation.Error), EN_US)
			assert.Equal(t, tc.errorMap, errorMap)
		})
	}
}

func TestTranslatedFieldNames(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	en_US_translator := en_US_locale.New()
	fallback_translator := en_US_translator

	var universalTranslator = ut.New(
		fallback_translator,
		en_US_translator,
	)
	var validator = validation.New(universalTranslator)

	validation.RegisterTranslations(validator,
		en_US_translation.RegisterDefaultTranslations, EN_US)

	err := validation.AddValidation(validator, "tnt", func(fieldValue any) bool { return fieldValue != "" })
	require.NoError(t, err)

	err = validator.AddFieldNameTranslation("123", User{}, "Username", "The username")
	require.Error(t, err)
	require.Equal(t, err.Error(), "translator with locale 123 not found")

	err = validator.AddFieldNameTranslation(EN_US, User{}, "Details.FirstName", "First name")
	require.Error(t, err)
	require.Equal(t, err.Error(), "field name translation for struct type User contains invalid namespace: Details.FirstName")

	validator.AddFieldNameTranslation(EN_US, User{Details: &Details{}}, "Details.FirstName", "First name")
	err = validator.AddFieldNameTranslation(EN_US, User{Details: &Details{}}, "Details.FirstName", "First name")
	require.NoError(t, err)

	validator.AddFieldNameTranslation(EN_US,
		User{},
		"Username", "The user's username")

	validator.AddFieldNameTranslation(EN_US,
		User{Details: &Details{
			Addresses: &[]*Address{{}},
		}},
		"Details.Addresses.Field", "Translated Field")

	validator.AddFieldNameTranslation(EN_US,
		Address{},
		"Field", "Field translated to another context")

	validator.AddFieldNameTranslation(EN_US,
		AddressesMap{&map[string]*Address{"": {}}},
		"AddressesMap.Field", "A third context")

	testCases := []struct {
		err         error
		errorMap    validation.ErrorMap
		errorString string
	}{
		{
			validation.Validate(validator, User{
				"", "", &Details{"", "", &[]*Address{{"", ""}}},
			}),
			validation.ErrorMap{
				"details.addresses[0].full_address": "details.addresses[0].full_address is a required field",
				"details.addresses[0].field":        "Translated Field is a required field",
				"details.first_name":                "First name is a required field",
				"details.last_name":                 "details.last_name is a required field",
				"username":                          "The user's username is a required field",
				"tag_not_translated":                "validation failed on tag_not_translated (tnt)",
			},
			"validation failed on username (required)",
		},
		{
			validation.Validate(validator, User{
				"a", "a", &Details{"a", "a", nil},
			}),
			validation.ErrorMap{
				"details.addresses": "details.addresses is a required field",
			},
			"validation failed on details.addresses (required)",
		},
		{
			validation.Validate(validator, Address{"", ""}),
			validation.ErrorMap{
				"full_address": "full_address is a required field",
				"field":        "Field translated to another context is a required field",
			},
			"validation failed on full_address (required)",
		},
		{
			validation.Validate(validator, AddressesMap{&map[string]*Address{"a1": {"", ""}}}),
			validation.ErrorMap{
				"addresses_map[a1].full_address": "addresses_map[a1].full_address is a required field",
				"addresses_map[a1].field":        "A third context is a required field",
			},
			"validation failed on addresses_map[a1].full_address (required)",
		},
		{
			validation.Validate(validator, AddressesMap{&map[string]*Address{"a1": nil}}),
			validation.ErrorMap{
				"addresses_map[a1]": "addresses_map[a1] is a required field",
			},
			"validation failed on addresses_map[a1] (required)",
		},
		{
			validation.Validate(validator, AddressesMap{nil}),
			validation.ErrorMap{
				"addresses_map": "addresses_map is a required field",
			},
			"validation failed on addresses_map (required)",
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			require.IsType(t, &validation.Error{}, tc.err)

			errorMap := validator.MapErrors(tc.err.(*validation.Error), EN_US)
			errorString := tc.err.(*validation.Error).Error()
			assert.Equal(t, tc.errorMap, errorMap)
			assert.Equal(t, tc.errorString, errorString)
		})
	}
}
