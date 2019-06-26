package extractor_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/assetstore/extractor"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestCommon_Metadata(t *testing.T) {
	t.Run("Filled", func(t *testing.T) {
		// given
		metadata := &runtime.RawExtension{Raw: []byte(`{"json":"true","complex":{"data":"true"}}`)}
		expected := map[string]interface{}{"complex": map[string]interface{}{"data": "true"}, "json": "true"}
		converter := new(extractor.Common)

		// when
		result, err := converter.Metadata(metadata)

		// then
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Empty", func(t *testing.T) {
		// given
		metadata := &runtime.RawExtension{Raw: []byte(`{}`)}
		expected := make(map[string]interface{})
		converter := new(extractor.Common)

		// when
		result, err := converter.Metadata(metadata)

		// then
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		expected := make(map[string]interface{})
		converter := new(extractor.Common)

		// when
		result, err := converter.Metadata(nil)

		// then
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Invalid", func(t *testing.T) {
		// given
		metadata := &runtime.RawExtension{Raw: []byte(`{invalid`)}
		converter := new(extractor.Common)

		// when
		_, err := converter.Metadata(metadata)

		// then
		require.Error(t, err)
	})
}

func TestCommon_Status(t *testing.T) {
	t.Run("Pending", func(t *testing.T) {
		// given
		status := v1alpha2.CommonAssetStatus{Phase: v1alpha2.AssetPending, Reason: "test reason", Message: "test message"}
		expected := gqlschema.AssetStatus{Phase: gqlschema.AssetPhaseTypePending, Reason: "test reason", Message: "test message"}
		converter := new(extractor.Common)

		// when
		result := converter.Status(status)

		// then
		assert.Equal(t, expected, result)
	})

	t.Run("Ready", func(t *testing.T) {
		// given
		status := v1alpha2.CommonAssetStatus{Phase: v1alpha2.AssetReady, Reason: "test reason", Message: "test message"}
		expected := gqlschema.AssetStatus{Phase: gqlschema.AssetPhaseTypeReady, Reason: "test reason", Message: "test message"}
		converter := new(extractor.Common)

		// when
		result := converter.Status(status)

		// then
		assert.Equal(t, expected, result)
	})

	t.Run("Failed", func(t *testing.T) {
		// given
		status := v1alpha2.CommonAssetStatus{Phase: v1alpha2.AssetFailed, Reason: "test reason", Message: "test message"}
		expected := gqlschema.AssetStatus{Phase: gqlschema.AssetPhaseTypeFailed, Reason: "test reason", Message: "test message"}
		converter := new(extractor.Common)

		// when
		result := converter.Status(status)

		// then
		assert.Equal(t, expected, result)
	})

	t.Run("Phase unknown", func(t *testing.T) {
		// given
		status := v1alpha2.CommonAssetStatus{Reason: "test reason", Message: "test message"}
		expected := gqlschema.AssetStatus{Phase: gqlschema.AssetPhaseTypeFailed, Reason: "test reason", Message: "test message"}
		converter := new(extractor.Common)

		// when
		result := converter.Status(status)

		// then
		assert.Equal(t, expected, result)
	})
}
