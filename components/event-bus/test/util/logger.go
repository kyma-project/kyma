package util

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
)

// Logger ...
func Logger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s requested %s", r.RemoteAddr, r.URL)
		dump, err := httputil.DumpRequest(r, true)
		log.Printf("%q", dump)
		if err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}
		h.ServeHTTP(w, r)
	})
}
