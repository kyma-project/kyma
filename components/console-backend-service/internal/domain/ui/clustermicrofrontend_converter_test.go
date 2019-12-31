package ui

import (
	"encoding/json"
	"testing"

	"github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
		experimental := false
		navigationNode := fixNavigationNode(t)

		item := v1alpha1.ClusterMicroFrontend{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: v1alpha1.ClusterMicroFrontendSpec{
				Placement: placement,
				CommonMicroFrontendSpec: v1alpha1.CommonMicroFrontendSpec{
					Version:     version,
					Category:    category,
					ViewBaseURL: viewBaseUrl,
					PreloadURL:  preloadUrl,
					NavigationNodes: []v1alpha1.NavigationNode{
						navigationNode,
					},
				},
			},
		}

		expectedNavigationNode := fixGqlNavigationNode()

		expected := gqlschema.ClusterMicroFrontend{
			Name:         name,
			Version:      version,
			Category:     category,
			ViewBaseURL:  viewBaseUrl,
			Placement:    placement,
			PreloadURL:   preloadUrl,
			Experimental: experimental,
			NavigationNodes: []gqlschema.NavigationNode{
				expectedNavigationNode,
			},
		}

		result, err := converter.ToGQL(&item)

		assert.Nil(t, err)
		assert.Equal(t, &expected, result)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := newClusterMicroFrontendConverter()
		item, err := converter.ToGQL(&v1alpha1.ClusterMicroFrontend{})

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
	experimental := false
	navigationNode := fixNavigationNode(t)
	item := v1alpha1.ClusterMicroFrontend{
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
					navigationNode,
				},
			},
		},
	}

	expectedNavigationNode := fixGqlNavigationNode()
	expected := gqlschema.ClusterMicroFrontend{
		Name:         name,
		Version:      version,
		Category:     category,
		ViewBaseURL:  viewBaseUrl,
		Placement:    placement,
		Experimental: experimental,
		NavigationNodes: []gqlschema.NavigationNode{
			expectedNavigationNode,
		},
	}

	t.Run("Success", func(t *testing.T) {
		clusterMicroFrontends := []*v1alpha1.ClusterMicroFrontend{
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
		var clusterMicroFrontends []*v1alpha1.ClusterMicroFrontend

		converter := newClusterMicroFrontendConverter()
		result, err := converter.ToGQLs(clusterMicroFrontends)

		assert.Nil(t, err)
		assert.Empty(t, result)
	})

	t.Run("With nil", func(t *testing.T) {
		clusterMicroFrontends := []*v1alpha1.ClusterMicroFrontend{
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

func fixSettings() ([]byte, error) {
	settings := map[string]interface{}{
		"readOnly": true,
	}
	return json.Marshal(settings)
}

func fixNavigationNode(t *testing.T) v1alpha1.NavigationNode {
	settings, err := fixSettings()
	assert.Nil(t, err)
	return v1alpha1.NavigationNode{
		Label:            "test-mf",
		NavigationPath:   "test-path",
		ViewURL:          "/test/viewUrl",
		ShowInNavigation: false,
		Order:            2,
		Settings: &runtime.RawExtension{
			Raw: settings,
		},
		ExternalLink: "link",
		RequiredPermissions: []v1alpha1.RequiredPermission{
			v1alpha1.RequiredPermission{
				Verbs:    []string{"foo", "bar"},
				Resource: "resource",
				APIGroup: "apigroup",
			},
		},
	}
}

func fixGqlNavigationNode() gqlschema.NavigationNode {
	externalLinkValue := "link"
	return gqlschema.NavigationNode{
		Label:            "test-mf",
		NavigationPath:   "test-path",
		ViewURL:          "/test/viewUrl",
		ShowInNavigation: false,
		Order:            2,
		Settings: gqlschema.Settings{
			"readOnly": true,
		},
		ExternalLink: &externalLinkValue,
		RequiredPermissions: []gqlschema.RequiredPermission{
			gqlschema.RequiredPermission{
				Verbs:    []string{"foo", "bar"},
				Resource: "resource",
				APIGroup: "apigroup",
			},
		},
	}
}
