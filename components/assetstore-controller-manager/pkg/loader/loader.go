package loader

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha1"
)

type loader struct {
	temporaryDir string

	// for testing
	osRemoveAllFunc func(string) error
	osCreateFunc    func(name string) (*os.File, error)
	httpGetFunc     func(url string) (*http.Response, error)
	ioutilTempDir   func(dir, prefix string) (string, error)
}

//go:generate mockery -name=Loader -output=automock -outpkg=automock -case=underscore
type Loader interface {
	Load(src, assetName string, mode v1alpha1.AssetMode, filter string) (string, []string, error)
	Clean(path string) error
}

func New(temporaryDir string) Loader {
	if len(temporaryDir) == 0 {
		temporaryDir = os.TempDir()
	}

	return &loader{
		temporaryDir:    temporaryDir,
		osRemoveAllFunc: os.RemoveAll,
		osCreateFunc:    os.Create,
		httpGetFunc:     http.Get,
		ioutilTempDir:   ioutil.TempDir,
	}
}

func (l *loader) Load(src, assetName string, mode v1alpha1.AssetMode, filter string) (string, []string, error) {
	switch mode {
	case v1alpha1.AssetSingle:
		return l.loadSingle(src, assetName)
	case v1alpha1.AssetPackage:
		return l.loadPackage(src, assetName, filter)
	}

	return "", nil, fmt.Errorf("not supported source mode %+v", mode)
}

func (l *loader) Clean(path string) error {
	return l.osRemoveAllFunc(path)
}

func (l *loader) download(destination, source string) error {
	file, err := l.osCreateFunc(destination)
	if err != nil {
		return err
	}
	defer file.Close()

	response, err := l.httpGetFunc(source)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}

func (l *loader) fileName(source string) string {
	_, filename := path.Split(source)
	if len(filename) == 0 {
		filename = "asset"
	}

	return filename
}
