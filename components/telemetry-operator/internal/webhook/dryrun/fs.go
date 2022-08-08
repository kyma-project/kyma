package dryrun

import (
	"fmt"
	"os"
)

type fileHandler struct {
	path string
	name string
	data string
}

type fileSystem struct {
}

func (fs *fileSystem) CreateAndWrite(f fileHandler) error {
	if err := os.MkdirAll(f.path, 0755); err != nil {
		return err
	}

	file, err := os.Create(fmt.Sprintf("%s/%s", f.path, f.name))
	if err != nil {
		return err
	}

	if _, err := file.Write([]byte(f.data)); err != nil {
		return err
	}

	return nil
}

func (fs *fileSystem) RemoveDirectory(path string) error {
	return os.RemoveAll(path)
}
