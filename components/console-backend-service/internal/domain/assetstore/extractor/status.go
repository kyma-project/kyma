package extractor

import (
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

type AssetStatusExtractor struct{}

func (e *AssetStatusExtractor) Status(status v1alpha2.CommonAssetStatus) gqlschema.AssetStatus {
	return gqlschema.AssetStatus{
		Phase:   e.phase(status.Phase),
		Reason:  status.Reason,
		Message: status.Message,
	}
}

func (e *AssetStatusExtractor) phase(phase v1alpha2.AssetPhase) gqlschema.AssetPhaseType {
	switch phase {
	case v1alpha2.AssetReady:
		return gqlschema.AssetPhaseTypeReady
	case v1alpha2.AssetPending:
		return gqlschema.AssetPhaseTypePending
	default:
		return gqlschema.AssetPhaseTypeFailed
	}
}
