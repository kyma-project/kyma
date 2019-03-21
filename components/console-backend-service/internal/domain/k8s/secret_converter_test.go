package k8s

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
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
		actualQL := sut.ToGQL(&givenSecret)
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
		result := converter.ToGQL(&v1.Secret{})
		expected := &gqlschema.Secret{Data: make(gqlschema.JSON), Labels: make(gqlschema.JSON), Annotations: make(gqlschema.JSON)}
		assert.Equal(t, result, expected)
	})

	t.Run("Nil", func(t *testing.T) {
		converter := secretConverter{}
		result := converter.ToGQL(nil)

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
		actualQL := sut.ToGQLs([]*v1.Secret{firstSecret, secondSecret})
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
		result := converter.ToGQLs([]*v1.Secret{})
		expected := []gqlschema.Secret(nil)
		assert.Equal(t, result, expected)
	})

	t.Run("Nil", func(t *testing.T) {
		converter := secretConverter{}
		result := converter.ToGQLs(nil)

		assert.Nil(t, result)
	})
}
