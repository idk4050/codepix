package validator_test

import (
	"codepix/bank-api/adapters/validator"
	"codepix/bank-api/lib/validation"
	"embed"
	"fmt"
	"os"
	"runtime"
	"testing"

	en_US_translation "github.com/go-playground/validator/v10/translations/en"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type SimpleStruct struct {
	Field        string
	AnotherField string `validate:"field_tag"`
}

type TestStruct struct {
	TestField1         string                   `validate:"test_tag1"`
	Struct2            TestNestedStruct         `validate:"test_tag2"`
	PointerStruct3     *TestPointerStruct       `validate:"test_tag3"`
	Array4             []TestStructArray        `validate:"test_tag4"`
	Map5               map[string]TestStructMap `validate:"test_tag5"`
	NonTranslatedField string                   `validate:"required"`
	TranslatedField    string                   `validate:"required"`
}
type TestNestedStruct struct {
	TestField6 string `validate:"test_tag6"`
}
type TestPointerStruct struct {
	TestField7 string `validate:"test_tag7"`
}
type TestStructArray struct {
	TestField8 string `validate:"test_tag8"`
}
type TestStructMap struct {
	TestField9 string `validate:"test_tag9"`
}

//go:embed testdata
var testData embed.FS

func TestLoadTranslationFile(t *testing.T) {
	st := TestStruct{PointerStruct3: &TestPointerStruct{}}

	testValidator, err := validator.New()
	require.NoError(t, err)

	_, thisFilePath, _, _ := runtime.Caller(0)
	thisfile, err := os.Open(thisFilePath)
	require.NoError(t, err)

	test, err := testData.Open("testdata/test.json")
	require.NoError(t, err)
	test2, err := testData.Open("testdata/test2.json")
	require.NoError(t, err)
	test3, err := testData.Open("testdata/test3.json")
	require.NoError(t, err)
	test4, err := testData.Open("testdata/test4.json")
	require.NoError(t, err)
	test5, err := testData.Open("testdata/test5.json")
	require.NoError(t, err)
	test6, err := testData.Open("testdata/test6.json")
	require.NoError(t, err)

	assert.ErrorContains(t,
		validator.LoadTranslationFile(testValidator, thisfile, st),
		"translation file could not be parsed: invalid character 'p' looking for beginning of value",
	)
	assert.ErrorContains(t,
		validator.LoadTranslationFile(testValidator, test2, st),
		"translation file does not contain struct type TestStruct",
	)
	assert.ErrorContains(t,
		validator.LoadTranslationFile(testValidator, test3, st),
		"translation file contains invalid locale TestStruct.123",
	)
	assert.ErrorContains(t,
		validator.LoadTranslationFile(testValidator, test4, st),
		"translation file contains invalid category TestStruct.en_US.123",
	)
	assert.ErrorContains(t,
		validator.LoadTranslationFile(testValidator, test5, st),
		"could not add error message translation for TestStruct.en_US.error_messages.test_tag1: "+
			"error: missing bracket '{}', in translation. locale: 'en_US' key: 'test_tag1' text: 'broken field name template {0'",
	)
	assert.ErrorContains(t,
		validator.LoadTranslationFile(testValidator, test6, st),
		"could not add field name translation for TestStruct.en_US.field_names.NonExistingField: "+
			"field name translation for struct type TestStruct contains invalid namespace: NonExistingField",
	)

	err = validation.AddValidation(testValidator, "field_tag", func(field string) bool {
		return field != ""
	})
	require.NoError(t, err)

	st2 := SimpleStruct{}
	err = validator.LoadTranslationFile(testValidator, test, st, st2)
	require.NoError(t, err)

	containsStuff := func(field interface{}) bool {
		switch field := field.(type) {
		case string:
			return field != ""
		case TestNestedStruct:
			return field != TestNestedStruct{}
		case *TestPointerStruct:
			return field != nil
		case []TestStructArray:
			return len(field) > 0
		case map[string]TestStructMap:
			return len(field) > 0
		default:
			return false
		}
	}

	for i := 1; i < 10; i++ {
		err := validation.AddValidation(testValidator, "test_tag"+fmt.Sprint(i), containsStuff)
		require.NoError(t, err)
	}

	testCases := []struct {
		err      error
		errorMap validation.ErrorMap
	}{
		{
			validation.Validate(testValidator, TestStruct{}),
			validation.ErrorMap{
				"translated_field":      "The Translated Field is a required field",
				"non_translated_field":  "non_translated_field is a required field",
				"array_4":               "4: It failed on field The Array!",
				"map_5":                 "5: It failed on field The Map!",
				"pointer_struct_3":      "3: It failed on field The PointerStruct!",
				"struct_2.test_field_6": "6: It failed on field The Field 6!",
				"test_field_1":          "1: It failed on field The Field 1!",
			},
		},
		{
			validation.Validate(testValidator, TestStruct{
				TestField1:     "",
				Struct2:        TestNestedStruct{},
				PointerStruct3: &TestPointerStruct{},
				Array4:         []TestStructArray{{}},
				Map5:           map[string]TestStructMap{"A": {}},
			}),
			validation.ErrorMap{
				"translated_field":              "The Translated Field is a required field",
				"non_translated_field":          "non_translated_field is a required field",
				"pointer_struct_3.test_field_7": "7: It failed on field The Field 7!",
				"struct_2.test_field_6":         "6: It failed on field The Field 6!",
				"test_field_1":                  "1: It failed on field The Field 1!",
			},
		},
	}

	validation.RegisterTranslations(testValidator,
		en_US_translation.RegisterDefaultTranslations, validator.EN_US)

	for i, tc := range testCases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			errorMap := testValidator.MapErrors(tc.err.(*validation.Error), validator.EN_US)
			assert.Equal(t, tc.errorMap, errorMap)
		})
	}
}
