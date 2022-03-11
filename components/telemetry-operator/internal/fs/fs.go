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

func CreateAndWrite(s File) error {
	var err error

	appfs := afero.NewOsFs()
	if err = appfs.MkdirAll(s.Path, 0755); err != nil {
		return err
	}

	var file afero.File
	if file, err = appfs.Create(fmt.Sprintf("%s/%s", s.Path, s.Name)); err != nil {
		return err
	}

	if _, err = file.Write([]byte(s.Data)); err != nil {
		return err
	}

	return nil
}

func RemoveDirectory(path string) error {
	appfs := afero.NewOsFs()

	return appfs.RemoveAll(path)
}
