package fs

import (
	"fmt"

	"github.com/spf13/afero"
)

type File struct {
	Path string
	Name string
	Data string
}

//go:generate mockery --name Wrapper --filename wrapper.go
type Wrapper interface {
	CreateAndWrite(s File) error
	RemoveDirectory(path string) error
}

type wrapper struct {
	fs afero.Fs
}

func NewWrapper() Wrapper {
	return &wrapper{
		fs: afero.NewOsFs(),
	}
}

func (w *wrapper) CreateAndWrite(s File) error {
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

func (w *wrapper) RemoveDirectory(path string) error {
	return w.fs.RemoveAll(path)
}
