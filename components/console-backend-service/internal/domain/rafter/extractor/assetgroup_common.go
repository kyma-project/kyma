package extractor

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
)

type AssetGroupCommonExtractor struct{}

func (e *AssetGroupCommonExtractor) Status(status v1beta1.CommonAssetGroupStatus) gqlschema.AssetGroupStatus {
	return gqlschema.AssetGroupStatus{
		Phase:   e.phase(status.Phase),
		Reason:  string(status.Reason),
		Message: status.Message,
	}
}

func (e *AssetGroupCommonExtractor) phase(phase v1beta1.AssetGroupPhase) gqlschema.AssetGroupPhaseType {
	switch phase {
	case v1beta1.AssetGroupReady:
		return gqlschema.AssetGroupPhaseTypeReady
	case v1beta1.AssetGroupPending:
		return gqlschema.AssetGroupPhaseTypePending
	default:
		return gqlschema.AssetGroupPhaseTypeFailed
	}
}
