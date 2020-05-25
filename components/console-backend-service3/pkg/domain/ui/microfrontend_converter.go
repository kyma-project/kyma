package ui

import (
	uiv1alpha1 "github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/graph/model"
)

type MicroFrontendList []*uiv1alpha1.MicroFrontend
func(l *MicroFrontendList) Append() interface{} {
	converted := &uiv1alpha1.MicroFrontend{}
	*l = append(*l, converted)
	return converted
}

type MicroFrontendConverter struct {
	navigationNodeConverter navigationNodeConverter
}

func NewMicroFrontendConverter() MicroFrontendConverter {
	return MicroFrontendConverter{
		navigationNodeConverter: navigationNodeConverter{},
	}
}

func (c *MicroFrontendConverter) ToGQL(in *uiv1alpha1.MicroFrontend) (*model.MicroFrontend, error) {
	if in == nil {
		return nil, nil
	}

	navigationNodes, err := c.navigationNodeConverter.ToGQLs(in.Spec.CommonMicroFrontendSpec.NavigationNodes)
	if err != nil {
		return nil, err
	}
	mf := model.MicroFrontend{
		Name:            in.Name,
		Version:         in.Spec.CommonMicroFrontendSpec.Version,
		Category:        in.Spec.CommonMicroFrontendSpec.Category,
		ViewBaseURL:     in.Spec.CommonMicroFrontendSpec.ViewBaseURL,
		NavigationNodes: navigationNodes,
	}

	return &mf, nil
}

func (c *MicroFrontendConverter) ToGQLs(in []*uiv1alpha1.MicroFrontend) ([]*model.MicroFrontend, error) {
	var result []*model.MicroFrontend
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
