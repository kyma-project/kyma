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

func TestMicroFrontendResolver_MicroFrontendsQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		name := "test-name"
		namespace := "test-namespace"
		version := "v1"
		category := "test-category"
		viewBaseUrl := "http://test-viewBaseUrl.com"

		item := &v1alpha1.MicroFrontend{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: v1alpha1.MicroFrontendSpec{
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

		items := []*v1alpha1.MicroFrontend{
			item, item,
		}

		expectedItem := gqlschema.MicroFrontend{
			Name:        name,
			Version:     version,
			Category:    category,
			ViewBaseURL: viewBaseUrl,
			NavigationNodes: []gqlschema.NavigationNode{
				gqlschema.NavigationNode{
					Label:            "test-mf",
					NavigationPath:   "test-path",
					ViewURL:          "/test/viewUrl",
					ShowInNavigation: true,
				},
			},
		}

		expectedItems := []gqlschema.MicroFrontend{
			expectedItem, expectedItem,
		}

		resourceGetter := automock.NewMicroFrontendService()
		resourceGetter.On("List", namespace).Return(items, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewMicroFrontendConverter()
		converter.On("ToGQLs", items).Return(expectedItems, nil)
		defer converter.AssertExpectations(t)

		resolver := ui.NewMicroFrontendResolver(resourceGetter)
		resolver.SetMicroFrontendConverter(converter)

		result, err := resolver.MicroFrontendsQuery(nil, namespace)

		require.NoError(t, err)
		assert.Equal(t, expectedItems, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		namespace := "test-namespace"
		var items []*v1alpha1.MicroFrontend

		resourceGetter := automock.NewMicroFrontendService()
		resourceGetter.On("List", namespace).Return(items, nil).Once()
		defer resourceGetter.AssertExpectations(t)
		resolver := ui.NewMicroFrontendResolver(resourceGetter)
		var expected []gqlschema.MicroFrontend

		result, err := resolver.MicroFrontendsQuery(nil, namespace)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Error", func(t *testing.T) {
		namespace := "test-namespace"
		expected := errors.New("Test")

		var items []*v1alpha1.MicroFrontend

		resourceGetter := automock.NewMicroFrontendService()
		resourceGetter.On("List", namespace).Return(items, expected).Once()
		defer resourceGetter.AssertExpectations(t)
		resolver := ui.NewMicroFrontendResolver(resourceGetter)

		_, err := resolver.MicroFrontendsQuery(nil, namespace)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}
