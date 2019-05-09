package ui

import (
	"encoding/json"

	uiV1alpha1v "github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/ui/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
)

type microfrontendConverter struct{}

func (c *microfrontendConverter) ToGQL(in *uiV1alpha1v.MicroFrontend) (*gqlschema.Microfrontend, error) {
	if in == nil {
		return nil, nil
	}

	navigationNodes, err := c.navigationNodesToGQLs(in.Spec.CommonMicroFrontendSpec.NavigationNodes)
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

func (c *microfrontendConverter) navigationNodeToGQL(in *uiV1alpha1v.NavigationNode) (*gqlschema.NavigationNode, error) {
	if in == nil {
		return nil, nil
	}

	settingsGqlJSON, err := c.settingsToGQLJSON(in)
	if err != nil {
		return nil, err
	}

	navigationNode := gqlschema.NavigationNode{
		Label:            in.Label,
		NavigationPath:   in.NavigationPath,
		ViewURL:          in.ViewURL,
		ShowInNavigation: in.ShowInNavigation,
		Order:            in.Order,
		Settings:         settingsGqlJSON,
	}

	return &navigationNode, nil
}

func (c *microfrontendConverter) navigationNodesToGQLs(in []uiV1alpha1v.NavigationNode) ([]gqlschema.NavigationNode, error) {
	var result []gqlschema.NavigationNode
	for _, u := range in {
		converted, err := c.navigationNodeToGQL(&u)
		if err != nil {
			return nil, err
		}
		if converted != nil {
			result = append(result, *converted)
		}
	}
	return result, nil
}

func (c *microfrontendConverter) settingsToGQLJSON(in *uiV1alpha1v.NavigationNode) (gqlschema.Settings, error) {
	if in == nil {
		return nil, nil
	}

	jsonByte, err := json.Marshal(in.Settings)
	if err != nil {
		return nil, errors.Wrapf(err, "while marshalling %s with ViewURL `%s`", pretty.NavigationNode, in.ViewURL)
	}

	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonByte, &jsonMap)
	if err != nil {
		return nil, errors.Wrapf(err, "while unmarshalling %s with ViewURL `%s` to map", pretty.NavigationNode, in.ViewURL)
	}

	var result gqlschema.Settings
	err = result.UnmarshalGQL(jsonMap)
	if err != nil {
		return nil, errors.Wrapf(err, "while unmarshalling %s with ViewURL `%s` to GQL JSON", pretty.NavigationNode, in.ViewURL)
	}

	return result, nil
}
