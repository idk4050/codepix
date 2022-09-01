package httputils

import "net/http"

func NotAllowed(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
}
