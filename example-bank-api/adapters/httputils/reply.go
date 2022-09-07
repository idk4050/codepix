package httputils

import (
	"codepix/example-bank-api/lib/repositories"
	"codepix/example-bank-api/lib/validation"
	"encoding/json"
	"errors"
	"net/http"

	"golang.org/x/text/language"
)

type Mapping map[any]int

func Error(w http.ResponseWriter, r *http.Request, err error, errorMappings ...Mapping) {
	if err == nil {
		return
	}
	for _, errorMapping := range errorMappings {
		for mappedError, mappedStatusCode := range errorMapping {
			_, ok := mappedError.(error)

			if ok && errors.As(err, &mappedError) {
				http.Error(w, mappedError.(error).Error(), mappedStatusCode)
				return
			}
		}
	}
	var notFoundError *repositories.NotFoundError
	var alreadyExistsError *repositories.AlreadyExistsError
	switch {
	case errors.As(err, &notFoundError):
		http.Error(w, notFoundError.Error(), http.StatusNotFound)
	case errors.As(err, &alreadyExistsError):
		http.Error(w, alreadyExistsError.Error(), http.StatusConflict)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func ValidationError(w http.ResponseWriter, r *http.Request, validator *validation.Validator,
	err *validation.Error) {
	acceptLanguages, _, _ := language.ParseAcceptLanguage(r.Header.Get("Accept-Language"))

	locales := []string{}
	for _, acceptLanguage := range acceptLanguages {
		locales = append(locales, acceptLanguage.String())
	}
	errorMap := validator.MapErrors(err, locales...)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(errorMap)
}

func Json(w http.ResponseWriter, body interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(body)
}
