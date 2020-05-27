package ui_test

import (
	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/graph/model"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource/fake"
	"testing"
	"time"

	"github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/domain/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var microFrontendTypeMeta = metav1.TypeMeta{
	APIVersion: v1alpha1.SchemeGroupVersion.String(),
	Kind:       "MicroFrontend",
}

func TestMicroFrontendResolver_MicroFrontendsQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		name := "test-name"
		namespace := "test-namespace"
		version := "v1"
		category := "test-category"
		viewBaseUrl := "http://test-viewBaseUrl.com"

		item := &v1alpha1.MicroFrontend{
			TypeMeta: microFrontendTypeMeta,
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

		expectedItem := &model.MicroFrontend{
			Name:        name,
			Version:     version,
			Category:    category,
			ViewBaseURL: viewBaseUrl,
			NavigationNodes: []*model.NavigationNode{
				{
					Label:            "test-mf",
					NavigationPath:   "test-path",
					ViewURL:          "/test/viewUrl",
					ShowInNavigation: true,
				},
			},
		}

		expectedItems := []*model.MicroFrontend{
			expectedItem,
		}

		sf, err := fake.NewFakeServiceFactory(v1alpha1.AddToScheme, item)
		require.NoError(t, err)

		resolver := ui.NewMicroFrontendResolver(sf)
		testingUtils.WaitForInformerFactoryStartAtMost(t, 10*time.Second, sf.InformerFactory)

		result, err := resolver.MicroFrontendsQuery(nil, namespace)

		require.NoError(t, err)
		assert.Equal(t, expectedItems, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		namespace := "test-namespace"

		sf, err := fake.NewFakeServiceFactory(v1alpha1.AddToScheme)
		require.NoError(t, err)
		resolver := ui.NewMicroFrontendResolver(sf)
		testingUtils.WaitForInformerFactoryStartAtMost(t, time.Second, sf.InformerFactory)

		var expected []*model.MicroFrontend

		result, err := resolver.MicroFrontendsQuery(nil, namespace)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})
}
