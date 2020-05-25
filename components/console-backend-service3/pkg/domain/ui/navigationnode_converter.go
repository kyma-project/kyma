package ui

import (
	"encoding/json"

	uiV1alpha1v "github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/domain/ui/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/graph/model"
	"github.com/pkg/errors"
)

type navigationNodeConverter struct{}

func (c *navigationNodeConverter) ToGQL(in *uiV1alpha1v.NavigationNode) (*model.NavigationNode, error) {
	if in == nil {
		return nil, nil
	}

	settingsGqlJSON, err := c.settingsToGQLJSON(in)
	if err != nil {
		return nil, err
	}

	requiredPermissions := c.requiredPermissionsToGQLs(in.RequiredPermissions)

	navigationNode := model.NavigationNode{
		Label:               in.Label,
		NavigationPath:      in.NavigationPath,
		ViewURL:             in.ViewURL,
		ShowInNavigation:    in.ShowInNavigation,
		Order:               in.Order,
		Settings:            settingsGqlJSON,
		ExternalLink:        in.ExternalLink,
		RequiredPermissions: requiredPermissions,
	}

	return &navigationNode, nil
}

func (c *navigationNodeConverter) ToGQLs(in []uiV1alpha1v.NavigationNode) ([]*model.NavigationNode, error) {
	var result []*model.NavigationNode
	for _, u := range in {
		converted, err := c.ToGQL(&u)
		if err != nil {
			return nil, err
		}
		if converted != nil {
			result = append(result, converted)
		}
	}
	return result, nil
}

func (c *navigationNodeConverter) requiredPermissionsToGQLs(in []uiV1alpha1v.RequiredPermission) []*model.RequiredPermission {
	var result []*model.RequiredPermission
	for _, u := range in {
		converted := &model.RequiredPermission{
			Verbs:    u.Verbs,
			Resource: u.Resource,
			APIGroup: u.APIGroup,
		}
		result = append(result, converted)
	}
	return result
}

func (c *navigationNodeConverter) settingsToGQLJSON(in *uiV1alpha1v.NavigationNode) (map[string]interface{}, error) {
	if in == nil {
		return nil, nil
	}

	jsonByte, err := json.Marshal(in.Settings)
	if err != nil {
		return nil, errors.Wrapf(err, "while marshalling %s with ViewURL `%s`", pretty.NavigationNode, in.ViewURL)
	}

	var result map[string]interface{}
	err = json.Unmarshal(jsonByte, &result)
	if err != nil {
		return nil, errors.Wrapf(err, "while unmarshalling %s with ViewURL `%s` to map", pretty.NavigationNode, in.ViewURL)
	}
	return result, nil
}
