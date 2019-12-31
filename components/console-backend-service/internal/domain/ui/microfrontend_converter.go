package ui

import (
	uiV1alpha1v "github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

type microFrontendConverter struct {
	navigationNodeConverter qglNavigationNodeConverter
}

func newMicroFrontendConverter() *microFrontendConverter {
	return &microFrontendConverter{
		navigationNodeConverter: &navigationNodeConverter{},
	}
}

func (c *microFrontendConverter) ToGQL(in *uiV1alpha1v.MicroFrontend) (*gqlschema.MicroFrontend, error) {
	if in == nil {
		return nil, nil
	}

	navigationNodes, err := c.navigationNodeConverter.ToGQLs(in.Spec.CommonMicroFrontendSpec.NavigationNodes)
	if err != nil {
		return nil, err
	}
	mf := gqlschema.MicroFrontend{
		Name:            in.Name,
		Version:         in.Spec.CommonMicroFrontendSpec.Version,
		Category:        in.Spec.CommonMicroFrontendSpec.Category,
		ViewBaseURL:     in.Spec.CommonMicroFrontendSpec.ViewBaseURL,
		Experimental:    in.Spec.CommonMicroFrontendSpec.Experimental,
		NavigationNodes: navigationNodes,
	}

	return &mf, nil
}

func (c *microFrontendConverter) ToGQLs(in []*uiV1alpha1v.MicroFrontend) ([]gqlschema.MicroFrontend, error) {
	var result []gqlschema.MicroFrontend
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
