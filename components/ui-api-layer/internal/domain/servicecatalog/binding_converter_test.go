package servicecatalog

import (
	"testing"

	api "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestServiceBindingConverter_ToGQL(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		converter := serviceBindingConverter{}
		result := converter.ToGQL(&api.ServiceBinding{})

		assert.Equal(t, fixEmptyServiceBindingToGQL(), result)
	})

	t.Run("Nil", func(t *testing.T) {
		converter := serviceBindingConverter{}
		result := converter.ToGQL(nil)

		assert.Nil(t, result)
	})
}

func TestServiceBindingConverter_ToGQLs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		bindings := []*api.ServiceBinding{
			fixBinding(api.ServiceBindingConditionReady),
			fixBinding(api.ServiceBindingConditionFailed),
			fixBinding(api.ServiceBindingConditionType("")),
		}

		expected := gqlschema.ServiceBindings{
			ServiceBindings: []gqlschema.ServiceBinding{
				{
					Name:                "service-binding",
					Environment:         "production",
					ServiceInstanceName: "instance",
					SecretName:          "secret-name",
					Status: gqlschema.ServiceBindingStatus{
						Type: gqlschema.ServiceBindingStatusTypeReady,
					},
				},
				{
					Name:                "service-binding",
					Environment:         "production",
					ServiceInstanceName: "instance",
					SecretName:          "secret-name",
					Status: gqlschema.ServiceBindingStatus{
						Type: gqlschema.ServiceBindingStatusTypeFailed,
					},
				},
				{
					Name:                "service-binding",
					Environment:         "production",
					ServiceInstanceName: "instance",
					SecretName:          "secret-name",
					Status: gqlschema.ServiceBindingStatus{
						Type: gqlschema.ServiceBindingStatusTypeUnknown,
					},
				},
			},
			Stats: gqlschema.ServiceBindingsStats{
				Ready:   1,
				Failed:  1,
				Unknown: 1,
			},
		}

		converter := serviceBindingConverter{}
		result := converter.ToGQLs(bindings)

		assert.Equal(t, expected, result)
	})

	t.Run("Empty", func(t *testing.T) {
		var bindings []*api.ServiceBinding

		converter := serviceBindingConverter{}
		result := converter.ToGQLs(bindings)

		assert.Empty(t, result.ServiceBindings)
	})

	t.Run("With nil", func(t *testing.T) {
		bindings := []*api.ServiceBinding{
			nil,
			fixBinding(api.ServiceBindingConditionReady),
			nil,
		}

		expected := gqlschema.ServiceBindings{
			ServiceBindings: []gqlschema.ServiceBinding{
				{
					Name:                "service-binding",
					Environment:         "production",
					ServiceInstanceName: "instance",
					SecretName:          "secret-name",
					Status: gqlschema.ServiceBindingStatus{
						Type: gqlschema.ServiceBindingStatusTypeReady,
					},
				},
			},
			Stats: gqlschema.ServiceBindingsStats{
				Ready: 1,
			},
		}

		converter := serviceBindingConverter{}
		result := converter.ToGQLs(bindings)

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
	actual := sut.ToGQL(fixBinding(api.ServiceBindingConditionReady))
	// THEN
	assert.Equal(t, "service-binding", actual.Name)
	assert.Equal(t, "production", actual.Environment)
	assert.Equal(t, "secret-name", actual.SecretName)
	assert.Equal(t, "instance", actual.ServiceInstanceName)
}

func TestServicebindingConversionToCreateOutputGQL(t *testing.T) {
	// GIVEN
	sut := serviceBindingConverter{}
	// WHEN
	actual := sut.ToCreateOutputGQL(fixBinding(api.ServiceBindingConditionReady))
	// THEN
	assert.Equal(t, "service-binding", actual.Name)
	assert.Equal(t, "production", actual.Environment)
	assert.Equal(t, "instance", actual.ServiceInstanceName)
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

func fixEmptyServiceBindingToGQL() *gqlschema.ServiceBinding {
	return &gqlschema.ServiceBinding{
		Status: gqlschema.ServiceBindingStatus{
			Type: gqlschema.ServiceBindingStatusTypePending,
		},
	}
}
