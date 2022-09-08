package rpc

import (
	"codepix/example-bank-api/adapters/httputils"
	"net/http"

	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

func ErrorToHTTP(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}
	if s, ok := grpcstatus.FromError(err); ok {
		StatusToHTTP(w, s.Proto())
		return
	}
	httputils.Error(w, r, err)
}

func StatusToHTTP(w http.ResponseWriter, status *status.Status) {
	if status == nil {
		return
	}
	code := ToHTTPCode(codes.Code(status.Code))
	if len(status.Details) > 0 {
		httputils.Json(w, status.Details[0], code)
		return
	}
	http.Error(w, status.Message, code)
}

func ToHTTPCode(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.Canceled:
		return http.StatusRequestTimeout
	case codes.Unknown:
		return http.StatusInternalServerError
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.FailedPrecondition:
		return http.StatusBadRequest
	case codes.Aborted:
		return http.StatusConflict
	case codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Internal:
		return http.StatusInternalServerError
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DataLoss:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
