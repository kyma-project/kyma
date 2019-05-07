package ui

import (
	"encoding/json"

	uiV1alpha1v "github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/ui/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
)

type clusterMicrofrontendConverter struct{}

func (c *clusterMicrofrontendConverter) ToGQL(in *uiV1alpha1v.ClusterMicroFrontend) *gqlschema.ClusterMicrofrontend {
	if in == nil {
		return nil
	}

	navigationNodes := c.navigationNodesToGQLs(in.Spec.CommonMicroFrontendSpec.NavigationNodes)

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

func (c *clusterMicrofrontendConverter) navigationNodeToGQL(in *uiV1alpha1v.NavigationNode) *gqlschema.NavigationNode {
	if in == nil {
		return nil
	}

	settingsGqlJSON, _ := c.settingsToGQLJSON(in)

	navigationNode := gqlschema.NavigationNode{
		Label:            in.Label,
		NavigationPath:   in.NavigationPath,
		ViewURL:          in.ViewURL,
		ShowInNavigation: in.ShowInNavigation,
		Order:            in.Order,
		Settings:         settingsGqlJSON,
	}

	return &navigationNode
}

func (c *clusterMicrofrontendConverter) navigationNodesToGQLs(in []uiV1alpha1v.NavigationNode) []gqlschema.NavigationNode {
	var result []gqlschema.NavigationNode
	for _, u := range in {
		converted := c.navigationNodeToGQL(&u)
		if converted != nil {
			result = append(result, *converted)
		}
	}
	return result
}

func (c *clusterMicrofrontendConverter) settingsToGQLJSON(in *uiV1alpha1v.NavigationNode) (gqlschema.Settings, error) {
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
