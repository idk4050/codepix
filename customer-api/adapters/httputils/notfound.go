package httputils

import "net/http"

func NotFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}
