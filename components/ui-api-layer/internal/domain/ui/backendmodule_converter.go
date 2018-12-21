package ui

import (
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/apis/ui/v1alpha1"
)

type backendModuleConverter struct{}

func (c *backendModuleConverter) ToGQL(in *v1alpha1.BackendModule) (*gqlschema.BackendModule, error) {
	if in == nil {
		return nil, nil
	}


	module := gqlschema.BackendModule{
		Name:                in.Name,
	}

	return &module, nil
}
