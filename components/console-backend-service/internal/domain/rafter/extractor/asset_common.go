package extractor

import (
	"encoding/json"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

type AssetCommonExtractor struct{}

func (e *AssetCommonExtractor) Status(status v1beta1.CommonAssetStatus) gqlschema.AssetStatus {
	return gqlschema.AssetStatus{
		Phase:   e.phase(status.Phase),
		Reason:  string(status.Reason),
		Message: status.Message,
	}
}

func (*AssetCommonExtractor) Parameters(parameters *runtime.RawExtension) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	if parameters == nil {
		return result, nil
	}

	err := json.Unmarshal(parameters.Raw, &result)
	if err != nil {
		return nil, errors.Wrap(err, "while unmarshalling parameters")
	}

	return result, nil
}

func (e *AssetCommonExtractor) phase(phase v1beta1.AssetPhase) gqlschema.AssetPhaseType {
	switch phase {
	case v1beta1.AssetReady:
		return gqlschema.AssetPhaseTypeReady
	case v1beta1.AssetPending:
		return gqlschema.AssetPhaseTypePending
	default:
		return gqlschema.AssetPhaseTypeFailed
	}
}
