package servicecatalogaddons

import (
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAddonsConfigurationConverter_ToGQL(t *testing.T) {
	converter := NewAddonsConfigurationConverter()
	for tn, tc := range map[string]struct {
		givenAddon           *v1alpha1.AddonsConfiguration
		expectedAddonsConfig *gqlschema.AddonsConfiguration
	}{
		"empty": {
			givenAddon:           &v1alpha1.AddonsConfiguration{},
			expectedAddonsConfig: &gqlschema.AddonsConfiguration{},
		},
		"full": {
			givenAddon: &v1alpha1.AddonsConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Labels: map[string]string{
						"add": "it",
						"ion": "al",
					},
				},
				Spec: v1alpha1.AddonsConfigurationSpec{
					CommonAddonsConfigurationSpec: v1alpha1.CommonAddonsConfigurationSpec{
						Repositories: []v1alpha1.SpecRepository{
							{URL: "ww.fix.k"},
						},
					},
				},
			},
			expectedAddonsConfig: &gqlschema.AddonsConfiguration{
				Name: "test",
				Labels: gqlschema.Labels{
					"add": "it",
					"ion": "al",
				},
				Urls: []string{"ww.fix.k"},
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			assert.Equal(t, tc.expectedAddonsConfig, converter.ToGQL(tc.givenAddon))
		})
	}
}

func TestAddonsConfigurationConverter_ToGQLs(t *testing.T) {
	converter := NewAddonsConfigurationConverter()

	for tn, tc := range map[string]struct {
		givenAddons          []*v1alpha1.AddonsConfiguration
		expectedAddonsConfig []gqlschema.AddonsConfiguration
	}{
		"empty": {
			givenAddons:          []*v1alpha1.AddonsConfiguration{},
			expectedAddonsConfig: []gqlschema.AddonsConfiguration(nil),
		},
		"full": {
			givenAddons: []*v1alpha1.AddonsConfiguration{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
						Labels: map[string]string{
							"test": "test",
						},
					},
					Spec: v1alpha1.AddonsConfigurationSpec{
						CommonAddonsConfigurationSpec: v1alpha1.CommonAddonsConfigurationSpec{
							Repositories: []v1alpha1.SpecRepository{
								{URL: "www.example.com"},
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test2",
						Labels: map[string]string{
							"test2": "test2",
						},
					},
					Spec: v1alpha1.AddonsConfigurationSpec{
						CommonAddonsConfigurationSpec: v1alpha1.CommonAddonsConfigurationSpec{
							Repositories: []v1alpha1.SpecRepository{
								{URL: "www.next.com"},
							},
						},
					}},
			},
			expectedAddonsConfig: []gqlschema.AddonsConfiguration{
				{
					Name: "test",
					Labels: gqlschema.Labels{
						"test": "test",
					},
					Urls: []string{"www.example.com"},
				},
				{
					Name: "test2",
					Labels: gqlschema.Labels{
						"test2": "test2",
					},
					Urls: []string{"www.next.com"},
				},
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			assert.Equal(t, tc.expectedAddonsConfig, converter.ToGQLs(tc.givenAddons))
		})
	}
}
