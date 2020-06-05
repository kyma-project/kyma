package application

import (
	"testing"

	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/spec"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestEventActivationConverter_ToGQL(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		converter := &eventActivationConverter{}

		result := converter.ToGQL(fixEventActivation())

		assert.Equal(t, &gqlschema.EventActivation{
			Name:        "name",
			DisplayName: "test",
			SourceID:    "picco-bello",
		}, result)
	})

	t.Run("Nil", func(t *testing.T) {
		converter := &eventActivationConverter{}

		result := converter.ToGQL(nil)

		assert.Nil(t, result)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := &eventActivationConverter{}
		converter.ToGQL(&v1alpha1.EventActivation{})
	})
}

func TestEventActivationConverter_ToGQLs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		eventActivations := []*v1alpha1.EventActivation{
			fixEventActivation(),
			fixEventActivation(),
		}

		converter := eventActivationConverter{}
		result := converter.ToGQLs(eventActivations)

		assert.Len(t, result, 2)
		assert.Equal(t, "name", result[0].Name)
	})

	t.Run("Empty", func(t *testing.T) {
		var eventActivations []*v1alpha1.EventActivation

		converter := eventActivationConverter{}
		result := converter.ToGQLs(eventActivations)

		assert.Empty(t, result)
	})

	t.Run("With nil", func(t *testing.T) {
		eventActivations := []*v1alpha1.EventActivation{
			nil,
			fixEventActivation(),
			nil,
		}

		converter := eventActivationConverter{}
		result := converter.ToGQLs(eventActivations)

		assert.Len(t, result, 1)
		assert.Equal(t, "name", result[0].Name)
	})
}

func TestEventActivationConverter_ToGQLEvents(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		converter := &eventActivationConverter{}

		result := converter.ToGQLEvents(fixAsyncApiSpec())

		assert.Len(t, result, 2)
		assert.Contains(t, result, gqlschema.EventActivationEvent{
			EventType:   "sell",
			Version:     "v1",
			Description: "desc",
			Schema: gqlschema.JSON{
				"type": "string",
			},
		})
		assert.Contains(t, result, gqlschema.EventActivationEvent{
			EventType:   "sell",
			Version:     "v2",
			Description: "desc",
			Schema: gqlschema.JSON{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type": "string",
					},
				},
			},
		})
	})

	t.Run("Nil", func(t *testing.T) {
		converter := &eventActivationConverter{}

		result := converter.ToGQL(nil)

		assert.Empty(t, result)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := &eventActivationConverter{}

		result := converter.ToGQL(&v1alpha1.EventActivation{})

		assert.Empty(t, result)
	})

	t.Run("Without topics", func(t *testing.T) {
		converter := &eventActivationConverter{}
		asyncApi := fixAsyncApiSpec()
		asyncApi.Data.Channels = map[string]interface{}{}

		result := converter.ToGQLEvents(asyncApi)

		assert.Empty(t, result)
	})

	t.Run("Channels without version", func(t *testing.T) {
		converter := &eventActivationConverter{}

		result := converter.ToGQLEvents(fixAsyncApiSpecWithoutVersion())

		assert.Len(t, result, 1)
		assert.Contains(t, result, gqlschema.EventActivationEvent{
			EventType:   "sell",
			Version:     "",
			Description: "desc",
		})
	})
}

func fixEventActivation() *v1alpha1.EventActivation {
	return &v1alpha1.EventActivation{
		Spec: v1alpha1.EventActivationSpec{
			DisplayName: "test",
			SourceID:    "picco-bello",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "name",
		},
	}
}

func fixAsyncApiSpec() *spec.AsyncAPISpec {
	return &spec.AsyncAPISpec{
		Data: spec.AsyncAPISpecData{
			AsyncAPI: "1.0.0",
			Channels: map[string]interface{}{
				"sell.v1": map[string]interface{}{
					"subscribe": map[string]interface{}{
						"message": map[string]interface{}{
							"summary": "desc",
							"payload": map[string]interface{}{
								"type": "string",
							},
						},
					},
				},
				"sell.v2": map[string]interface{}{
					"subscribe": map[string]interface{}{
						"message": map[string]interface{}{
							"summary": "desc",
							"payload": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"name": map[string]interface{}{
										"type": "string",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func fixAsyncApiSpecWithoutVersion() *spec.AsyncAPISpec {
	return &spec.AsyncAPISpec{
		Data: spec.AsyncAPISpecData{
			AsyncAPI: "2.0.0",
			Channels: map[string]interface{}{
				"sell": map[string]interface{}{
					"subscribe": map[string]interface{}{
						"message": map[string]interface{}{
							"summary": "desc",
						},
					},
				},
			},
		},
	}
}
