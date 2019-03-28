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

func TestSecretConverter_ToGQL(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t1 := time.Unix(1552643464696, 0)
		// GIVEN
		givenSecret := v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "my-secret",
				Namespace:         "production",
				CreationTimestamp: metav1.NewTime(t1),
				Labels:            map[string]string{"label1": "data"},
				Annotations:       map[string]string{"annotation1": "annotation"},
			},
			Data: map[string][]byte{
				"password": []byte("secret"),
			},
			Type: "custom-type",
		}
		sut := secretConverter{}
		// WHEN
		actualQL, err := sut.ToGQL(&givenSecret)
		require.NoError(t, err)
		// THEN
		assert.Equal(t, "my-secret", actualQL.Name)
		assert.Equal(t, "production", actualQL.Namespace)
		assert.Equal(t, t1, actualQL.CreationTime)
		assert.Equal(t, gqlschema.JSON{"label1": "data"}, actualQL.Labels)
		assert.Equal(t, gqlschema.JSON{"annotation1": "annotation"}, actualQL.Annotations)
		assert.Equal(t, "custom-type", actualQL.Type)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := secretConverter{}
		result, err := converter.ToGQL(&v1.Secret{})
		require.NoError(t, err)
		gqlJson, err := converter.secretToGQLJSON(&v1.Secret{})
		assert.NoError(t, err)
		expected := &gqlschema.Secret{Data: make(gqlschema.JSON), Labels: make(gqlschema.JSON), Annotations: make(gqlschema.JSON), JSON: gqlJson}
		assert.Equal(t, result, expected)
	})

	t.Run("Nil", func(t *testing.T) {
		converter := secretConverter{}
		result, err := converter.ToGQL(nil)
		require.NoError(t, err)

		assert.Nil(t, result)
	})
}

func TestSecretConverter_ToGQLs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t1 := time.Unix(1552643464696, 0)
		// GIVEN
		firstSecret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "my-secret",
				Namespace:         "production",
				CreationTimestamp: metav1.NewTime(t1),
				Labels:            map[string]string{"label1": "data"},
				Annotations:       map[string]string{"annotation1": "annotation"},
			},
			Data: map[string][]byte{
				"password": []byte("secret"),
			},
			Type: "custom-type",
		}

		secondSecret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "second-sec",
				Namespace:         "production",
				CreationTimestamp: metav1.NewTime(t1),
				Labels:            map[string]string{"sec-label": "data"},
				Annotations:       map[string]string{"second-annotation": "content"},
			},
			Data: map[string][]byte{
				"pass": []byte("sec"),
			},
			Type: "second-type",
		}

		sut := secretConverter{}
		// WHEN
		actualQL, err := sut.ToGQLs([]*v1.Secret{firstSecret, secondSecret})
		require.NoError(t, err)
		// THEN
		assert.Equal(t, "my-secret", actualQL[0].Name)
		assert.Equal(t, "production", actualQL[0].Namespace)
		assert.Equal(t, t1, actualQL[0].CreationTime)
		assert.Equal(t, gqlschema.JSON{"label1": "data"}, actualQL[0].Labels)
		assert.Equal(t, gqlschema.JSON{"annotation1": "annotation"}, actualQL[0].Annotations)
		assert.Equal(t, "custom-type", actualQL[0].Type)

		assert.Equal(t, "second-sec", actualQL[1].Name)
		assert.Equal(t, "production", actualQL[1].Namespace)
		assert.Equal(t, t1, actualQL[1].CreationTime)
		assert.Equal(t, gqlschema.JSON{"sec-label": "data"}, actualQL[1].Labels)
		assert.Equal(t, gqlschema.JSON{"second-annotation": "content"}, actualQL[1].Annotations)
		assert.Equal(t, "second-type", actualQL[1].Type)
	})

	t.Run("EmptyList", func(t *testing.T) {
		converter := secretConverter{}
		result, err := converter.ToGQLs([]*v1.Secret{})
		require.NoError(t, err)
		expected := []gqlschema.Secret(nil)
		assert.Equal(t, result, expected)
	})

	t.Run("Nil", func(t *testing.T) {
		converter := secretConverter{}
		result, err := converter.ToGQLs(nil)
		require.NoError(t, err)

		assert.Nil(t, result)
	})
}

func TestSecretConverter_GQLJSONToPod(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		converter := secretConverter{}
		inMap := map[string]interface{}{
			"kind":       "exampleKind",
			"apiversion": "someApiVersion",
			"type":       "custom-type",
			"metadata": map[string]interface{}{
				"name": "exampleName",
				"labels": map[string]interface{}{
					"exampleKey":  "exampleValue",
					"exampleKey2": "exampleValue2",
				},
				"annotations": map[string]interface{}{
					"exampleKeyAnnotation":  "exampleValueAnnotation",
					"exampleKey2Annotation": "exampleValue2Annotation",
				},
				"creationTimestamp": nil,
			},
		}
		expected := v1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind:       "exampleKind",
				APIVersion: "someApiVersion",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "exampleName",
				Labels: map[string]string{
					"exampleKey":  "exampleValue",
					"exampleKey2": "exampleValue2",
				},
				Annotations: map[string]string{
					"exampleKeyAnnotation":  "exampleValueAnnotation",
					"exampleKey2Annotation": "exampleValue2Annotation",
				},
				CreationTimestamp: metav1.Time{},
			},
			Type: "custom-type",
		}

		inJSON := new(gqlschema.JSON)
		err := inJSON.UnmarshalGQL(inMap)
		require.NoError(t, err)

		result, err := converter.GQLJSONToSecret(*inJSON)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})
}
