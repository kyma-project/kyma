package fixture

import "github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared"

func ClusterAssetGroup(name string) shared.ClusterAssetGroup {
	return shared.ClusterAssetGroup{
		Name:        name,
		GroupName:   AssetGroupGroupName,
		DisplayName: AssetGroupDisplayName,
		Description: AssetGroupDescription,
	}
}
