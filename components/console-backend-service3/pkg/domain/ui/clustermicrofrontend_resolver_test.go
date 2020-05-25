package ui_test

import (
	testingUtils "github.com/kyma-project/kyma/components/console-backend-service3/internal/testing"
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/resource/fake"
	"testing"
	"time"

	uiv1alpha1 "github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/domain/ui"
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/graph/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var clusterMicroFrontendTypeMeta = metav1.TypeMeta{
	APIVersion: v1alpha1.SchemeGroupVersion.String(),
	Kind:       "ClusterMicroFrontend",
}

func TestClusterMicroFrontendResolver_ClusterMicroFrontendsQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		name := "test-name"
		version := "v1"
		category := "test-category"
		viewBaseUrl := "http://test-viewBaseUrl.com"
		preloadUrl := "http://test-preloadUrl.com/#preload"
		placement := "cluster"


		item := &uiv1alpha1.ClusterMicroFrontend{
			TypeMeta: clusterMicroFrontendTypeMeta,
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: uiv1alpha1.ClusterMicroFrontendSpec{
				Placement: placement,
				CommonMicroFrontendSpec: uiv1alpha1.CommonMicroFrontendSpec{
					Version:     version,
					Category:    category,
					ViewBaseURL: viewBaseUrl,
					PreloadURL:  preloadUrl,
					NavigationNodes: []uiv1alpha1.NavigationNode{
						{
							Label:            "test-mf",
							NavigationPath:   "test-path",
							ViewURL:          "/test/viewUrl",
							ShowInNavigation: true,
						},
					},
				},
			},
		}

		expectedItem := &model.ClusterMicroFrontend{
			Name:        name,
			Version:     version,
			Category:    category,
			ViewBaseURL: viewBaseUrl,
			PreloadURL:  preloadUrl,
			Placement:   placement,
			NavigationNodes: []*model.NavigationNode{
				{
					Label:            "test-mf",
					NavigationPath:   "test-path",
					ViewURL:          "/test/viewUrl",
					ShowInNavigation: true,
				},
			},
		}

		expectedItems := []*model.ClusterMicroFrontend{
			expectedItem,
		}

		sf, err := fake.NewFakeServiceFactory(v1alpha1.AddToScheme, item)
		require.NoError(t, err)

		resolver := ui.NewClusterMicroFrontendResolver(sf)
		testingUtils.WaitForInformerFactoryStartAtMost(t, time.Second, sf.InformerFactory)

		result, err := resolver.ClusterMicroFrontendsQuery(nil)

		require.NoError(t, err)
		assert.Equal(t, expectedItems, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		sf, err := fake.NewFakeServiceFactory(v1alpha1.AddToScheme)
		require.NoError(t, err)
		resolver := ui.NewClusterMicroFrontendResolver(sf)
		testingUtils.WaitForInformerFactoryStartAtMost(t, time.Second, sf.InformerFactory)

		var expected []*model.ClusterMicroFrontend

		result, err := resolver.ClusterMicroFrontendsQuery(nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})
}
