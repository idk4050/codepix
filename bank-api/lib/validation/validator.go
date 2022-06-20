package validation

import (
	"fmt"
	"reflect"
	"strings"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

type InvalidValidationError = validator.InvalidValidationError
type FieldLevel = validator.FieldLevel

// Validator is a custom wrapper around the library's validator type.
type Validator struct {
	validator           *validator.Validate
	universalTranslator *ut.UniversalTranslator
}

func New(ut *ut.UniversalTranslator) *Validator {
	v := validator.New()
	v.RegisterTagNameFunc(formatFieldNameForErrorMessage)
	return &Validator{v, ut}
}

// for struct fields, the struct's type name is added to the FieldError's namespace
func formatFieldNameForErrorMessage(field reflect.StructField) string {
	fieldName := field.Name
	typ := field.Type
	kind := typ.Kind()

	if kind == reflect.Struct {
		fieldName = fmt.Sprintf("%s/%s", fieldName, typ.Name())
	} else if kind == reflect.Pointer && typ.Elem().Kind() == reflect.Struct {
		fieldName = fmt.Sprintf("%s/%s", fieldName, typ.Elem().Name())
	}

	return wrapFieldName(fieldName)
}

// Validate validates a struct's exposed fields including nested structs.
//
// It returns validation.InvalidValidationError if st isn't a struct or a pointer to a struct,
// validation.Error if the validation failed and nil otherwise.
func Validate[T any](v *Validator, st T) error {
	err := v.validator.Struct(st)

	switch err := err.(type) {
	case *validator.InvalidValidationError:
		return (*InvalidValidationError)(err)
	case validator.ValidationErrors:
		if len(err) > 0 {
			return &Error{err}
		} else {
			return nil
		}
	default:
		return nil
	}
}

// IsValid validates a single value against a tag.
func IsValid[T any](v *Validator, value T, tag string) bool {
	return v.validator.Var(value, tag) == nil
}

// AddValidation adds a field-level validation for the given tag.
//
// If the tag already exists, the previous validation function will be replaced.
func AddValidation[T any](v *Validator, tag string, isValid func(fieldValue T) bool) error {
	return v.validator.RegisterValidation(tag, func(field validator.FieldLevel) bool {
		fieldValue := field.Field().Interface().(T)
		return isValid(fieldValue)
	})
}

type StructValidation[T any] struct {
	Field   string
	Tag     string
	IsValid func(st *T) bool
}

// AddStructValidations adds a list of struct-level validations for a type.
//
// If the struct validation for the type already exists,
// the previous validation list will be replaced.
func AddStructValidations[T any](v *Validator, validations ...StructValidation[T]) error {
	stInstance := new(T)
	stType := reflect.TypeOf(*stInstance)

	for _, sv := range validations {
		_, fieldFound := stType.FieldByName(sv.Field)
		if !fieldFound {
			return fmt.Errorf("add struct validation: field '%s' not found", sv.Field)
		}
	}
	v.validator.RegisterStructValidation(func(sl validator.StructLevel) {
		slReflection := sl.Current()
		st := slReflection.Interface().(T)

		for _, sv := range validations {
			if !sv.IsValid(&st) {
				fieldValue := slReflection.FieldByName(sv.Field).Interface()
				sl.ReportError(fieldValue, sv.Field, sv.Field, sv.Tag, sv.Field)
			}
		}
	}, stInstance)

	return nil
}

// GetTranslator accepts both en-US and en_US formats
func (v *Validator) GetTranslator(locale string) (Translator, bool) {
	locale = strings.ReplaceAll(locale, "-", "_")
	translator, found := v.universalTranslator.GetTranslator(locale)
	return translator, found
}

type TranslationHandler = func(v *validator.Validate, tr ut.Translator) error

// RegisterTranslations receives a handler that registers translations for the given translator.
func RegisterTranslations(v *Validator, handler TranslationHandler, locale string) error {
	translator, _ := v.GetTranslator(locale)
	return handler(v.validator, translator)
}

// MapErrors makes a validation.ErrorMap out of a validation.Error and translates
// all error messages and field names that have been given a custom translation.
func (v *Validator) MapErrors(err *Error, locales ...string) ErrorMap {
	for _, locale := range locales {
		translator, found := v.GetTranslator(locale)
		if found {
			return err.ToTranslatedErrorMap(translator)
		}
	}
	translator, _ := v.GetTranslator("")
	return err.ToTranslatedErrorMap(translator)
}
