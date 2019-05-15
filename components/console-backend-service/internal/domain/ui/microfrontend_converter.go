package ui

import (
	uiV1alpha1v "github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

type microfrontendConverter struct {
	navigationNodeConverter qglNavigationNodeConverter
}

func newMicrofrontendConverter() *microfrontendConverter {
	return &microfrontendConverter{
		navigationNodeConverter: &navigationNodeConverter{},
	}
}

func (c *microfrontendConverter) ToGQL(in *uiV1alpha1v.MicroFrontend) (*gqlschema.Microfrontend, error) {
	if in == nil {
		return nil, nil
	}

	navigationNodes, err := c.navigationNodeConverter.ToGQLs(in.Spec.CommonMicroFrontendSpec.NavigationNodes)
	if err != nil {
		return nil, err
	}
	mf := gqlschema.Microfrontend{
		Name:            in.Name,
		Version:         in.Spec.CommonMicroFrontendSpec.Version,
		Category:        in.Spec.CommonMicroFrontendSpec.Category,
		ViewBaseURL:     in.Spec.CommonMicroFrontendSpec.ViewBaseURL,
		NavigationNodes: navigationNodes,
	}

	return &mf, nil
}

func (c *microfrontendConverter) ToGQLs(in []*uiV1alpha1v.MicroFrontend) ([]gqlschema.Microfrontend, error) {
	var result []gqlschema.Microfrontend
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
