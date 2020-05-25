package ui

import (
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/graph/model"
)

type BackendModuleConverter struct{}

type BackendModuleList []*v1alpha1.BackendModule
func(l *BackendModuleList) Append() interface{} {
	converted := &v1alpha1.BackendModule{}
	*l = append(*l, converted)
	return converted
}

func (c *BackendModuleConverter) ToGQL(in *v1alpha1.BackendModule) (*model.BackendModule, error) {
	if in == nil {
		return nil, nil
	}

	module := model.BackendModule{
		Name: in.Name,
	}

	return &module, nil
}

func (c *BackendModuleConverter) ToGQLs(in []*v1alpha1.BackendModule) ([]model.BackendModule, error) {
	var result []model.BackendModule
	for _, u := range in {
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
