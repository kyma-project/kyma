package content

import (
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/storage"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
)

type contentConverter struct{}

func (c *contentConverter) ToGQL(in *storage.Content) *gqlschema.JSON {
	if in == nil {
		return nil
	}

	result := make(gqlschema.JSON)
	for k, v := range in.Raw {
		result[k] = v
	}

	return &result
}
