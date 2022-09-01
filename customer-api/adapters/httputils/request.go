package httputils

import (
	"codepix/customer-api/adapters/modifier"
	"codepix/customer-api/lib/validation"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/mitchellh/mapstructure"
)

type bodykey struct{}
type paramskey struct{}

var bodyKey = bodykey{}
var paramsKey = paramskey{}

// Body gets the value in context bound by the validator.
func Body[T any](r *http.Request, destination T) T {
	value, ok := r.Context().Value(bodyKey).(T)
	if !ok {
		return *new(T)
	}
	return value
}

// Params gets the value in context bound by the validator.
func Params[T any](r *http.Request, destination T) T {
	value, ok := r.Context().Value(paramsKey).(T)
	if !ok {
		return *new(T)
	}
	return value
}

// WithBody validates the request body and binds it to the request's context.
func WithBody[T any](validator *validation.Validator, bindType T,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bind := new(T)

			err := json.NewDecoder(r.Body).Decode(bind)
			if err != nil {
				http.Error(w, "bad JSON", http.StatusBadRequest)
				return
			}

			err = moldAndValidate(validator, bind)
			if err != nil {
				if verr, ok := err.(*validation.Error); ok {
					ValidationError(w, r, validator, verr)
				} else {
					msg := fmt.Errorf("validate body: %w", err).Error()
					http.Error(w, msg, http.StatusInternalServerError)
				}
				return
			}

			ctxWithBind := context.WithValue(r.Context(), bodyKey, *bind)
			next.ServeHTTP(w, r.WithContext(ctxWithBind))
		})
	}
}

// WithParams validates the URL params and query params and binds them to the request's context.
// The fields in bindType must have a `param:"name"` tag for URL params used in httprouter
// and a `query:"name"` tag for query params.
func WithParams[T any](validator *validation.Validator, bindType T,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bind := new(T)

			err := bindURLParams(r, bind)
			if err != nil {
				msg := fmt.Errorf("validate params: %w", err).Error()
				http.Error(w, msg, http.StatusInternalServerError)
				return
			}
			err = bindQueryParams(r, bind)
			if err != nil {
				msg := fmt.Errorf("validate params: %w", err).Error()
				http.Error(w, msg, http.StatusInternalServerError)
				return
			}

			err = moldAndValidate(validator, bind)
			if err != nil {
				if verr, ok := err.(*validation.Error); ok {
					ValidationError(w, r, validator, verr)
				} else {
					msg := fmt.Errorf("validate params: %w", err).Error()
					http.Error(w, msg, http.StatusInternalServerError)
				}
				return
			}

			ctxWithBind := context.WithValue(r.Context(), paramsKey, *bind)
			next.ServeHTTP(w, r.WithContext(ctxWithBind))
		})
	}
}

func moldAndValidate(validator *validation.Validator, bind any) error {
	err := modifier.Mold(bind)
	if err != nil {
		return err
	}
	return validation.Validate(validator, bind)
}

func bindURLParams(r *http.Request, destination any) error {
	paramMap := map[string]string{}

	urlParams := httprouter.ParamsFromContext(r.Context())
	for _, param := range urlParams {
		paramMap[param.Key] = param.Value
	}
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:    "param",
		Result:     destination,
		DecodeHook: queryParamDecoder,
	})
	if err != nil {
		return err
	}
	err = decoder.Decode(paramMap)
	if err != nil {
		return err
	}
	return nil
}

func bindQueryParams(r *http.Request, destination any) error {
	paramMap := map[string]string{}

	queryParams := r.URL.Query()
	for k, v := range queryParams {
		paramMap[k] = v[0]
	}
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:          "query",
		Result:           destination,
		DecodeHook:       queryParamDecoder,
		WeaklyTypedInput: true,
	})
	if err != nil {
		return err
	}
	err = decoder.Decode(paramMap)
	if err != nil {
		return err
	}
	return nil
}

func queryParamDecoder(in reflect.Type, out reflect.Type, data interface{}) (interface{}, error) {
	switch out {
	case reflect.TypeOf(uuid.UUID{}):
		data, ok := data.(string)
		if !ok {
			return data, nil
		}
		return uuid.Parse(data)
	case reflect.TypeOf(time.Time{}):
		data, ok := data.(string)
		if !ok {
			return data, nil
		}
		return time.Parse(time.RFC3339, data)
	default:
		return data, nil
	}
}
