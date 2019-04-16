package fixture

import "github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared"

func ClusterAsset(typeArg string) shared.ClusterAsset {
	return shared.ClusterAsset{
		Type: typeArg,
	}
}
