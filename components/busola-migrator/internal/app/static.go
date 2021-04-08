package app

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

func (a App) HandleStaticWebsite(w http.ResponseWriter, r *http.Request) {
	rctx := chi.RouteContext(r.Context())
	pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
	fs := http.StripPrefix(pathPrefix, http.FileServer(a.fsRoot))
	fs.ServeHTTP(w, r)
}
