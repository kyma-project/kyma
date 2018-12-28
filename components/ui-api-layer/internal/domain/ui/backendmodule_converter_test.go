package ui

import (
	"testing"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/apis/ui/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBackendModuleConverter_ToGQL(t *testing.T) {
	t.Run("All properties are given", func(t *testing.T) {
		converter := backendModuleConverter{}
		name := "test-name"

		item := v1alpha1.BackendModule{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		}
		expected := gqlschema.BackendModule{
			Name: name,
		}

		result, err := converter.ToGQL(&item)

		require.NoError(t, err)
		assert.Equal(t, &expected, result)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := &backendModuleConverter{}
		_, err := converter.ToGQL(&v1alpha1.BackendModule{})
		require.NoError(t, err)
	})

	t.Run("Nil", func(t *testing.T) {
		converter := &backendModuleConverter{}
		item, err := converter.ToGQL(nil)

		require.NoError(t, err)
		assert.Nil(t, item)
	})
}

func TestBackendModuleConverter_ToGQLs(t *testing.T) {
	name := "example-name"
	module := v1alpha1.BackendModule{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	t.Run("Success", func(t *testing.T) {
		instances := []*v1alpha1.BackendModule{
			&module,
			&module,
		}

		converter := backendModuleConverter{}
		result, err := converter.ToGQLs(instances)

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, name, result[0].Name)
	})

	t.Run("Empty", func(t *testing.T) {
		var instances []*v1alpha1.BackendModule

		converter := backendModuleConverter{}
		result, err := converter.ToGQLs(instances)

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("With nil", func(t *testing.T) {
		instances := []*v1alpha1.BackendModule{
			nil,
			&module,
			nil,
		}

		converter := backendModuleConverter{}
		result, err := converter.ToGQLs(instances)

		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, name, result[0].Name)
	})
}
