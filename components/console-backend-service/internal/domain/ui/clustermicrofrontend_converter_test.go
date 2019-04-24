package ui

import (
	"testing"

	"github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestClusterMicrofrontendConverter_ToGQL(t *testing.T) {
	t.Run("All properties are given", func(t *testing.T) {
		converter := clusterMicrofrontendConverter{}
		name := "test-name"
		version := "v1"
		category := "test-category"
		viewBaseUrl := "http://test-viewBaseUrl.com"
		placement := "cluster"
		navigationNodes := []v1alpha1.NavigationNode{
			v1alpha1.NavigationNode{
				Label:            "test-mf",
				NavigationPath:   "test-path",
				ViewURL:          "/test/viewUrl",
				ShowInNavigation: true,
			},
		}

		item := v1alpha1.ClusterMicroFrontend{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: v1alpha1.ClusterMicroFrontendSpec{
				Placement: placement,
				CommonMicroFrontendSpec: v1alpha1.CommonMicroFrontendSpec{
					Version:         version,
					Category:        category,
					ViewBaseURL:     viewBaseUrl,
					NavigationNodes: navigationNodes,
				},
			},
		}

		expected := gqlschema.ClusterMicrofrontend{
			Name:            name,
			Version:         version,
			Category:        category,
			ViewBaseURL:     viewBaseUrl,
			Placement:       placement,
			NavigationNodes: make([]gqlschema.NavigationNode, 0, len(navigationNodes)),
		}

		result, err := converter.ToGQL(&item)

		require.NoError(t, err)
		assert.Equal(t, &expected, result)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := &clusterMicrofrontendConverter{}
		_, err := converter.ToGQL(&v1alpha1.ClusterMicroFrontend{})
		require.NoError(t, err)
	})

	t.Run("Nil", func(t *testing.T) {
		converter := &clusterMicrofrontendConverter{}
		item, err := converter.ToGQL(nil)

		require.NoError(t, err)
		assert.Nil(t, item)
	})
}

func TestClusterMicrofrontendConverter_ToGQLs(t *testing.T) {
	name := "test-name"
	version := "v1"
	category := "test-category"
	viewBaseUrl := "http://test-viewBaseUrl.com"
	placement := "cluster"
	navigationNodes := []v1alpha1.NavigationNode{
		v1alpha1.NavigationNode{
			Label:            "test-mf",
			NavigationPath:   "test-path",
			ViewURL:          "/test/viewUrl",
			ShowInNavigation: true,
		},
	}

	item := v1alpha1.ClusterMicroFrontend{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.ClusterMicroFrontendSpec{
			Placement: placement,
			CommonMicroFrontendSpec: v1alpha1.CommonMicroFrontendSpec{
				Version:         version,
				Category:        category,
				ViewBaseURL:     viewBaseUrl,
				NavigationNodes: navigationNodes,
			},
		},
	}

	expected := gqlschema.ClusterMicrofrontend{
		Name:            name,
		Version:         version,
		Category:        category,
		ViewBaseURL:     viewBaseUrl,
		Placement:       placement,
		NavigationNodes: make([]gqlschema.NavigationNode, 0, len(navigationNodes)),
	}

	t.Run("Success", func(t *testing.T) {
		clusterMicrofrontends := []*v1alpha1.ClusterMicroFrontend{
			&item,
			&item,
		}

		converter := clusterMicrofrontendConverter{}
		result, err := converter.ToGQLs(clusterMicrofrontends)

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, expected, result[0])
	})

	t.Run("Empty", func(t *testing.T) {
		var clusterMicrofrontends []*v1alpha1.ClusterMicroFrontend

		converter := clusterMicrofrontendConverter{}
		result, err := converter.ToGQLs(clusterMicrofrontends)

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("With nil", func(t *testing.T) {
		clusterMicrofrontends := []*v1alpha1.ClusterMicroFrontend{
			nil,
			&item,
			nil,
		}

		converter := clusterMicrofrontendConverter{}
		result, err := converter.ToGQLs(clusterMicrofrontends)

		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, expected, result[0])
	})
}
