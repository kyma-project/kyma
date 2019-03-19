package matador

import (
	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/asset-metadata-service/pkg/fileheader"
	"github.com/ericaro/frontmatter"
	"github.com/pkg/errors"
	"io/ioutil"
)

//go:generate mockery -name=Matador -output=automock -outpkg=automock -case=underscore
type Matador interface {
	ReadMetadata(fileHeader fileheader.FileHeader) (map[string]interface{}, error)
}

type matador struct {}

func New() Matador {
	return &matador{}
}

func (m *matador) ReadMetadata(fileHeader fileheader.FileHeader) (map[string]interface{}, error) {
	f, err := fileHeader.Open()
	if err != nil {
		return nil, errors.Wrapf(err, "while opening file %s", fileHeader.Filename())
	}
	defer func() {
		err := f.Close()
		if err != nil {
			glog.Error(err)
		}
	}()

	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.Wrapf(err, "while reading file %s", fileHeader.Filename())
	}


	var metadata map[string]interface{}
	err = frontmatter.Unmarshal(bytes, &metadata)
	if err != nil {
		return nil, errors.Wrapf(err, "while reading metadata from file %s", fileHeader.Filename())
	}

	return metadata, nil
}

