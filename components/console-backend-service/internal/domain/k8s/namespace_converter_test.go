package k8s

import (
	"testing"

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
		in := []*v1.Namespace{
			nil,
			&v1.Namespace{
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

func TestNamespaceConverter_ToGQL(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		converter := namespaceConverter{}
		expectedName := "exampleName"
		in := v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: expectedName,
			},
		}

		result, err := converter.ToGQL(&in)

		require.NoError(t, err)
		assert.Equal(t, expectedName, result.Name)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := namespaceConverter{}
		var in *v1.Namespace

		result, err := converter.ToGQL(in)

		require.NoError(t, err)
		assert.Empty(t, result)
	})
}
