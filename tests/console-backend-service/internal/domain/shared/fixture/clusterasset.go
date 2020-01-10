package fixture

import (
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
)

func ClusterAsset(typeArg v1beta1.AssetGroupSourceType) shared.ClusterAsset {
	return shared.ClusterAsset{
		Type: typeArg,
	}
}
