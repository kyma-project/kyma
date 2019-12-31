package ui

import (
	uiV1alpha1v "github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

//go:generate mockery -name=qglNavigationNodeConverter -output=automock -outpkg=automock -case=underscore
type qglNavigationNodeConverter interface {
	ToGQL(in *uiV1alpha1v.NavigationNode) (*gqlschema.NavigationNode, error)
	ToGQLs(in []uiV1alpha1v.NavigationNode) ([]gqlschema.NavigationNode, error)
}

type clusterMicroFrontendConverter struct {
	navigationNodeConverter qglNavigationNodeConverter
}

func newClusterMicroFrontendConverter() *clusterMicroFrontendConverter {
	return &clusterMicroFrontendConverter{
		navigationNodeConverter: &navigationNodeConverter{},
	}
}

func (c *clusterMicroFrontendConverter) ToGQL(in *uiV1alpha1v.ClusterMicroFrontend) (*gqlschema.ClusterMicroFrontend, error) {
	if in == nil {
		return nil, nil
	}

	navigationNodes, err := c.navigationNodeConverter.ToGQLs(in.Spec.CommonMicroFrontendSpec.NavigationNodes)
	if err != nil {
		return nil, err
	}

	cmf := gqlschema.ClusterMicroFrontend{
		Name:            in.Name,
		Placement:       in.Spec.Placement,
		Version:         in.Spec.CommonMicroFrontendSpec.Version,
		Category:        in.Spec.CommonMicroFrontendSpec.Category,
		PreloadURL:      in.Spec.CommonMicroFrontendSpec.PreloadURL,
		ViewBaseURL:     in.Spec.CommonMicroFrontendSpec.ViewBaseURL,
		Experimental:    in.Spec.CommonMicroFrontendSpec.Experimental,
		NavigationNodes: navigationNodes,
	}

	return &cmf, nil
}

func (c *clusterMicroFrontendConverter) ToGQLs(in []*uiV1alpha1v.ClusterMicroFrontend) ([]gqlschema.ClusterMicroFrontend, error) {
	var result []gqlschema.ClusterMicroFrontend
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
