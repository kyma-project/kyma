package ui

import (
	uiV1alpha1v "github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

type microfrontendConverter struct{}

func (c *microfrontendConverter) ToGQL(in *uiV1alpha1v.MicroFrontend) *gqlschema.Microfrontend {
	if in == nil {
		return nil
	}

	navigationNodes := c.navigationNodesToGQLs(in.Spec.CommonMicroFrontendSpec.NavigationNodes)
	mf := gqlschema.Microfrontend{
		Name:            in.Name,
		Version:         in.Spec.CommonMicroFrontendSpec.Version,
		Category:        in.Spec.CommonMicroFrontendSpec.Category,
		ViewBaseURL:     in.Spec.CommonMicroFrontendSpec.ViewBaseURL,
		NavigationNodes: navigationNodes,
	}

	return &mf
}

func (c *microfrontendConverter) ToGQLs(in []*uiV1alpha1v.MicroFrontend) []gqlschema.Microfrontend {
	var result []gqlschema.Microfrontend
	for _, u := range in {
		converted := c.ToGQL(u)
		if converted != nil {
			result = append(result, *converted)
		}
	}
	return result
}

func (c *microfrontendConverter) navigationNodeToGQL(in *uiV1alpha1v.NavigationNode) *gqlschema.NavigationNode {
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

func (c *microfrontendConverter) navigationNodesToGQLs(in []uiV1alpha1v.NavigationNode) []gqlschema.NavigationNode {
	var result []gqlschema.NavigationNode
	for _, u := range in {
		converted := c.navigationNodeToGQL(&u)
		if converted != nil {
			result = append(result, *converted)
		}
	}
	return result
}
