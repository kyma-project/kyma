package matador

import (
	"github.com/gernest/front"
	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/asset-metadata-service/pkg/fileheader"
	"github.com/pkg/errors"
)

//go:generate mockery -name=Matador -output=automock -outpkg=automock -case=underscore
type Matador interface {
	ReadMetadata(fileHeader fileheader.FileHeader) (map[string]interface{}, error)
}

type matador struct {
	frontMatter *front.Matter
}

func New() Matador {
	f := front.NewMatter()
	f.Handle("---", front.YAMLHandler)

	return &matador{
		frontMatter: f,
	}
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

	metadata, _, err := m.frontMatter.Parse(f)
	if err != nil {
		return nil, errors.Wrapf(err, "while reading metadata from file %s", fileHeader.Filename())
	}

	return metadata, nil
}

