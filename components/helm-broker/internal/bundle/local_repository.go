package bundle

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

// IndexReader returns index.yaml file from local repository
func (p *LocalRepository) IndexReader() (io.Reader, func(), error) {
	fName := fmt.Sprintf("%s/%s", p.path, "index.yaml")
	f, err := os.Open(fName)
	if err != nil {
		return nil, nil, err
	}
	return f, func() { f.Close() }, nil
}

// BundleReader calls repository for a specific bundle and returns means to read bundle content.
func (p *LocalRepository) BundleReader(name Name, version Version) (r io.Reader, closer func(), err error) {
	f, err := os.Open(fmt.Sprintf("%s/%s-%s.tgz", p.path, name, version))
	if err != nil {
		return nil, nil, err
	}
	return f, func() { f.Close() }, nil
}
