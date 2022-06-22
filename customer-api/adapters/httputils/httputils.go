package httputils

import "net/http"

type RouterHandler func(method string, path string, handler http.Handler)
