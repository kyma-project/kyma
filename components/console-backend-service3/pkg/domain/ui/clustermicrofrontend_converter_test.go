package ui

import (
	"testing"

	uiv1alpha1 "github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/graph/model"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestClusterMicroFrontendConverter_ToGQL(t *testing.T) {
	t.Run("All properties are given", func(t *testing.T) {
		converter := newClusterMicroFrontendConverter()
		name := "test-name"
		version := "v1"
		category := "test-category"
		viewBaseUrl := "http://test-viewBaseUrl.com"
		preloadUrl := "http://test-preloadUrl.com/#preload"
		placement := "cluster"
		navigationNode := fixNavigationNode(t)

		item := uiv1alpha1.ClusterMicroFrontend{
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
						navigationNode,
					},
				},
			},
		}

		expectedNavigationNode := fixGqlNavigationNode()

		expected := model.ClusterMicroFrontend{
			Name:        name,
			Version:     version,
			Category:    category,
			ViewBaseURL: viewBaseUrl,
			Placement:   placement,
			PreloadURL:  preloadUrl,
			NavigationNodes: []*model.NavigationNode{
				expectedNavigationNode,
			},
		}

		result, err := converter.ToGQL(&item)

		assert.Nil(t, err)
		assert.Equal(t, &expected, result)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := newClusterMicroFrontendConverter()
		item, err := converter.ToGQL(&uiv1alpha1.ClusterMicroFrontend{})

		assert.Nil(t, err)
		assert.Empty(t, item)
	})

	t.Run("Nil", func(t *testing.T) {
		converter := newClusterMicroFrontendConverter()
		item, err := converter.ToGQL(nil)

		assert.Nil(t, err)
		assert.Nil(t, item)
	})
}

func TestClusterMicroFrontendConverter_ToGQLs(t *testing.T) {
	name := "test-name"
	version := "v1"
	category := "test-category"
	viewBaseUrl := "http://test-viewBaseUrl.com"
	placement := "cluster"
	navigationNode := fixNavigationNode(t)
	item := uiv1alpha1.ClusterMicroFrontend{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: uiv1alpha1.ClusterMicroFrontendSpec{
			Placement: placement,
			CommonMicroFrontendSpec: uiv1alpha1.CommonMicroFrontendSpec{
				Version:     version,
				Category:    category,
				ViewBaseURL: viewBaseUrl,
				NavigationNodes: []uiv1alpha1.NavigationNode{
					navigationNode,
				},
			},
		},
	}

	expectedNavigationNode := fixGqlNavigationNode()
	expected := &model.ClusterMicroFrontend{
		Name:        name,
		Version:     version,
		Category:    category,
		ViewBaseURL: viewBaseUrl,
		Placement:   placement,
		NavigationNodes: []*model.NavigationNode{
			expectedNavigationNode,
		},
	}

	t.Run("Success", func(t *testing.T) {
		clusterMicroFrontends := []*uiv1alpha1.ClusterMicroFrontend{
			&item,
			&item,
		}

		converter := newClusterMicroFrontendConverter()
		result, err := converter.ToGQLs(clusterMicroFrontends)

		assert.Nil(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, expected, result[0])
	})

	t.Run("Empty", func(t *testing.T) {
		var clusterMicroFrontends []*uiv1alpha1.ClusterMicroFrontend

		converter := newClusterMicroFrontendConverter()
		result, err := converter.ToGQLs(clusterMicroFrontends)

		assert.Nil(t, err)
		assert.Empty(t, result)
	})

	t.Run("With nil", func(t *testing.T) {
		clusterMicroFrontends := []*uiv1alpha1.ClusterMicroFrontend{
			nil,
			&item,
			nil,
		}

		converter := newClusterMicroFrontendConverter()
		result, err := converter.ToGQLs(clusterMicroFrontends)

		assert.Nil(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, expected, result[0])
	})
}
