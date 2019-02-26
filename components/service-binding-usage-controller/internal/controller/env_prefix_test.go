package controller

import (
	"testing"

	sbuTypes "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/stretchr/testify/assert"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestEnvPrefixGetterGetPrefix(t *testing.T) {
	tests := map[string]struct {
		givenSBU       sbuTypes.ServiceBindingUsage
		expectedPrefix string
	}{
		"Parameters on ServiceBindingUsage not provided": {
			givenSBU: sbuTypes.ServiceBindingUsage{
				ObjectMeta: metaV1.ObjectMeta{
					Name: "fix-sbu-name",
				},
				Spec: sbuTypes.ServiceBindingUsageSpec{},
			},
			expectedPrefix: "",
		},
		"Empty Parameters on ServiceBindingUsage": {
			givenSBU: sbuTypes.ServiceBindingUsage{
				ObjectMeta: metaV1.ObjectMeta{
					Name: "fix-sbu-name",
				},
				Spec: sbuTypes.ServiceBindingUsageSpec{
					Parameters: &sbuTypes.Parameters{},
				},
			},
			expectedPrefix: "",
		},
		"Empty EnvPrefix passed in Parameters on ServiceBindingUsage": {
			givenSBU: sbuTypes.ServiceBindingUsage{
				ObjectMeta: metaV1.ObjectMeta{
					Name: "aol-1",
				},
				Spec: sbuTypes.ServiceBindingUsageSpec{
					Parameters: &sbuTypes.Parameters{
						EnvPrefix: &sbuTypes.EnvPrefix{},
					},
				}},
			expectedPrefix: "",
		},
		"Not empty EnvPrefix passed in Parameters on ServiceBindingUsage": {
			givenSBU: sbuTypes.ServiceBindingUsage{
				ObjectMeta: metaV1.ObjectMeta{
					Name: "aol-1",
				},
				Spec: sbuTypes.ServiceBindingUsageSpec{
					Parameters: &sbuTypes.Parameters{
						EnvPrefix: &sbuTypes.EnvPrefix{
							Name: "TEST_PREFIX_",
						},
					},
				}},
			expectedPrefix: "TEST_PREFIX_",
		},
	}

	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			prefixGetter := envPrefixGetter{}

			// when
			prefix := prefixGetter.GetPrefix(&tc.givenSBU)

			// then
			assert.Equal(t, tc.expectedPrefix, prefix)
		})
	}
}
