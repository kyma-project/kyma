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

func TestClusterMicrofrontendConverter_ToGQL(t *testing.T) {
	t.Run("All properties are given", func(t *testing.T) {
		converter := clusterMicrofrontendConverter{}
		name := "test-name"
		version := "v1"
		category := "test-category"
		viewBaseUrl := "http://test-viewBaseUrl.com"
		placement := "cluster"
		settings, err := fixSettings()
		assert.Nil(t, err)

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
						v1alpha1.NavigationNode{
							Label:            "test-mf",
							NavigationPath:   "test-path",
							ViewURL:          "/test/viewUrl",
							ShowInNavigation: false,
							Order:            2,
							Settings: &runtime.RawExtension{
								Raw: settings,
							},
						},
					},
				},
			},
		}

		expected := gqlschema.ClusterMicrofrontend{
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
					ShowInNavigation: false,
					Order:            2,
					Settings: gqlschema.Settings{
						"readOnly": true,
					},
				},
			},
		}

		result, err := converter.ToGQL(&item)

		assert.Nil(t, err)
		assert.Equal(t, &expected, result)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := &clusterMicrofrontendConverter{}
		item, err := converter.ToGQL(&v1alpha1.ClusterMicroFrontend{})

		assert.Nil(t, err)
		assert.Empty(t, item)
	})

	t.Run("Nil", func(t *testing.T) {
		converter := &clusterMicrofrontendConverter{}
		item, err := converter.ToGQL(nil)

		assert.Nil(t, err)
		assert.Nil(t, item)
	})
}

func TestClusterMicrofrontendConverter_ToGQLs(t *testing.T) {
	name := "test-name"
	version := "v1"
	category := "test-category"
	viewBaseUrl := "http://test-viewBaseUrl.com"
	placement := "cluster"
	settings, err := fixSettings()
	assert.Nil(t, err)

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
					v1alpha1.NavigationNode{
						Label:            "test-mf",
						NavigationPath:   "test-path",
						ViewURL:          "/test/viewUrl",
						ShowInNavigation: true,
						Settings: &runtime.RawExtension{
							Raw: settings,
						},
					},
				},
			},
		},
	}

	expected := gqlschema.ClusterMicrofrontend{
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
				Settings: gqlschema.Settings{
					"readOnly": true,
				},
			},
		},
	}

	t.Run("Success", func(t *testing.T) {
		clusterMicrofrontends := []*v1alpha1.ClusterMicroFrontend{
			&item,
			&item,
		}

		converter := clusterMicrofrontendConverter{}
		result, err := converter.ToGQLs(clusterMicrofrontends)

		assert.Nil(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, expected, result[0])
	})

	t.Run("Empty", func(t *testing.T) {
		var clusterMicrofrontends []*v1alpha1.ClusterMicroFrontend

		converter := clusterMicrofrontendConverter{}
		result, err := converter.ToGQLs(clusterMicrofrontends)

		assert.Nil(t, err)
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
