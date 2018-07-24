package k8s

import (
	"testing"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSecretConverter_ToGQL(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		givenSecret := v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-secret",
				Namespace: "production",
			},
			Data: map[string][]byte{
				"password": []byte("secret"),
			},
		}
		sut := secretConverter{}
		// WHEN
		actualQL := sut.ToGQL(&givenSecret)
		// THEN
		assert.Equal(t, "my-secret", actualQL.Name)
		assert.Equal(t, "production", actualQL.Environment)
		assert.Equal(t, gqlschema.JSON{"password": "secret"}, actualQL.Data)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := secretConverter{}
		converter.ToGQL(&v1.Secret{})
	})

	t.Run("Nil", func(t *testing.T) {
		converter := secretConverter{}
		result := converter.ToGQL(nil)

		assert.Nil(t, result)
	})
}
