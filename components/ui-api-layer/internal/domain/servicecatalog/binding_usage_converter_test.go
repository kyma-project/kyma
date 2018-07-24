package servicecatalog

import (
	"testing"

	api "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/automock"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBindingUsageConversionToGQLCornerCases(t *testing.T) {
	t.Run("nil objs", func(t *testing.T) {
		// GIVEN
		var (
			givenK8sSBU *api.ServiceBindingUsage       = nil
			expGQLSBU   *gqlschema.ServiceBindingUsage = nil
		)
		sut := bindingUsageConverter{}

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
				Kind: gqlschema.BindingUsageReferenceTypeFunction,
			},
		}

		statusExtractorMock := automock.NewStatusBindingUsageExtractor()
		defer statusExtractorMock.AssertExpectations(t)
		statusExtractorMock.
			On("Status", mock.Anything).
			Return(gqlschema.ServiceBindingUsageStatus{})

		sut := bindingUsageConverter{statusExtractorMock}

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
					Kind: gqlschema.BindingUsageReferenceTypeDeployment,
				},
				ServiceBindingName: "redis-binding",
				Environment:        "production",
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
					Kind: gqlschema.BindingUsageReferenceTypeDeployment,
				},
				ServiceBindingName: "redis-binding",
				Environment:        "production",
				Parameters: &gqlschema.ServiceBindingUsageParameters{
					EnvPrefix: &gqlschema.EnvPrefix{Name: "ENV_PREFIX"},
				},
			},
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// GIVEN
			statusExtractorMock := automock.NewStatusBindingUsageExtractor()
			defer statusExtractorMock.AssertExpectations(t)
			statusExtractorMock.
				On("Status", tc.givenK8sSBU.Status.Conditions).
				Return(gqlschema.ServiceBindingUsageStatus{})

			sut := bindingUsageConverter{statusExtractorMock}

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
			sut := newBindingUsageConverter()
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
					Kind: gqlschema.BindingUsageReferenceTypeFunction,
				},
			},
			expK8sSBU: &api.ServiceBindingUsage{
				TypeMeta: v1.TypeMeta{
					Kind:       "ServiceBindingUsage",
					APIVersion: "servicecatalog.kyma.cx/v1alpha1",
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
				Name:        "usage",
				Environment: "production",
				ServiceBindingRef: gqlschema.ServiceBindingRefInput{
					Name: "redis-binding",
				},
				UsedBy: gqlschema.LocalObjectReferenceInput{
					Name: "app",
					Kind: gqlschema.BindingUsageReferenceTypeDeployment,
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
					APIVersion: "servicecatalog.kyma.cx/v1alpha1",
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
				Name:        "usage",
				Environment: "production",
				ServiceBindingRef: gqlschema.ServiceBindingRefInput{
					Name: "redis-binding",
				},
				UsedBy: gqlschema.LocalObjectReferenceInput{
					Name: "app",
					Kind: gqlschema.BindingUsageReferenceTypeDeployment,
				},
			},
			expK8sSBU: &api.ServiceBindingUsage{
				ObjectMeta: v1.ObjectMeta{
					Name: "usage",
				},
				TypeMeta: v1.TypeMeta{
					Kind:       "ServiceBindingUsage",
					APIVersion: "servicecatalog.kyma.cx/v1alpha1",
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
			sut := bindingUsageConverter{}
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
