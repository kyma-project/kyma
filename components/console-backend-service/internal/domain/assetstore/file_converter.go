package assetstore

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

type fileConverter struct{}

func (c *fileConverter) ToGQL(file *File) (*gqlschema.File, error) {
	if file == nil {
		return nil, nil
	}

	metadata, err := c.extractMetadata(file.Metadata)
	if err != nil {
		return nil, errors.Wrapf(err, "while unmarshalling Metadata field of File %s", file.URL)
	}

	result := gqlschema.File{
		URL:      file.URL,
		Metadata: metadata,
	}
	return &result, nil
}

func (c *fileConverter) ToGQLs(files []*File) ([]gqlschema.File, error) {
	var result []gqlschema.File
	for _, u := range files {
		converted, err := c.ToGQL(u)
		if err != nil {
			return nil, err
		}

		if converted != nil {
			result = append(result, *converted)
		}
	}
	return result, nil
}

func (c *fileConverter) extractMetadata(metadata *runtime.RawExtension) (gqlschema.JSON, error) {
	if metadata == nil {
		return nil, nil
	}

	extracted, err := resource.ExtractRawToMap("Metadata", metadata.Raw)
	if err != nil {
		return nil, err
	}

	result := make(gqlschema.JSON)
	for k, v := range extracted {
		result[k] = v
	}

	return result, err
}
