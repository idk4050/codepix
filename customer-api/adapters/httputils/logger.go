package httputils

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-logr/logr"
)

type requestLogger struct {
	http.ResponseWriter
	status int
	body   []byte
}

func (r *requestLogger) WriteHeader(status int) {
	if r.status == 0 {
		r.status = status
	}
	r.ResponseWriter.WriteHeader(status)
}
func (r *requestLogger) Write(b []byte) (int, error) {
	if !(r.status >= 200 && r.status < 400) {
		r.body = make([]byte, len(b))
		copy(r.body, b)
	}
	return r.ResponseWriter.Write(b)
}

func Logger(logger logr.Logger) func(http.Handler) http.Handler {
	log := logger.WithName("http")

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
				"start", start.Format("02 Jan 2006 15:04:05 -0700"),
				"duration", duration.String(),
				"user-agent", r.Header.Get("User-Agent"),
			}
			kvs = append(kvs, getIP(r)...)

			if writer.status >= 200 && writer.status < 400 {
				log.Info("request", kvs...)
				return
			}
			response := string(writer.body)
			log.Error(errors.New(response), "request failed", kvs...)
		})
	}
}

func PanicLogger(logger logr.Logger) func(w http.ResponseWriter, r *http.Request, rcv any) {
	log := logger.WithName("http.panic")

	return func(w http.ResponseWriter, r *http.Request, rcv any) {
		kvs := []any{
			"method", r.Method,
			"path", r.URL.Path,
			"user-agent", r.Header.Get("User-Agent"),
		}
		kvs = append(kvs, getIP(r)...)

		var err error = nil
		if rcvErr, ok := rcv.(error); ok {
			err = rcvErr
		} else {
			kvs = append(kvs, "recover", rcv)
		}
		log.Error(err, "request panic", kvs...)

		if writer, ok := w.(*requestLogger); ok {
			if writer.status == 0 {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
	}
}

func getIP(r *http.Request) []any {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		return []any{
			"ip", xff,
			"proxy", r.RemoteAddr,
		}
	} else {
		return []any{
			"ip", r.RemoteAddr,
		}
	}
}
