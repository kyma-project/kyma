package ui

import (
	uiV1alpha1v "github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

type clusterMicrofrontendConverter struct{}

func (c *clusterMicrofrontendConverter) ToGQL(in *uiV1alpha1v.ClusterMicroFrontend) *gqlschema.ClusterMicrofrontend {
	if in == nil {
		return nil
	}

	navigationNodes := c.NavigationNodesToGQLs(in.Spec.CommonMicroFrontendSpec.NavigationNodes)

	cmf := gqlschema.ClusterMicrofrontend{
		Name:            in.Name,
		Placement:       in.Spec.Placement,
		Version:         in.Spec.CommonMicroFrontendSpec.Version,
		Category:        in.Spec.CommonMicroFrontendSpec.Category,
		ViewBaseURL:     in.Spec.CommonMicroFrontendSpec.ViewBaseURL,
		NavigationNodes: navigationNodes,
	}

	return &cmf
}

func (c *clusterMicrofrontendConverter) ToGQLs(in []*uiV1alpha1v.ClusterMicroFrontend) []gqlschema.ClusterMicrofrontend {
	var result []gqlschema.ClusterMicrofrontend
	for _, u := range in {
		converted := c.ToGQL(u)
		if converted != nil {
			result = append(result, *converted)
		}
	}
	return result
}

func (c *clusterMicrofrontendConverter) NavigationNodeToGQL(in *uiV1alpha1v.NavigationNode) *gqlschema.NavigationNode {
	if in == nil {
		return nil
	}

	navigationNode := gqlschema.NavigationNode{
		Label:            in.Label,
		NavigationPath:   in.NavigationPath,
		ViewURL:          in.ViewURL,
		ShowInNavigation: in.ShowInNavigation,
		Order:            in.Order,
		Settings: gqlschema.Settings{
			ReadOnly: in.Settings.ReadOnly,
		},
	}

	return &navigationNode
}

func (c *clusterMicrofrontendConverter) NavigationNodesToGQLs(in []uiV1alpha1v.NavigationNode) []gqlschema.NavigationNode {
	var result []gqlschema.NavigationNode
	for _, u := range in {
		converted := c.NavigationNodeToGQL(&u)
		if converted != nil {
			result = append(result, *converted)
		}
	}
	return result
}
