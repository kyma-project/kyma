package extractor_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/extractor"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestAssetCommonExtractor_Metadata(t *testing.T) {
	t.Run("Filled", func(t *testing.T) {
		// given
		parameters := &runtime.RawExtension{Raw: []byte(`{"json":"true","complex":{"data":"true"}}`)}
		expected := map[string]interface{}{"complex": map[string]interface{}{"data": "true"}, "json": "true"}
		converter := new(extractor.AssetCommonExtractor)

		// when
		result, err := converter.Parameters(parameters)

		// then
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Empty", func(t *testing.T) {
		// given
		parameters := &runtime.RawExtension{Raw: []byte(`{}`)}
		expected := make(map[string]interface{})
		converter := new(extractor.AssetCommonExtractor)

		// when
		result, err := converter.Parameters(parameters)

		// then
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		expected := make(map[string]interface{})
		converter := new(extractor.AssetCommonExtractor)

		// when
		result, err := converter.Parameters(nil)

		// then
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Invalid", func(t *testing.T) {
		// given
		parameters := &runtime.RawExtension{Raw: []byte(`{invalid`)}
		converter := new(extractor.AssetCommonExtractor)

		// when
		_, err := converter.Parameters(parameters)

		// then
		require.Error(t, err)
	})
}

func TestAssetCommonExtractor_Status(t *testing.T) {
	t.Run("Pending", func(t *testing.T) {
		// given
		status := v1beta1.CommonAssetStatus{Phase: v1beta1.AssetPending, Reason: "test reason", Message: "test message"}
		expected := gqlschema.AssetStatus{Phase: gqlschema.AssetPhaseTypePending, Reason: "test reason", Message: "test message"}
		converter := new(extractor.AssetCommonExtractor)

		// when
		result := converter.Status(status)

		// then
		assert.Equal(t, expected, result)
	})

	t.Run("Ready", func(t *testing.T) {
		// given
		status := v1beta1.CommonAssetStatus{Phase: v1beta1.AssetReady, Reason: "test reason", Message: "test message"}
		expected := gqlschema.AssetStatus{Phase: gqlschema.AssetPhaseTypeReady, Reason: "test reason", Message: "test message"}
		converter := new(extractor.AssetCommonExtractor)

		// when
		result := converter.Status(status)

		// then
		assert.Equal(t, expected, result)
	})

	t.Run("Failed", func(t *testing.T) {
		// given
		status := v1beta1.CommonAssetStatus{Phase: v1beta1.AssetFailed, Reason: "test reason", Message: "test message"}
		expected := gqlschema.AssetStatus{Phase: gqlschema.AssetPhaseTypeFailed, Reason: "test reason", Message: "test message"}
		converter := new(extractor.AssetCommonExtractor)

		// when
		result := converter.Status(status)

		// then
		assert.Equal(t, expected, result)
	})

	t.Run("Phase unknown", func(t *testing.T) {
		// given
		status := v1beta1.CommonAssetStatus{Reason: "test reason", Message: "test message"}
		expected := gqlschema.AssetStatus{Phase: gqlschema.AssetPhaseTypeFailed, Reason: "test reason", Message: "test message"}
		converter := new(extractor.AssetCommonExtractor)

		// when
		result := converter.Status(status)

		// then
		assert.Equal(t, expected, result)
	})
}
