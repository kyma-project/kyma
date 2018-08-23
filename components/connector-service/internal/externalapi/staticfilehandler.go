package externalapi

import (
	"net/http"
)

// NewStaticFileHandler creates handler for returning API spec
func NewStaticFileHandler(path string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path)
	})
}
