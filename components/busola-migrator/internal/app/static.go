package app

import (
	"net/http"
	"path/filepath"
)

func (a App) HandleStaticWebsite(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(a.staticFilesDir, "index.html"))
}
