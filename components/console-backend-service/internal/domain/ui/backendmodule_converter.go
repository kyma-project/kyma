package ui

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/apis/ui/v1alpha1"
)

type backendModuleConverter struct{}

func (c *backendModuleConverter) ToGQL(in *v1alpha1.BackendModule) (*gqlschema.BackendModule, error) {
	if in == nil {
		return nil, nil
	}

	module := gqlschema.BackendModule{
		Name: in.Name,
	}

	return &module, nil
}

func (c *backendModuleConverter) ToGQLs(in []*v1alpha1.BackendModule) ([]*gqlschema.BackendModule, error) {
	var result []*gqlschema.BackendModule
	for _, u := range in {
		converted, err := c.ToGQL(u)
		if err != nil {
			return nil, err
		}

		if converted != nil {
			result = append(result, converted)
		}
	}
	return result, nil
}
