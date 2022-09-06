package validator

import (
	"codepix/example-bank-api/lib/validation"
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	en_US_locale "github.com/go-playground/locales/en_US"
	pt_BR_locale "github.com/go-playground/locales/pt_BR"
	ut "github.com/go-playground/universal-translator"
	en_US_translation "github.com/go-playground/validator/v10/translations/en"
	pt_BR_translation "github.com/go-playground/validator/v10/translations/pt_BR"
)

const EN_US = "en_US"
const PT_BR = "pt_BR"

func New() (*validation.Validator, error) {
	var en_US_translator = en_US_locale.New()
	var pt_BR_translator = pt_BR_locale.New()
	var fallback_translator = en_US_translator

	universalTranslator := ut.New(
		fallback_translator,
		en_US_translator,
		pt_BR_translator,
	)
	val := validation.New(universalTranslator)
	err := validation.RegisterTranslations(val, en_US_translation.RegisterDefaultTranslations, EN_US)
	if err != nil {
		return nil, err
	}
	err = validation.RegisterTranslations(val, pt_BR_translation.RegisterDefaultTranslations, PT_BR)
	if err != nil {
		return nil, err
	}
	return val, nil
}

// LoadTranslationFile loads up a JSON file containing translations.
//
// structTypes is a list of empty structs that correspond to the structs being translated.
// Any translated pointer fields in any structType must also be initialized.
//
// The JSON file should be built in the following format:
//
//	[struct_name] -> [locale] -> "error_messages" -> [tag]: [error message]
//	[struct_name] -> [locale] -> "field_names" -> [field namespace]: [field name translation]
//
// Where [field namespace] can contain fields that are
// deeply nested inside structs, arrays or maps.
//
// Field namespaces are independent of each other across different structs.
// Allowing both Struct1.Struct2.Field and Struct2.Field to have their own translations.
func LoadTranslationFile(validator *validation.Validator, file io.Reader, structTypes ...any,
) error {
	type categories = map[string]string
	type locales = map[string]categories
	type structNames = map[string]locales
	translationData := map[string]structNames{}

	err := json.NewDecoder(file).Decode(&translationData)
	if err != nil {
		return fmt.Errorf("translation file could not be parsed: %w", err)
	}

	structTypeMap := map[string]any{}

	for _, typ := range structTypes {
		structName := reflect.TypeOf(typ).Name()

		for fileStructName := range translationData {
			if structName == fileStructName {
				structTypeMap[structName] = typ
			}
		}
		if _, ok := structTypeMap[structName]; !ok {
			return fmt.Errorf("translation file does not contain struct type %s", structName)
		}
	}

	for structName, locales := range translationData {
		structType := structTypeMap[structName]

		for locale, categories := range locales {
			_, translatorFound := validator.GetTranslator(locale)
			if !translatorFound {
				return fmt.Errorf("translation file contains invalid locale %s.%s",
					structName, locale)
			}

			for category := range categories {
				switch category {
				case "error_messages":
					for tag, template := range categories[category] {
						err = validator.AddErrorMessageTranslation(locale, tag, template)

						if err != nil {
							return fmt.Errorf("could not add error message translation for %s.%s.%s.%s: %w",
								structName, locale, category, tag, err)
						}
					}
				case "field_names":
					for namespace, translation := range categories[category] {
						err = validator.AddFieldNameTranslation(locale, structType, namespace, translation)

						if err != nil {
							return fmt.Errorf("could not add field name translation for %s.%s.%s.%s: %w",
								structName, locale, category, namespace, err)
						}
					}
				default:
					return fmt.Errorf("translation file contains invalid category %s.%s.%s",
						structName, locale, category)
				}
			}
		}
	}
	return nil
}
