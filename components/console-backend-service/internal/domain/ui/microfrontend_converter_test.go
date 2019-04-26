package ui

import (
	"testing"

	"github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMicrofrontendConverter_ToGQL(t *testing.T) {
	t.Run("All properties are given", func(t *testing.T) {
		converter := microfrontendConverter{}
		name := "test-name"
		namespace := "test-namespace"
		version := "v1"
		category := "test-category"
		viewBaseUrl := "http://test-viewBaseUrl.com"

		item := v1alpha1.MicroFrontend{
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
							Order:            2,
							Settings: v1alpha1.Settings{
								ReadOnly: true,
							},
						},
					},
				},
			},
		}

		expected := gqlschema.Microfrontend{
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
					Order:            2,
					Settings: gqlschema.Settings{
						ReadOnly: true,
					},
				},
			},
		}

		result := converter.ToGQL(&item)

		assert.Equal(t, &expected, result)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := &microfrontendConverter{}
		item := converter.ToGQL(&v1alpha1.MicroFrontend{})

		assert.Empty(t, item)
	})

	t.Run("Nil", func(t *testing.T) {
		converter := &microfrontendConverter{}
		item := converter.ToGQL(nil)

		assert.Nil(t, item)
	})
}

func TestMicrofrontendConverter_ToGQLs(t *testing.T) {
	name := "test-name"
	namespace := "test-namespace"
	version := "v1"
	category := "test-category"
	viewBaseUrl := "http://test-viewBaseUrl.com"

	item := v1alpha1.MicroFrontend{
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
						ShowInNavigation: false,
						Order:            2,
						Settings: v1alpha1.Settings{
							ReadOnly: false,
						},
					},
				},
			},
		},
	}

	expected := gqlschema.Microfrontend{
		Name:        name,
		Version:     version,
		Category:    category,
		ViewBaseURL: viewBaseUrl,
		NavigationNodes: []gqlschema.NavigationNode{
			gqlschema.NavigationNode{
				Label:            "test-mf",
				NavigationPath:   "test-path",
				ViewURL:          "/test/viewUrl",
				ShowInNavigation: false,
				Order:            2,
				Settings: gqlschema.Settings{
					ReadOnly: false,
				},
			},
		},
	}

	t.Run("Success", func(t *testing.T) {
		microfrontends := []*v1alpha1.MicroFrontend{
			&item,
			&item,
		}

		converter := microfrontendConverter{}
		result := converter.ToGQLs(microfrontends)

		assert.Len(t, result, 2)
		assert.Equal(t, expected, result[0])
	})

	t.Run("Empty", func(t *testing.T) {
		var microfrontends []*v1alpha1.MicroFrontend

		converter := microfrontendConverter{}
		result := converter.ToGQLs(microfrontends)

		assert.Empty(t, result)
	})

	t.Run("With nil", func(t *testing.T) {
		microfrontends := []*v1alpha1.MicroFrontend{
			nil,
			&item,
			nil,
		}

		converter := microfrontendConverter{}
		result := converter.ToGQLs(microfrontends)

		assert.Len(t, result, 1)
		assert.Equal(t, expected, result[0])
	})
}
