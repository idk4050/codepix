package validation

// This translation setup is meant to be used only for translating validation error messages.

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	ut "github.com/go-playground/universal-translator"
	"github.com/mcuadros/go-lookup"
)

type Translator = ut.Translator

// AddErrorMessageTranslation adds a field-level error message translation
// for the given tag.
//
// The tag is not partitioned by struct type and can only be registered once
// for the entire locale.
//
// The template can contain a {0} placeholder for the field name.
func (v *Validator) AddErrorMessageTranslation(locale string, tag string, template string,
) error {
	translator, found := v.universalTranslator.GetTranslator(locale)
	if !found {
		return fmt.Errorf("translator with locale %s not found", locale)
	}

	key := tag

	return v.validator.RegisterTranslation(tag, translator, func(tr Translator) error {
		return tr.Add(key, template, true)
	}, func(tr Translator, fe FieldError) string {
		translation, _ := tr.T(key, fe.Field())
		return translation
	})
}

// This is a hack used to add field name translations to the existing key space
// without having to setup new translators just for the field names.
const fieldNameKeyPrefix = "field-name-"

// AddFieldNameTranslation adds a translation for the given field's namespace.
//
// The structType is an empty struct that corresponds to the struct being translated.
// Any pointer fields in structType that are going to be used must also be initialized.
//
// The fieldNamespace must be a string containing nested field names from top struct
// to field, separated by "."
func (v *Validator) AddFieldNameTranslation(locale string, structType interface{},
	fieldNamespace string, translation string) error {

	translator, found := v.universalTranslator.GetTranslator(locale)
	if !found {
		return fmt.Errorf("translator with locale %s not found", locale)
	}

	structTypeName := reflect.TypeOf(structType).Name()

	_, err := lookup.LookupString(structType, fieldNamespace)
	if err != nil {
		return fmt.Errorf("field name translation for struct type %s "+
			"contains invalid namespace: %s", structTypeName, fieldNamespace,
		)
	}

	key := fmt.Sprintf("%s%s.%s", fieldNameKeyPrefix, structTypeName, fieldNamespace)

	return v.validator.RegisterTranslation(key, translator, func(tr Translator) error {
		return tr.Add(key, translation, true)
	}, func(tr Translator, fe FieldError) string {
		translation, _ := tr.T(key, "")
		return translation
	})
}

// findFieldNameTranslation searches each step of the error's namespace for a translation
func findFieldNameTranslation(fe FieldError, tr Translator) (string, bool) {
	namespace := namespaceToTranslationKey(fe.Namespace())
	ns := namespace

	for {
		step := mapNamespaceFields(ns, removeFieldTypeIfPresent)
		translation := findFieldNameTranslationStep(step, tr)
		if translation != "" {
			return translation, true
		}

		step = mapNamespaceFields(ns, keepFieldTypeIfPresent)
		translation = findFieldNameTranslationStep(step, tr)
		if translation != "" {
			return translation, true
		}

		nsHeadAndTail := strings.SplitN(ns, ".", 2)
		if len(nsHeadAndTail) == 1 {
			break
		}
		ns = nsHeadAndTail[1]
	}
	return "", false
}

func findFieldNameTranslationStep(namespace string, tr Translator) string {
	for {
		translation, notFoundErr := tr.T(fieldNameKeyPrefix+namespace, "")
		if notFoundErr == nil {
			return translation
		}

		nsHeadAndTail := strings.SplitN(namespace, ".", 2)
		if len(nsHeadAndTail) == 1 {
			break
		}
		namespace = nsHeadAndTail[1]
	}
	return ""
}

var noArrayIndexes = regexp.MustCompile(`\[\d+\]`) // [unsigned integer]
var noMapKeys = regexp.MustCompile(`\[.*?\]*\]`)   // [anything or empty]

// namespaceToTranslationKey returns a key in a format that
// can be more easily defined by configuration files.
func namespaceToTranslationKey(namespace string) string {
	ns := mapNamespaceFields(namespace, unwrapFieldName)
	ns = noArrayIndexes.ReplaceAllLiteralString(ns, "")
	ns = noMapKeys.ReplaceAllLiteralString(ns, "")
	return ns
}

// removeFieldTypeIfPresent removes the field's type if it has one.
func removeFieldTypeIfPresent(field string) string {
	return strings.SplitN(field, "/", 2)[0]
}

// keepFieldTypeIfPresent returns the field's type if it has one or the name if it doesn't.
func keepFieldTypeIfPresent(field string) string {
	parts := strings.SplitN(field, "/", 2)
	return parts[len(parts)-1]
}
