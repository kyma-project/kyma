package k8s

import (
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNamespaceConverter_ToGQLs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		converter := namespaceConverter{}
		expectedName := "exampleName"
		in := []*v1.Namespace{
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
		converter := namespaceConverter{}
		var in []*v1.Namespace

		result, err := converter.ToGQLs(in)

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("With nil", func(t *testing.T) {
		converter := namespaceConverter{}
		expectedName := "exampleName"
		expectedLabels := gqlschema.Labels{"test": "label"}
		in := []*v1.Namespace{
			nil,
			&v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   expectedName,
					Labels: map[string]string{"test": "label"},
				},
			},
			nil,
		}

		result, err := converter.ToGQLs(in)

		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, expectedName, result[0].Name)
		assert.Equal(t, expectedLabels, result[0].Labels)
	})
}

func TestNamespaceConverter_ToGQL(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		converter := namespaceConverter{}
		expectedName := "exampleName"
		expectedLabels := gqlschema.Labels{"test": "label"}
		in := v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   expectedName,
				Labels: map[string]string{"test": "label"},
			},
		}

		result, err := converter.ToGQL(&in)

		require.NoError(t, err)
		assert.Equal(t, expectedName, result.Name)
		assert.Equal(t, expectedLabels, result.Labels)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := namespaceConverter{}
		var in *v1.Namespace

		result, err := converter.ToGQL(in)

		require.NoError(t, err)
		assert.Empty(t, result)
	})
}
