package fileutils

import (
	"fmt"
	"github.com/spf13/afero"
)

func Write(path string, name string, data []byte) error {
	appfs := afero.NewOsFs()
	err := appfs.MkdirAll(path, 0755)
	if err != nil {
		return err
	}
	file, err := appfs.Create(fmt.Sprintf("%s/%s", path, name))

	if err != nil {
		return err
	}

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func Delete(path string) error {
	return nil
}
