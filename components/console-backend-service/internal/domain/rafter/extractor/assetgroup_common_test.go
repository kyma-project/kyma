package extractor_test

import (
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/extractor"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAssetGroupCommonExtractor_Status(t *testing.T) {
	t.Run("Pending", func(t *testing.T) {
		// given
		status := v1beta1.CommonAssetGroupStatus{Phase: v1beta1.AssetGroupPending, Reason: "test reason", Message: "test message"}
		expected := gqlschema.AssetGroupStatus{Phase: gqlschema.AssetGroupPhaseTypePending, Reason: "test reason", Message: "test message"}
		converter := new(extractor.AssetGroupCommonExtractor)

		// when
		result := converter.Status(status)

		// then
		assert.Equal(t, expected, result)
	})

	t.Run("Ready", func(t *testing.T) {
		// given
		status := v1beta1.CommonAssetGroupStatus{Phase: v1beta1.AssetGroupReady, Reason: "test reason", Message: "test message"}
		expected := gqlschema.AssetGroupStatus{Phase: gqlschema.AssetGroupPhaseTypeReady, Reason: "test reason", Message: "test message"}
		converter := new(extractor.AssetGroupCommonExtractor)

		// when
		result := converter.Status(status)

		// then
		assert.Equal(t, expected, result)
	})

	t.Run("Failed", func(t *testing.T) {
		// given
		status := v1beta1.CommonAssetGroupStatus{Phase: v1beta1.AssetGroupFailed, Reason: "test reason", Message: "test message"}
		expected := gqlschema.AssetGroupStatus{Phase: gqlschema.AssetGroupPhaseTypeFailed, Reason: "test reason", Message: "test message"}
		converter := new(extractor.AssetGroupCommonExtractor)

		// when
		result := converter.Status(status)

		// then
		assert.Equal(t, expected, result)
	})

	t.Run("Phase unknown", func(t *testing.T) {
		// given
		status := v1beta1.CommonAssetGroupStatus{Reason: "test reason", Message: "test message"}
		expected := gqlschema.AssetGroupStatus{Phase: gqlschema.AssetGroupPhaseTypeFailed, Reason: "test reason", Message: "test message"}
		converter := new(extractor.AssetGroupCommonExtractor)

		// when
		result := converter.Status(status)

		// then
		assert.Equal(t, expected, result)
	})
}
