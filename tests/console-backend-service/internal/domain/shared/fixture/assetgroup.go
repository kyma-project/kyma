package fixture

import "github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared"

const (
	AssetGroupViewContext = "docs-ui"
	AssetGroupGroupName   = "example-group-name"
	AssetGroupDisplayName = "Asset Group Sample"
	AssetGroupDescription = "Asset Group Description"
)

func AssetGroup(namespace, name string) shared.AssetGroup {
	return shared.AssetGroup{
		Name:        name,
		Namespace:   namespace,
		GroupName:   AssetGroupGroupName,
		DisplayName: AssetGroupDisplayName,
		Description: AssetGroupDescription,
	}
}
