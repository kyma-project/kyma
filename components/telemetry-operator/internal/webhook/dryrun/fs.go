package dryrun

import (
	"fmt"

	"github.com/spf13/afero"
)

type File struct {
	Path string
	Name string
	Data string
}

//go:generate mockery --name FileSystem --filename fs.go
type FileSystem interface {
	CreateAndWrite(s File) error
	RemoveDirectory(path string) error
}

type fileSystem struct {
	fs afero.Fs
}

func NewFileSystem() FileSystem {
	return &fileSystem{
		fs: afero.NewOsFs(),
	}
}

func (w *fileSystem) CreateAndWrite(s File) error {
	var err error

	if err = w.fs.MkdirAll(s.Path, 0755); err != nil {
		return err
	}

	var file afero.File
	if file, err = w.fs.Create(fmt.Sprintf("%s/%s", s.Path, s.Name)); err != nil {
		return err
	}

	if _, err = file.Write([]byte(s.Data)); err != nil {
		return err
	}

	return nil
}

func (w *fileSystem) RemoveDirectory(path string) error {
	return w.fs.RemoveAll(path)
}
