package app

import (
	"net/http"
	"os"
	"path"
)

type App struct {
	busolaURL string
	fsRoot    http.FileSystem
}

func New(busolaURL, staticFilesDir string) App {
	wd, _ := os.Getwd()
	dir := path.Join(wd, "static")
	if staticFilesDir != "" {
		dir = staticFilesDir
	}

	return App{
		busolaURL: busolaURL,
		fsRoot:    http.Dir(dir),
	}
}
