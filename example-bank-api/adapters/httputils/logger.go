package httputils

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-logr/logr"
)

const startFormat = "02 Jan 2006 15:04:05 -0700"

type requestLogger struct {
	http.ResponseWriter
	status int
	body   []byte
}

func (w *requestLogger) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
	}
	w.ResponseWriter.WriteHeader(status)
}
func (w *requestLogger) Write(b []byte) (int, error) {
	if !(w.status >= 200 && w.status < 400) {
		w.body = make([]byte, len(b))
		copy(w.body, b)
	}
	return w.ResponseWriter.Write(b)
}

func Logger(logger logr.Logger) func(http.Handler) http.Handler {
	logger = logger.WithName("http")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			writer := &requestLogger{ResponseWriter: w}

			start := time.Now()
			next.ServeHTTP(writer, r)
			duration := time.Since(start)

			kvs := []any{
				"method", r.Method,
				"path", r.URL.Path,
				"status", fmt.Sprintf("%d %s", writer.status, http.StatusText(writer.status)),
				"duration", duration.String(),
				"start", start.Format(startFormat),
				"ip", r.RemoteAddr,
				"forwarded", r.Header.Get("Forwarded"),
				"x-forwarded-for", r.Header.Get("X-Forwarded-For"),
				"user-agent", r.Header.Get("User-Agent"),
			}
			if writer.status >= 200 && writer.status < 400 {
				logger.Info("request handled", kvs...)
				return
			}
			response := string(writer.body)
			logger.Error(errors.New(response), "request failed", kvs...)
		})
	}
}
