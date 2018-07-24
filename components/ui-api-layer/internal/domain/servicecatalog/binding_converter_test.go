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
			fixBinding(),
			fixBinding(),
		}

		expected := []gqlschema.ServiceBinding{
			{
				Name:                "service-binding",
				Environment:         "production",
				ServiceInstanceName: "instance",
				SecretName:          "secret-name",
				Status: gqlschema.ServiceBindingStatus{
					Type: gqlschema.ServiceBindingStatusTypePending,
				},
			},
			{
				Name:                "service-binding",
				Environment:         "production",
				ServiceInstanceName: "instance",
				SecretName:          "secret-name",
				Status: gqlschema.ServiceBindingStatus{
					Type: gqlschema.ServiceBindingStatusTypePending,
				},
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

		assert.Empty(t, result)
	})

	t.Run("With nil", func(t *testing.T) {
		bindings := []*api.ServiceBinding{
			nil,
			fixBinding(),
			nil,
		}

		expected := []gqlschema.ServiceBinding{
			{
				Name:                "service-binding",
				Environment:         "production",
				ServiceInstanceName: "instance",
				SecretName:          "secret-name",
				Status: gqlschema.ServiceBindingStatus{
					Type: gqlschema.ServiceBindingStatusTypePending,
				},
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
	actual := sut.ToGQL(fixBinding())
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
	actual := sut.ToCreateOutputGQL(fixBinding())
	// THEN
	assert.Equal(t, "service-binding", actual.Name)
	assert.Equal(t, "production", actual.Environment)
	assert.Equal(t, "instance", actual.ServiceInstanceName)
}

func fixBinding() *api.ServiceBinding {
	return &api.ServiceBinding{
		ObjectMeta: v1.ObjectMeta{
			Name:      "service-binding",
			Namespace: "production",
		},
		Spec: api.ServiceBindingSpec{
			ServiceInstanceRef: api.LocalObjectReference{Name: "instance"},
			SecretName:         "secret-name",
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
