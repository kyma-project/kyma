package remoteenvironment

import (
	"testing"

	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/remoteenvironment/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/storage"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestEventActivationConverter_ToGQL(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		converter := &eventActivationConverter{}

		result := converter.ToGQL(fixEventActivation())

		assert.Equal(t, &gqlschema.EventActivation{
			Name:        "name",
			DisplayName: "test",
			Source: gqlschema.EventActivationSource{
				Type:        "nope",
				Namespace:   "nms",
				Environment: "env",
			},
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
		})
		assert.Contains(t, result, gqlschema.EventActivationEvent{
			EventType:   "sell",
			Version:     "v2",
			Description: "desc",
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
		asyncApi.Data.Topics = map[string]interface{}{}

		result := converter.ToGQLEvents(asyncApi)

		assert.Empty(t, result)
	})

	t.Run("Topics without version", func(t *testing.T) {
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
			Source: v1alpha1.Source{
				Environment: "env",
				Namespace:   "nms",
				Type:        "nope",
			},
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "name",
		},
	}
}

func fixAsyncApiSpec() *storage.AsyncApiSpec {
	return &storage.AsyncApiSpec{
		Data: storage.AsyncApiSpecData{
			AsyncAPI: "1.0.0",
			Topics: map[string]interface{}{
				"sell.v1": map[string]interface{}{
					"subscribe": map[string]interface{}{
						"summary": "desc",
					},
				},
				"sell.v2": map[string]interface{}{
					"subscribe": map[string]interface{}{
						"summary": "desc",
					},
				},
			},
		},
	}
}

func fixAsyncApiSpecWithoutVersion() *storage.AsyncApiSpec {
	return &storage.AsyncApiSpec{
		Data: storage.AsyncApiSpecData{
			AsyncAPI: "1.0.0",
			Topics: map[string]interface{}{
				"sell": map[string]interface{}{
					"subscribe": map[string]interface{}{
						"summary": "desc",
					},
				},
			},
		},
	}
}
