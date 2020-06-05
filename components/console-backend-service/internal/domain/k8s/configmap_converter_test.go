package k8s

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConfigMapConverter_ToGQL(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		converter := &configMapConverter{}
		in := v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "exampleName",
				Namespace:         "exampleNamespace",
				CreationTimestamp: metav1.Time{},
				Labels: map[string]string{
					"exampleKey":  "exampleValue",
					"exampleKey2": "exampleValue2",
				},
			},
		}
		expectedJSON, err := converter.configMapToGQLJSON(&in)
		require.NoError(t, err)
		expected := gqlschema.ConfigMap{
			Name:              "exampleName",
			Namespace:         "exampleNamespace",
			CreationTimestamp: time.Time{},
			Labels: map[string]string{
				"exampleKey":  "exampleValue",
				"exampleKey2": "exampleValue2",
			},
			JSON: expectedJSON,
		}

		result, err := converter.ToGQL(&in)

		require.NoError(t, err)
		assert.Equal(t, &expected, result)

	})

	t.Run("Empty", func(t *testing.T) {
		converter := &configMapConverter{}
		emptyConfigMapJSON, err := converter.configMapToGQLJSON(&v1.ConfigMap{})
		require.NoError(t, err)
		expected := &gqlschema.ConfigMap{
			JSON: emptyConfigMapJSON,
		}

		result, err := converter.ToGQL(&v1.ConfigMap{})

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Nil", func(t *testing.T) {
		converter := &configMapConverter{}

		result, err := converter.ToGQL(nil)

		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestConfigMapConverter_ToGQLs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		converter := configMapConverter{}
		expectedName := "exampleName"
		in := []*v1.ConfigMap{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: expectedName,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "exampleName2",
				},
			},
		}

		result, err := converter.ToGQLs(in)

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, expectedName, result[0].Name)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := configMapConverter{}
		var in []*v1.ConfigMap

		result, err := converter.ToGQLs(in)

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("With nil", func(t *testing.T) {
		converter := configMapConverter{}
		expectedName := "exampleName"
		in := []*v1.ConfigMap{
			nil,
			&v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: expectedName,
				},
			},
			nil,
		}

		result, err := converter.ToGQLs(in)

		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, expectedName, result[0].Name)
	})
}

func TestConfigMapConverter_ConfigMapToGQLJSON(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		converter := configMapConverter{}
		expectedMap := map[string]interface{}{
			"kind": "exampleKind",
			"metadata": map[string]interface{}{
				"name": "exampleName",
				"labels": map[string]interface{}{
					"exampleKey":  "exampleValue",
					"exampleKey2": "exampleValue2",
				},
				"creationTimestamp": nil,
			},
		}
		in := v1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				Kind: "exampleKind",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "exampleName",
				Labels: map[string]string{
					"exampleKey":  "exampleValue",
					"exampleKey2": "exampleValue2",
				},
				CreationTimestamp: metav1.Time{},
			},
		}

		expectedJSON := new(gqlschema.JSON)
		err := expectedJSON.UnmarshalGQL(expectedMap)
		require.NoError(t, err)

		result, err := converter.configMapToGQLJSON(&in)

		require.NoError(t, err)
		assert.Equal(t, *expectedJSON, result)
	})

	t.Run("NilPassed", func(t *testing.T) {
		converter := configMapConverter{}

		result, err := converter.configMapToGQLJSON(nil)

		require.Nil(t, result)
		require.NoError(t, err)
	})
}

func TestConfigMapConverter_GQLJSONToConfigMap(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		converter := configMapConverter{}
		inMap := map[string]interface{}{
			"kind": "exampleKind",
			"metadata": map[string]interface{}{
				"name": "exampleName",
				"labels": map[string]interface{}{
					"exampleKey":  "exampleValue",
					"exampleKey2": "exampleValue2",
				},
				"creationTimestamp": nil,
			},
		}
		expected := v1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				Kind: "exampleKind",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "exampleName",
				Labels: map[string]string{
					"exampleKey":  "exampleValue",
					"exampleKey2": "exampleValue2",
				},
				CreationTimestamp: metav1.Time{},
			},
		}

		inJSON := new(gqlschema.JSON)
		err := inJSON.UnmarshalGQL(inMap)
		require.NoError(t, err)

		result, err := converter.GQLJSONToConfigMap(*inJSON)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})
}
