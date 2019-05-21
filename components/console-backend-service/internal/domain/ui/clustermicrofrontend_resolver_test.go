package ui_test

import (
	"errors"
	"testing"

	"github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/ui"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/ui/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestClusterMicroFrontendResolver_ClusterMicroFrontendsQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		name := "test-name"
		version := "v1"
		category := "test-category"
		viewBaseUrl := "http://test-viewBaseUrl.com"
		placement := "cluster"

		item := &v1alpha1.ClusterMicroFrontend{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: v1alpha1.ClusterMicroFrontendSpec{
				Placement: placement,
				CommonMicroFrontendSpec: v1alpha1.CommonMicroFrontendSpec{
					Version:     version,
					Category:    category,
					ViewBaseURL: viewBaseUrl,
					NavigationNodes: []v1alpha1.NavigationNode{
						v1alpha1.NavigationNode{
							Label:            "test-mf",
							NavigationPath:   "test-path",
							ViewURL:          "/test/viewUrl",
							ShowInNavigation: true,
						},
					},
				},
			},
		}

		items := []*v1alpha1.ClusterMicroFrontend{
			item, item,
		}

		expectedItem := gqlschema.ClusterMicroFrontend{
			Name:        name,
			Version:     version,
			Category:    category,
			ViewBaseURL: viewBaseUrl,
			Placement:   placement,
			NavigationNodes: []gqlschema.NavigationNode{
				gqlschema.NavigationNode{
					Label:            "test-mf",
					NavigationPath:   "test-path",
					ViewURL:          "/test/viewUrl",
					ShowInNavigation: true,
				},
			},
		}

		expectedItems := []gqlschema.ClusterMicroFrontend{
			expectedItem, expectedItem,
		}

		resourceGetter := automock.NewClusterMicroFrontendService()
		resourceGetter.On("List").Return(items, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewClusterMicroFrontendConverter()
		converter.On("ToGQLs", items).Return(expectedItems, nil)
		defer converter.AssertExpectations(t)

		resolver := ui.NewClusterMicroFrontendResolver(resourceGetter)
		resolver.SetClusterMicroFrontendConverter(converter)

		result, err := resolver.ClusterMicroFrontendsQuery(nil)

		require.NoError(t, err)
		assert.Equal(t, expectedItems, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		var items []*v1alpha1.ClusterMicroFrontend

		resourceGetter := automock.NewClusterMicroFrontendService()
		resourceGetter.On("List").Return(items, nil).Once()
		defer resourceGetter.AssertExpectations(t)
		resolver := ui.NewClusterMicroFrontendResolver(resourceGetter)
		var expected []gqlschema.ClusterMicroFrontend

		result, err := resolver.ClusterMicroFrontendsQuery(nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Error", func(t *testing.T) {
		expected := errors.New("Test")

		var items []*v1alpha1.ClusterMicroFrontend

		resourceGetter := automock.NewClusterMicroFrontendService()
		resourceGetter.On("List").Return(items, expected).Once()
		defer resourceGetter.AssertExpectations(t)
		resolver := ui.NewClusterMicroFrontendResolver(resourceGetter)

		_, err := resolver.ClusterMicroFrontendsQuery(nil)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}
