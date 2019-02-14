package k8s

import (
	"bytes"
	"testing"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/k8s/types"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceConverter_GQLJSONToResource(t *testing.T) {
	const (
		kind       = "Pod"
		apiVersion = "v1"
		name       = "test-pod"
		namespace  = "test-namespace"
	)
	var (
		resourceJSON = gqlschema.JSON{
			"kind":       kind,
			"apiVersion": apiVersion,
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
		}
		invalidJSON = gqlschema.JSON{}
		expected    = types.Resource{
			APIVersion: apiVersion,
			Name:       name,
			Namespace:  namespace,
			Kind:       kind,
			Body:       nil,
		}
	)

	t.Run("Success", func(t *testing.T) {
		converter := &resourceConverter{}
		var buf bytes.Buffer
		resourceJSON.MarshalGQL(&buf)
		expected.Body = buf.Bytes()

		result, err := converter.GQLJSONToResource(resourceJSON)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		converter := &resourceConverter{}

		result, err := converter.GQLJSONToResource(invalidJSON)
		require.Error(t, err)
		assert.Empty(t, result)
	})
}

func TestResourceConverter_BodyToGQLJSON(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		converter := &resourceConverter{}
		expected := gqlschema.JSON{
			"test":  "test",
			"test2": "test2",
		}
		var buf bytes.Buffer
		expected.MarshalGQL(&buf)
		in := buf.Bytes()

		result, err := converter.BodyToGQLJSON(in)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("InvalidBytes", func(t *testing.T) {
		converter := &resourceConverter{}
		in := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 255}

		result, err := converter.BodyToGQLJSON(in)

		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("NilPassed", func(t *testing.T) {
		converter := &resourceConverter{}

		result, err := converter.BodyToGQLJSON(nil)

		require.Error(t, err)
		assert.Nil(t, result)
	})
}
