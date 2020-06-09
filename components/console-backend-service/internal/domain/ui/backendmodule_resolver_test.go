package ui_test

import (
	"errors"
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/ui"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/ui/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/apis/ui/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBackendModuleResolver_BackendModulesQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		resource :=
			&v1alpha1.BackendModule{
				ObjectMeta: v1.ObjectMeta{
					Name: "Test",
				},
			}
		resources := []*v1alpha1.BackendModule{
			resource, resource,
		}
		expected := []*gqlschema.BackendModule{
			{
				Name: "Test",
			},
			{
				Name: "Test",
			},
		}

		resourceGetter := automock.NewBackendModuleService()
		resourceGetter.On("List").Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewBackendModuleConverter()
		converter.On("ToGQLs", resources).Return(expected, nil)
		defer converter.AssertExpectations(t)

		resolver := ui.NewBackendModuleResolver(resourceGetter)
		resolver.SetInstanceConverter(converter)

		result, err := resolver.BackendModulesQuery(nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		var resources []*v1alpha1.BackendModule

		resourceGetter := automock.NewBackendModuleService()
		resourceGetter.On("List").Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)
		resolver := ui.NewBackendModuleResolver(resourceGetter)
		var expected []*gqlschema.BackendModule

		result, err := resolver.BackendModulesQuery(nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Error", func(t *testing.T) {
		expected := errors.New("Test")

		var resources []*v1alpha1.BackendModule

		resourceGetter := automock.NewBackendModuleService()
		resourceGetter.On("List").Return(resources, expected).Once()
		defer resourceGetter.AssertExpectations(t)
		resolver := ui.NewBackendModuleResolver(resourceGetter)

		_, err := resolver.BackendModulesQuery(nil)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}
