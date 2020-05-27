package ui

import (
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/graph/model"
)

type BackendModuleConverter struct{}

type BackendModuleList []*v1alpha1.BackendModule
func(l *BackendModuleList) Append() interface{} {
	converted := &v1alpha1.BackendModule{}
	*l = append(*l, converted)
	return converted
}

func (c BackendModuleConverter) ToGQL(in *v1alpha1.BackendModule) *model.BackendModule {
	module := model.BackendModule{
		Name: in.Name,
	}

	return &module
}

func (c BackendModuleConverter) ToGQLs(in []*v1alpha1.BackendModule) []*model.BackendModule {
	var result []*model.BackendModule
	for _, u := range in {
		converted := c.ToGQL(u)

		if converted != nil {
			result = append(result, converted)
		}
	}
	return result
}
