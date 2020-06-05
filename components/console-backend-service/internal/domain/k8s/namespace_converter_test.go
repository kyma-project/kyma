package k8s

import (
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNamespaceConverter_ToGQLs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expectedName := "exampleName"
		expectedName2 := "exampleName2"
		converter := namespaceConverter{
			systemNamespaces: []string{expectedName},
		}
		in := []*v1.Namespace{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: expectedName,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: expectedName2,
				},
			},
		}

		result := converter.ToGQLs(in)

		assert.Len(t, result, 2)
		assert.Equal(t, expectedName, result[0].Name)
		assert.Equal(t, expectedName2, result[1].Name)
		assert.True(t, result[0].IsSystemNamespace)
		assert.False(t, result[1].IsSystemNamespace)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := namespaceConverter{}
		var in []*v1.Namespace

		result := converter.ToGQLs(in)

		assert.Empty(t, result)
	})

	t.Run("With nil", func(t *testing.T) {
		expectedName := "exampleName"
		expectedLabels := gqlschema.Labels{"test": "label"}
		converter := namespaceConverter{
			systemNamespaces: []string{expectedName},
		}
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

		result := converter.ToGQLs(in)

		assert.Len(t, result, 1)
		assert.Equal(t, expectedName, result[0].Name)
		assert.Equal(t, expectedLabels, result[0].Labels)
		assert.True(t, result[0].IsSystemNamespace)
	})
}

func TestNamespaceConverter_ToGQL(t *testing.T) {
	t.Run("Success for system namespace", func(t *testing.T) {
		expectedName := "exampleName"
		expectedLabels := gqlschema.Labels{"test": "label"}
		converter := namespaceConverter{
			systemNamespaces: []string{expectedName},
		}

		in := v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   expectedName,
				Labels: map[string]string{"test": "label"},
			},
		}

		result := converter.ToGQL(&in)

		assert.Equal(t, expectedName, result.Name)
		assert.Equal(t, expectedLabels, result.Labels)
		assert.True(t, result.IsSystemNamespace)
	})

	t.Run("Success", func(t *testing.T) {
		expectedName := "exampleName"
		expectedLabels := gqlschema.Labels{"test": "label"}
		converter := namespaceConverter{
			systemNamespaces: []string{"systemNamespace"},
		}

		in := v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   expectedName,
				Labels: map[string]string{"test": "label"},
			},
		}

		result := converter.ToGQL(&in)

		assert.Equal(t, expectedName, result.Name)
		assert.Equal(t, expectedLabels, result.Labels)
		assert.False(t, result.IsSystemNamespace)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := namespaceConverter{}
		var in *v1.Namespace

		result := converter.ToGQL(in)

		assert.Empty(t, result)
	})
}
