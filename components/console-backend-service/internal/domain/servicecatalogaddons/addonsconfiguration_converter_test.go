package servicecatalogaddons_test

import (
	"testing"

	"github.com/kyma-project/helm-broker/pkg/apis/addons/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAddonsConfigurationConverter_ToGQL(t *testing.T) {
	converter := servicecatalogaddons.NewAddonsConfigurationConverter()
	url := "ww.fix.k"
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
							{
								URL: url,
								SecretRef: &v1.SecretReference{
									Namespace: "test",
									Name:      "test",
								},
							},
						},
					},
				},
				Status: v1alpha1.AddonsConfigurationStatus{
					CommonAddonsConfigurationStatus: v1alpha1.CommonAddonsConfigurationStatus{
						Phase: v1alpha1.AddonsConfigurationReady,
						Repositories: []v1alpha1.StatusRepository{
							{
								Status:  v1alpha1.RepositoryStatus("Failed"),
								Message: "fix",
								Reason:  v1alpha1.RepositoryStatusReason("reason"),
								URL:     "rul",
								Addons: []v1alpha1.Addon{
									{
										Status:  v1alpha1.AddonStatusFailed,
										Message: "test",
										Name:    "addon",
									},
								},
							},
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
				Urls: []string{url},
				Repositories: []gqlschema.AddonsConfigurationRepository{
					{
						URL: url,
						SecretRef: &gqlschema.ResourceRef{
							Namespace: "test",
							Name:      "test",
						},
					},
				},
				Status: gqlschema.AddonsConfigurationStatus{
					Phase: string(v1alpha1.AddonsConfigurationReady),
					Repositories: []gqlschema.AddonsConfigurationStatusRepository{
						{
							Status:  "Failed",
							URL:     "rul",
							Reason:  "reason",
							Message: "fix",
							Addons: []gqlschema.AddonsConfigurationStatusAddons{
								{
									Status:  "Failed",
									Message: "test",
									Name:    "addon",
								},
							},
						},
					},
				},
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			assert.Equal(t, tc.expectedAddonsConfig, converter.ToGQL(tc.givenAddon))
		})
	}
}
