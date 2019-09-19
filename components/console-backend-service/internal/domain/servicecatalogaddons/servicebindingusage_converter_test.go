package servicecatalogaddons_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	api "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBindingUsageConversionToGQLCornerCases(t *testing.T) {
	t.Run("nil objs", func(t *testing.T) {
		// GIVEN
		var (
			givenK8sSBU *api.ServiceBindingUsage       = nil
			expGQLSBU   *gqlschema.ServiceBindingUsage = nil
		)
		sut := servicecatalogaddons.NewServiceBindingUsageConverter()

		// WHEN
		gotGQLSBU, err := sut.ToGQL(givenK8sSBU)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, expGQLSBU, gotGQLSBU)
	})

	t.Run("only kind provided", func(t *testing.T) {
		// GIVEN
		givenK8sSBU := &api.ServiceBindingUsage{
			Spec: api.ServiceBindingUsageSpec{
				UsedBy: api.LocalReferenceByKindAndName{
					Kind: "Function",
				},
			},
		}
		expGQLSBU := &gqlschema.ServiceBindingUsage{
			UsedBy: gqlschema.LocalObjectReference{
				Kind: "Function",
			},
			Status: gqlschema.ServiceBindingUsageStatus{
				Type: gqlschema.ServiceBindingUsageStatusTypePending,
			},
		}
		sut := servicecatalogaddons.NewServiceBindingUsageConverter()

		// WHEN
		gotGQLSBU, err := sut.ToGQL(givenK8sSBU)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, expGQLSBU, gotGQLSBU)
	})
}

func TestBindingUsageConversionToGQL(t *testing.T) {
	tests := map[string]struct {
		givenK8sSBU *api.ServiceBindingUsage
		expGQLSBU   *gqlschema.ServiceBindingUsage
	}{
		"without env prefix": {
			givenK8sSBU: fixRedisUsage(),
			expGQLSBU: &gqlschema.ServiceBindingUsage{
				Name: "usage",
				UsedBy: gqlschema.LocalObjectReference{
					Name: "app",
					Kind: "Deployment",
				},
				ServiceBindingName: "redis-binding",
				Namespace:          "production",
				Status: gqlschema.ServiceBindingUsageStatus{
					Type: gqlschema.ServiceBindingUsageStatusTypePending,
				},
			},
		},
		"with env prefix": {
			givenK8sSBU: func() *api.ServiceBindingUsage {
				fix := fixRedisUsage()
				fix.Spec.Parameters = &api.Parameters{
					EnvPrefix: &api.EnvPrefix{Name: "ENV_PREFIX"},
				}
				return fix
			}(),
			expGQLSBU: &gqlschema.ServiceBindingUsage{
				Name: "usage",
				UsedBy: gqlschema.LocalObjectReference{
					Name: "app",
					Kind: "Deployment",
				},
				ServiceBindingName: "redis-binding",
				Namespace:          "production",
				Parameters: &gqlschema.ServiceBindingUsageParameters{
					EnvPrefix: &gqlschema.EnvPrefix{Name: "ENV_PREFIX"},
				},
				Status: gqlschema.ServiceBindingUsageStatus{
					Type: gqlschema.ServiceBindingUsageStatusTypePending,
				},
			},
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// GIVEN
			sut := servicecatalogaddons.NewServiceBindingUsageConverter()

			// WHEN
			gotGQLSBU, err := sut.ToGQL(tc.givenK8sSBU)

			// THEN
			require.NoError(t, err)
			assert.Equal(t, tc.expGQLSBU, gotGQLSBU)
		})
	}
}

func TestBindingUsageConversionToGQLs(t *testing.T) {
	tests := map[string]struct {
		givenK8sSBUs []*api.ServiceBindingUsage
	}{
		"with one entry": {
			givenK8sSBUs: []*api.ServiceBindingUsage{
				fixRedisUsage(),
			},
		},
		"with nil": {
			givenK8sSBUs: []*api.ServiceBindingUsage{
				nil,
				fixRedisUsage(),
				nil,
			},
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// GIVEN
			sut := servicecatalogaddons.NewServiceBindingUsageConverter()
			// WHEN
			actual, err := sut.ToGQLs(tc.givenK8sSBUs)
			// THEN
			require.NoError(t, err)
			assert.Len(t, actual, 1)
			assert.Equal(t, "usage", actual[0].Name)
		})
	}
}

