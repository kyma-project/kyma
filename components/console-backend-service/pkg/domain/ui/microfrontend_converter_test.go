package ui

import (
	"encoding/json"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"

	"github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/graph/model"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var microFrontendTypeMeta = metav1.TypeMeta{
	Kind:       v1alpha1.SchemeGroupVersion.String(),
	APIVersion: "microfrontends",
}

func TestMicroFrontendConverter_ToGQL(t *testing.T) {
	t.Run("All properties are given", func(t *testing.T) {
		converter := NewMicroFrontendConverter()
		name := "test-name"
		namespace := "test-namespace"
		version := "v1"
		category := "test-category"
		viewBaseUrl := "http://test-viewBaseUrl.com"

		navigationNode := fixNavigationNode(t)
		item := v1alpha1.MicroFrontend{
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
						navigationNode,
					},
				},
			},
		}

		expectedNavigationNode := fixGqlNavigationNode()
		expected := &model.MicroFrontend{
			Name:        name,
			Version:     version,
			Category:    category,
			ViewBaseURL: viewBaseUrl,
			NavigationNodes: []*model.NavigationNode{
				expectedNavigationNode,
			},
		}

		result, err := converter.ToGQL(&item)

		assert.Nil(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := NewMicroFrontendConverter()
		item, err := converter.ToGQL(&v1alpha1.MicroFrontend{})

		assert.Nil(t, err)
		assert.Empty(t, item)
	})

	t.Run("Nil", func(t *testing.T) {
		converter := NewMicroFrontendConverter()
		item, err := converter.ToGQL(nil)

		assert.Nil(t, err)
		assert.Nil(t, item)
	})
}

func TestMicroFrontendConverter_ToGQLs(t *testing.T) {
	name := "test-name"
	namespace := "test-namespace"
	version := "v1"
	category := "test-category"
	viewBaseUrl := "http://test-viewBaseUrl.com"
	navigationNode := fixNavigationNode(t)

	item := v1alpha1.MicroFrontend{
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
					navigationNode,
				},
			},
		},
	}

	expectedNavigationNode := fixGqlNavigationNode()
	expected := &model.MicroFrontend{
		Name:        name,
		Version:     version,
		Category:    category,
		ViewBaseURL: viewBaseUrl,
		NavigationNodes: []*model.NavigationNode{
			expectedNavigationNode,
		},
	}

	t.Run("Success", func(t *testing.T) {
		microFrontends := []*v1alpha1.MicroFrontend{
			&item,
			&item,
		}

		converter := NewMicroFrontendConverter()
		result, err := converter.ToGQLs(microFrontends)

		assert.Nil(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, expected, result[0])
	})

	t.Run("Empty", func(t *testing.T) {
		var microFrontends []*v1alpha1.MicroFrontend

		converter := NewMicroFrontendConverter()
		result, err := converter.ToGQLs(microFrontends)

		assert.Nil(t, err)
		assert.Empty(t, result)
	})

	t.Run("With nil", func(t *testing.T) {
		microFrontends := []*v1alpha1.MicroFrontend{
			nil,
			&item,
			nil,
		}

		converter := NewMicroFrontendConverter()
		result, err := converter.ToGQLs(microFrontends)

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
			{
				Verbs:    []string{"foo", "bar"},
				Resource: "resource",
				APIGroup: "apigroup",
			},
		},
	}
}

func fixGqlNavigationNode() *model.NavigationNode {
	externalLinkValue := "link"
	return &model.NavigationNode{
		Label:            "test-mf",
		NavigationPath:   "test-path",
		ViewURL:          "/test/viewUrl",
		ShowInNavigation: false,
		Order:            2,
		Settings: map[string]interface{}{
			"readOnly": true,
		},
		ExternalLink: externalLinkValue,
		RequiredPermissions: []*model.RequiredPermission{
			{
				Verbs:    []string{"foo", "bar"},
				Resource: "resource",
				APIGroup: "apigroup",
			},
		},
	}
}