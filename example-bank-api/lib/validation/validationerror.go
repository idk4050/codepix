package validation

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/iancoleman/strcase"
)

type Error struct {
	FieldErrors []FieldError
}

type FieldError = validator.FieldError

// Error returns a short message containing only the first field error.
func (e Error) Error() string {
	fe := e.FieldErrors[0]
	message := fieldErrorMessage(fe)
	message = strings.ReplaceAll(message, fe.Field(), snakeCasedNamespace(fe))
	return message
}

func fieldErrorMessage(fe FieldError) string {
	return fmt.Sprintf("validation failed on %s (%s)", fe.Field(), fe.Tag())
}

// ErrorMap has the following format:
//
//	[snake_cased_field_namespace]: "error message that can include snake_cased_field_namespace"
type ErrorMap = map[string]string

// ToErrorMap makes an ErrorMap out of a validation.Error.
func (e *Error) ToErrorMap() *ErrorMap {
	errMap := ErrorMap{}

	for _, fe := range e.FieldErrors {
		key := snakeCasedNamespace(fe)
		message := fieldErrorMessage(fe)
		message = strings.ReplaceAll(message, fe.Field(), key)
		errMap[key] = message
	}
	return &errMap
}

// ToTranslatedErrorMap makes an ErrorMap out of a validation.Error and translates
// all error messages and field names that have been given a custom translation.
func (e *Error) ToTranslatedErrorMap(tr Translator) ErrorMap {
	errorMap := ErrorMap{}
	for _, fe := range e.FieldErrors {
		message := fe.Translate(tr)

		translationFound := message != fe.Error()
		if !translationFound {
			message = fieldErrorMessage(fe)
		}

		translatedField, ok := findFieldNameTranslation(fe, tr)
		if ok {
			message = strings.ReplaceAll(message, fe.Field(), translatedField)
		} else {
			message = strings.ReplaceAll(message, fe.Field(), snakeCasedNamespace(fe))
		}
		namespace := snakeCasedNamespace(fe)
		errorMap[namespace] = message
	}
	return errorMap
}

func snakeCasedNamespace(fe FieldError) string {
	ns := removeTopLevelInNamespace(fe.Namespace())
	ns = mapNamespaceFields(ns, unwrapFieldName)
	ns = mapNamespaceFields(ns, fieldToSnakeCase)
	ns = mapNamespaceFields(ns, removeFieldTypeIfPresent)
	return ns
}

const fieldNamePrefix = "{{"
const fieldNameSuffix = "}}"

// wrapFieldName wraps a field name in delimiters.
func wrapFieldName(fieldName string) string {
	return fieldNamePrefix + fieldName + fieldNameSuffix
}

// unwrapFieldName removes the delimiters from a field name.
func unwrapFieldName(fieldName string) string {
	parts := strings.SplitN(fieldName, "[", 2)
	wrappedPart := parts[0]

	unwrappedField := strings.TrimSuffix(
		strings.TrimPrefix(wrappedPart, fieldNamePrefix), fieldNameSuffix)

	if len(parts) == 2 {
		bracketsPart := parts[1]
		return unwrappedField + "[" + bracketsPart
	}
	return unwrappedField
}

// removeTopLevelInNamespace removes the top level struct name in a namespace.
func removeTopLevelInNamespace(namespace string) string {
	return strings.SplitN(namespace, ".", 2)[1]
}

// mapNamespaceFields applies the given function to every namespace field.
func mapNamespaceFields(namespace string, mapper func(field string) string) string {
	namespaceParts := strings.Split(namespace, ".")
	newNamespace := bytes.NewBufferString("")

	for i, field := range namespaceParts {
		newNamespace.WriteString(mapper(field))

		if i < len(namespaceParts)-1 {
			newNamespace.WriteString(".")
		}
	}
	return newNamespace.String()
}

// fieldToSnakeCase converts a field name to snake case.
func fieldToSnakeCase(fieldName string) string {
	parts := strings.SplitN(fieldName, "[", 2)
	snakeCaseField := strcase.ToSnake(parts[0])

	hasBracketsPart := len(parts) == 2
	if hasBracketsPart {
		return snakeCaseField + "[" + parts[1]
	}
	return snakeCaseField
}
