package main

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/urfave/negroni"
)

var contentTypes = map[string]string{
	".tgz":  "application/gzip",
	".yaml": "text/plain", // text/plain allows to see the content in the browser.
}

func filteringMiddleware(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if r.URL.Path != "/index.yaml" && filepath.Ext(r.URL.Path) != ".tgz" {
		rw.WriteHeader(http.StatusNotFound)
		return
	}
	next(rw, r)
}

func contentTypeMiddleware(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	ext := filepath.Ext(r.URL.Path)
	if contentType, ok := contentTypes[ext]; ok {
		rw.Header().Set("Content-Type", contentType)
	}
	next(rw, r)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		panic("PORT env must be set")
	}
	contentPath := os.Getenv("CONTENT_PATH")
	if contentPath == "" {
		panic("CONTENT_PATH env must be set")
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(contentPath)))

	n := negroni.New(negroni.NewLogger())
	n.UseFunc(filteringMiddleware)
	n.UseFunc(contentTypeMiddleware)
	n.UseHandler(mux)

	err := http.ListenAndServe(":"+port, n)
	if err != nil {
		panic(err)
	}
}
