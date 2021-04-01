package app

import (
	"os"
	"path"
)

type App struct {
	busolaURL      string
	staticFilesDir string
}

func New(busolaURL, staticFilesDir string) App {
	wd, _ := os.Getwd()
	dir := path.Join(wd, "static")
	if staticFilesDir != "" {
		dir = staticFilesDir
	}

	return App{
		busolaURL:      busolaURL,
		staticFilesDir: dir,
	}
}
