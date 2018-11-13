package servicecatalog

import (
	"fmt"
	"testing"

	api "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestServiceBindingConverter_ToGQL(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		converter := serviceBindingConverter{}
		result, err := converter.ToGQL(&api.ServiceBinding{})
		require.NoError(t, err)

		assert.Equal(t, fixEmptyServiceBindingToGQL(), result)
	})

	t.Run("Nil", func(t *testing.T) {
		converter := serviceBindingConverter{}
		result, err := converter.ToGQL(nil)
		require.NoError(t, err)

		assert.Nil(t, result)
	})
}

func TestServiceBindingConverter_ToGQLs(t *testing.T) {
	expectedParams := map[string]interface{}{
		"json": "true",
	}

	t.Run("Success", func(t *testing.T) {
		bindings := []*api.ServiceBinding{
			fixBinding(api.ServiceBindingConditionReady),
			fixBinding(api.ServiceBindingConditionFailed),
			fixBinding(api.ServiceBindingConditionType("")),
		}
		expected := gqlschema.ServiceBindings{
			Items: []gqlschema.ServiceBinding{
				{
					Name:                "service-binding",
					Environment:         "production",
					ServiceInstanceName: "instance",
					SecretName:          "secret-name",
					Status: gqlschema.ServiceBindingStatus{
						Type: gqlschema.ServiceBindingStatusTypeReady,
					},
					Parameters: expectedParams,
				},
				{
					Name:                "service-binding",
					Environment:         "production",
					ServiceInstanceName: "instance",
					SecretName:          "secret-name",
					Status: gqlschema.ServiceBindingStatus{
						Type: gqlschema.ServiceBindingStatusTypeFailed,
					},
					Parameters: expectedParams,
				},
				{
					Name:                "service-binding",
					Environment:         "production",
					ServiceInstanceName: "instance",
					SecretName:          "secret-name",
					Status: gqlschema.ServiceBindingStatus{
						Type: gqlschema.ServiceBindingStatusTypeUnknown,
					},
					Parameters: expectedParams,
				},
			},
			Stats: gqlschema.ServiceBindingsStats{
				Ready:   1,
				Failed:  1,
				Unknown: 1,
			},
		}

		converter := serviceBindingConverter{}
		result, err := converter.ToGQLs(bindings)
		require.NoError(t, err)

		assert.Equal(t, expected, result)
	})

	t.Run("Empty", func(t *testing.T) {
		var bindings []*api.ServiceBinding

		converter := serviceBindingConverter{}
		result, err := converter.ToGQLs(bindings)
		require.NoError(t, err)

		assert.Empty(t, result.Items)
	})

	t.Run("With nil", func(t *testing.T) {
		bindings := []*api.ServiceBinding{
			nil,
			fixBinding(api.ServiceBindingConditionReady),
			nil,
		}
		expected := gqlschema.ServiceBindings{
			Items: []gqlschema.ServiceBinding{
				{
					Name:                "service-binding",
					Environment:         "production",
					ServiceInstanceName: "instance",
					SecretName:          "secret-name",
					Status: gqlschema.ServiceBindingStatus{
						Type: gqlschema.ServiceBindingStatusTypeReady,
					},
					Parameters: expectedParams,
				},
			},
			Stats: gqlschema.ServiceBindingsStats{
				Ready: 1,
			},
		}

		converter := serviceBindingConverter{}
		result, err := converter.ToGQLs(bindings)
		require.NoError(t, err)

		assert.Equal(t, expected, result)
	})
}

func TestServiceBindingConverter_ToCreateOutputGQL(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		converter := serviceBindingConverter{}
		result := converter.ToCreateOutputGQL(&api.ServiceBinding{})

		assert.Empty(t, result)
	})

	t.Run("Nil", func(t *testing.T) {
		converter := serviceBindingConverter{}
		result := converter.ToCreateOutputGQL(nil)

		assert.Nil(t, result)
	})
}

func TestServiceBindingConversionToGQL(t *testing.T) {
	// GIVEN
	sut := serviceBindingConverter{}
	// WHEN
	actual, err := sut.ToGQL(fixBinding(api.ServiceBindingConditionReady))
	// THEN
	require.NoError(t, err)
	assert.Equal(t, "service-binding", actual.Name)
	assert.Equal(t, "production", actual.Environment)
	assert.Equal(t, "secret-name", actual.SecretName)
	assert.Equal(t, "instance", actual.ServiceInstanceName)
}

func TestServiceBindingConversionToCreateOutputGQL(t *testing.T) {
	// GIVEN
	sut := serviceBindingConverter{}
	// WHEN
	actual := sut.ToCreateOutputGQL(fixBinding(api.ServiceBindingConditionReady))
	// THEN
	assert.Equal(t, "service-binding", actual.Name)
	assert.Equal(t, "production", actual.Environment)
	assert.Equal(t, "instance", actual.ServiceInstanceName)
}

func TestServiceBindingConversionError(t *testing.T) {
	// GIVEN
	var (
		sut         = serviceBindingConverter{}
		errBinding  = fixErrBinding()
		expectedErr = fmt.Sprintf("while extracting parameters from service binding [name: %s][environment: %s]: while unmarshalling binding parameters: invalid character 'o' in literal null (expecting 'u')", errBinding.Name, errBinding.Namespace)
	)

	// WHEN
	_, err := sut.ToGQL(errBinding)
	// THEN
	assert.EqualError(t, err, expectedErr)
}

func fixBinding(conditionType api.ServiceBindingConditionType) *api.ServiceBinding {
	return &api.ServiceBinding{
		ObjectMeta: v1.ObjectMeta{
			Name:      "service-binding",
			Namespace: "production",
		},
		Spec: api.ServiceBindingSpec{
			ServiceInstanceRef: api.LocalObjectReference{Name: "instance"},
			SecretName:         "secret-name",
			Parameters: &runtime.RawExtension{
				Raw: []byte(`{"json":"true"}`),
			},
		},
		Status: api.ServiceBindingStatus{
			Conditions: []api.ServiceBindingCondition{
				{
					Type:   conditionType,
					Status: api.ConditionTrue,
				},
			},
		},
	}
}

func fixErrBinding() *api.ServiceBinding {
	return &api.ServiceBinding{
		ObjectMeta: v1.ObjectMeta{
			Name:      "service-binding",
			Namespace: "production",
		},
		Spec: api.ServiceBindingSpec{
			Parameters: &runtime.RawExtension{
				Raw: []byte("not json xd"),
			},
		},
	}
}

func fixEmptyServiceBindingToGQL() *gqlschema.ServiceBinding {
	return &gqlschema.ServiceBinding{
		Status: gqlschema.ServiceBindingStatus{
			Type: gqlschema.ServiceBindingStatusTypePending,
		},
	}
}
