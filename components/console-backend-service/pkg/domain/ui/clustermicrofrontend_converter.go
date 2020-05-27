package ui

import (
	uiv1alpha1 "github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/graph/model"
)


type ClusterMicroFrontendList []*uiv1alpha1.ClusterMicroFrontend
func(l *ClusterMicroFrontendList) Append() interface{} {
	converted := &uiv1alpha1.ClusterMicroFrontend{}
	*l = append(*l, converted)
	return converted
}

type clusterMicroFrontendConverter struct {
	navigationNodeConverter navigationNodeConverter
}

func newClusterMicroFrontendConverter() *clusterMicroFrontendConverter {
	return &clusterMicroFrontendConverter{
		navigationNodeConverter: navigationNodeConverter{},
	}
}

func (c *clusterMicroFrontendConverter) ToGQL(in *uiv1alpha1.ClusterMicroFrontend) (*model.ClusterMicroFrontend, error) {
	if in == nil {
		return nil, nil
	}

	navigationNodes, err := c.navigationNodeConverter.ToGQLs(in.Spec.CommonMicroFrontendSpec.NavigationNodes)
	if err != nil {
		return nil, err
	}

	cmf := model.ClusterMicroFrontend{
		Name:            in.Name,
		Placement:       in.Spec.Placement,
		Version:         in.Spec.CommonMicroFrontendSpec.Version,
		Category:        in.Spec.CommonMicroFrontendSpec.Category,
		PreloadURL:      in.Spec.CommonMicroFrontendSpec.PreloadURL,
		ViewBaseURL:     in.Spec.CommonMicroFrontendSpec.ViewBaseURL,
		NavigationNodes: navigationNodes,
	}

	return &cmf, nil
}

func (c *clusterMicroFrontendConverter) ToGQLs(in []*uiv1alpha1.ClusterMicroFrontend) ([]*model.ClusterMicroFrontend, error) {
	var result []*model.ClusterMicroFrontend
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
