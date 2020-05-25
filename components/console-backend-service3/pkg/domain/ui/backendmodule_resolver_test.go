package ui_test

import (
	testingUtils "github.com/kyma-project/kyma/components/console-backend-service3/internal/testing"
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/resource/fake"
	"testing"
	"time"

	uiv1alpha1 "github.com/kyma-project/kyma/components/console-backend-service3/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/domain/ui"
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/graph/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBackendModuleResolver_BackendModulesQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		resource := &uiv1alpha1.BackendModule{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "ui.kyma-project.io/v1alpha1",
				Kind:       "BackendModule",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "Test",
			},
		}
		expected := []*model.BackendModule{
			{
				Name: "Test",
			},
		}

		sf, err := fake.NewFakeServiceFactory(uiv1alpha1.AddToScheme, resource)
		require.NoError(t, err)

		resolver := ui.NewBackendModuleResolver(sf)
		testingUtils.WaitForInformerFactoryStartAtMost(t, time.Second, sf.InformerFactory)

		result, err := resolver.BackendModulesQuery(nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		sf, err := fake.NewFakeServiceFactory(uiv1alpha1.AddToScheme)
		require.NoError(t, err)

		resolver := ui.NewBackendModuleResolver(sf)
		var expected []*model.BackendModule

		result, err := resolver.BackendModulesQuery(nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})
}