func TestBindingUsageConversionInput(t *testing.T) {
	tests := map[string]struct {
		givenSBUInput *gqlschema.CreateServiceBindingUsageInput
		expK8sSBU     *api.ServiceBindingUsage
	}{
		"only kind is provided": {
			givenSBUInput: &gqlschema.CreateServiceBindingUsageInput{
				UsedBy: gqlschema.LocalObjectReferenceInput{
					Kind: "Function",
				},
			},
			expK8sSBU: &api.ServiceBindingUsage{
				TypeMeta: v1.TypeMeta{
					Kind:       "ServiceBindingUsage",
					APIVersion: "servicecatalog.kyma-project.io/v1alpha1",
				},
				Spec: api.ServiceBindingUsageSpec{
					UsedBy: api.LocalReferenceByKindAndName{
						Kind: "Function",
					},
				},
			},
		},
		"nil": {
			givenSBUInput: nil,
			expK8sSBU:     nil,
		},
		"with env prefix": {
			givenSBUInput: &gqlschema.CreateServiceBindingUsageInput{
				Name: ptr("usage"),
				ServiceBindingRef: gqlschema.ServiceBindingRefInput{
					Name: "redis-binding",
				},
				UsedBy: gqlschema.LocalObjectReferenceInput{
					Name: "app",
					Kind: "Deployment",
				},
				Parameters: &gqlschema.ServiceBindingUsageParametersInput{
					EnvPrefix: &gqlschema.EnvPrefixInput{Name: "ENV_PREFIX"},
				},
			},
			expK8sSBU: &api.ServiceBindingUsage{
				ObjectMeta: v1.ObjectMeta{
					Name: "usage",
				},
				TypeMeta: v1.TypeMeta{
					Kind:       "ServiceBindingUsage",
					APIVersion: "servicecatalog.kyma-project.io/v1alpha1",
				},
				Spec: api.ServiceBindingUsageSpec{
					ServiceBindingRef: api.LocalReferenceByName{
						Name: "redis-binding",
					},
					UsedBy: api.LocalReferenceByKindAndName{
						Name: "app",
						Kind: "Deployment",
					},
					Parameters: &api.Parameters{
						EnvPrefix: &api.EnvPrefix{Name: "ENV_PREFIX"},
					},
				},
			},
		},
		"without env prefix": {
			givenSBUInput: &gqlschema.CreateServiceBindingUsageInput{
				Name: ptr("usage"),
				ServiceBindingRef: gqlschema.ServiceBindingRefInput{
					Name: "redis-binding",
				},
				UsedBy: gqlschema.LocalObjectReferenceInput{
					Name: "app",
					Kind: "Deployment",
				},
			},
			expK8sSBU: &api.ServiceBindingUsage{
				ObjectMeta: v1.ObjectMeta{
					Name: "usage",
				},
				TypeMeta: v1.TypeMeta{
					Kind:       "ServiceBindingUsage",
					APIVersion: "servicecatalog.kyma-project.io/v1alpha1",
				},
				Spec: api.ServiceBindingUsageSpec{
					ServiceBindingRef: api.LocalReferenceByName{
						Name: "redis-binding",
					},
					UsedBy: api.LocalReferenceByKindAndName{
						Name: "app",
						Kind: "Deployment",
					},
				},
			},
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// GIVEN
			sut := servicecatalogaddons.NewServiceBindingUsageConverter()
			// WHEN
			gotK8sSBU, err := sut.InputToK8s(tc.givenSBUInput)
			// THEN
			require.NoError(t, err)
			assert.Equal(t, tc.expK8sSBU, gotK8sSBU)
		})
	}
}

func fixRedisUsage() *api.ServiceBindingUsage {
	return &api.ServiceBindingUsage{
		TypeMeta: v1.TypeMeta{
			Kind:       "ServiceBindingUsage",
			APIVersion: api.SchemeGroupVersion.String(),
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "usage",
			Namespace: "production",
		},
		Spec: api.ServiceBindingUsageSpec{
			ServiceBindingRef: api.LocalReferenceByName{
				Name: "redis-binding",
			},
			UsedBy: api.LocalReferenceByKindAndName{
				Name: "app",
				Kind: "Deployment",
			},
		},
	}
}

func ptr(s string) *string {
	return &s
}
