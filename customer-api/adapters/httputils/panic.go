package httputils

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-logr/logr"
)

func PanicHandler(logger logr.Logger) func(http.Handler) http.Handler {
	logger = logger.WithName("http.panic")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			panicked := true
			defer func() {
				if panicked {
					duration := time.Since(start)

					var err error
					switch r := recover().(type) {
					case error:
						err = fmt.Errorf("internal error: %w", r)
					default:
						err = fmt.Errorf("internal error: %v", r)
					}
					kvs := []any{
						"method", r.Method,
						"path", r.URL.Path,
						"duration", duration.String(),
						"start", start.Format(startFormat),
						"ip", r.RemoteAddr,
						"forwarded", r.Header.Get("Forwarded"),
						"x-forwarded-for", r.Header.Get("X-Forwarded-For"),
						"user-agent", r.Header.Get("User-Agent"),
					}
					logger.Error(err, "request panic", kvs...)
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
			panicked = false
		})
	}
}
