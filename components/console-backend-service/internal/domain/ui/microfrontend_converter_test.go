package ui

import (
	"testing"

	"github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMicroFrontendConverter_ToGQL(t *testing.T) {
	t.Run("All properties are given", func(t *testing.T) {
		converter := newMicroFrontendConverter()
		name := "test-name"
		namespace := "test-namespace"
		version := "v1"
		category := "test-category"
		viewBaseUrl := "http://test-viewBaseUrl.com"
		experimental := true

		navigationNode := fixNavigationNode(t)
		item := v1alpha1.MicroFrontend{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: v1alpha1.MicroFrontendSpec{
				CommonMicroFrontendSpec: v1alpha1.CommonMicroFrontendSpec{
					Version:      version,
					Category:     category,
					ViewBaseURL:  viewBaseUrl,
					Experimental: experimental,
					NavigationNodes: []v1alpha1.NavigationNode{
						navigationNode,
					},
				},
			},
		}

		expectedNavigationNode := fixGqlNavigationNode()
		expected := gqlschema.MicroFrontend{
			Name:         name,
			Version:      version,
			Category:     category,
			ViewBaseURL:  viewBaseUrl,
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
		converter := newMicroFrontendConverter()
		item, err := converter.ToGQL(&v1alpha1.MicroFrontend{})

		assert.Nil(t, err)
		assert.Empty(t, item)
	})

	t.Run("Nil", func(t *testing.T) {
		converter := newMicroFrontendConverter()
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
	experimental := true
	navigationNode := fixNavigationNode(t)

	item := v1alpha1.MicroFrontend{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.MicroFrontendSpec{
			CommonMicroFrontendSpec: v1alpha1.CommonMicroFrontendSpec{
				Version:      version,
				Category:     category,
				ViewBaseURL:  viewBaseUrl,
				Experimental: experimental,
				NavigationNodes: []v1alpha1.NavigationNode{
					navigationNode,
				},
			},
		},
	}

	expectedNavigationNode := fixGqlNavigationNode()
	expected := gqlschema.MicroFrontend{
		Name:         name,
		Version:      version,
		Category:     category,
		ViewBaseURL:  viewBaseUrl,
		Experimental: experimental,
		NavigationNodes: []gqlschema.NavigationNode{
			expectedNavigationNode,
		},
	}

	t.Run("Success", func(t *testing.T) {
		microFrontends := []*v1alpha1.MicroFrontend{
			&item,
			&item,
		}

		converter := newMicroFrontendConverter()
		result, err := converter.ToGQLs(microFrontends)

		assert.Nil(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, expected, result[0])
	})

	t.Run("Empty", func(t *testing.T) {
		var microFrontends []*v1alpha1.MicroFrontend

		converter := newMicroFrontendConverter()
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

		converter := newMicroFrontendConverter()
		result, err := converter.ToGQLs(microFrontends)

		assert.Nil(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, expected, result[0])
	})
}
