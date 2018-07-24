package ybundle

import (
	"fmt"
	"io"
	"os"
)

// NewLocalRepository creates structure which allow us to access local repository
func NewLocalRepository(path string) *LocalRepository {
	return &LocalRepository{path: path}
}

// LocalRepository provide access to bundles repository
type LocalRepository struct {
	path string
}

// GetIndexFile returns index.yaml file from local repository
func (p *LocalRepository) GetIndexFile() (io.Reader, func(), error) {
	fName := fmt.Sprintf("%s/%s", p.path, "index.yaml")
	f, err := os.Open(fName)
	if err != nil {
		return nil, nil, err
	}
	return f, func() { f.Close() }, nil
}

// GetBundlePath returns path to the bundle from local repository
func (p *LocalRepository) GetBundlePath(name string) (string, error) {
	return fmt.Sprintf("%s/%s", p.path, name), nil
}
