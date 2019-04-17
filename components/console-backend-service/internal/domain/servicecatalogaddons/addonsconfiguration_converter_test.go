package servicecatalogaddons_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons"
)

func TestAddonsConfigurationConverter_ToGQL(t *testing.T) {
	converter := servicecatalogaddons.NewAddonsConfigurationConverter()

	for tn, tc := range map[string]struct {
		givenConfigMap       *v1.ConfigMap
		expectedAddonsConfig *gqlschema.AddonsConfiguration
	}{
		"empty": {
			givenConfigMap:       &v1.ConfigMap{},
			expectedAddonsConfig: &gqlschema.AddonsConfiguration{},
		},
		"full": {
			givenConfigMap: &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Labels: map[string]string{
						"test": "test",
					},
				},
				Data: map[string]string{
					"URLs": "www.example.com",
				},
			},
			expectedAddonsConfig: &gqlschema.AddonsConfiguration{
				Name: "test",
				Labels: gqlschema.Labels{
					"test": "test",
				},
				Urls: []string{"www.example.com"},
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			assert.Equal(t, tc.expectedAddonsConfig, converter.ToGQL(tc.givenConfigMap))
		})
	}
}

func TestAddonsConfigurationConverter_ToGQLs(t *testing.T) {
	converter := servicecatalogaddons.NewAddonsConfigurationConverter()

	for tn, tc := range map[string]struct {
		givenConfigMaps      []*v1.ConfigMap
		expectedAddonsConfig []gqlschema.AddonsConfiguration
	}{
		"empty": {
			givenConfigMaps:      []*v1.ConfigMap{},
			expectedAddonsConfig: []gqlschema.AddonsConfiguration(nil),
		},
		"full": {
			givenConfigMaps: []*v1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
						Labels: map[string]string{
							"test": "test",
						},
					},
					Data: map[string]string{
						"URLs": "www.example.com",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test2",
						Labels: map[string]string{
							"test2": "test2",
						},
					},
					Data: map[string]string{
						"URLs": "www.next.com",
					},
				},
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
			assert.Equal(t, tc.expectedAddonsConfig, converter.ToGQLs(tc.givenConfigMaps))
		})
	}
}
