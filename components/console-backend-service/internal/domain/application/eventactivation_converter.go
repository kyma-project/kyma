package application

import (
	"strings"

	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/assetstore/spec"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

type eventActivationConverter struct{}

func (c *eventActivationConverter) ToGQL(in *v1alpha1.EventActivation) *gqlschema.EventActivation {
	if in == nil {
		return nil
	}

	return &gqlschema.EventActivation{
		Name:        in.Name,
		DisplayName: in.Spec.DisplayName,
		SourceID:    in.Spec.SourceID,
	}
}

func (c *eventActivationConverter) ToGQLs(in []*v1alpha1.EventActivation) []gqlschema.EventActivation {
	var result []gqlschema.EventActivation
	for _, item := range in {
		converted := c.ToGQL(item)
		if converted != nil {
			result = append(result, *converted)
		}
	}

	return result
}

func (c *eventActivationConverter) ToGQLEvents(in *spec.AsyncAPISpec) []gqlschema.EventActivationEvent {
	if in == nil {
		return []gqlschema.EventActivationEvent{}
	}

	var events []gqlschema.EventActivationEvent
	for k, topic := range in.Data.Topics {
		if !c.isSubscribeEvent(topic) {
			continue
		}

		eventType, version := c.getEventVersionedType(k)
		events = append(events, gqlschema.EventActivationEvent{
			EventType:   eventType,
			Version:     version,
			Description: c.getSummary(topic),
			Schema:      c.getPayload(topic),
		})
	}

	return events
}

func (c *eventActivationConverter) getEventVersionedType(in string) (string, string) {
	lastDotIndex := strings.LastIndex(in, ".")
	if lastDotIndex < 0 {
		return in, ""
	}

	eventType := in[:lastDotIndex]
	version := in[(lastDotIndex + 1):]

	return eventType, version
}

func (c *eventActivationConverter) isSubscribeEvent(in interface{}) bool {
	_, exists := c.convertToMap(in)["subscribe"]
	return exists
}

func (c *eventActivationConverter) getSummary(in interface{}) string {
	subscribe, exists := c.convertToMap(in)["subscribe"]
	if !exists {
		return ""
	}

	summary, exists := c.convertToMap(subscribe)["summary"]
	if !exists {
		return ""
	}

	result, ok := summary.(string)
	if !ok {
		return ""
	}

	return result
}

func (c *eventActivationConverter) getPayload(in interface{}) map[string]interface{} {
	subscribe, exists := c.convertToMap(in)["subscribe"]
	if !exists {
		return nil
	}

	payload, exists := c.convertToMap(subscribe)["payload"]
	if !exists {
		return nil
	}

	result, ok := payload.(map[string]interface{})
	if !ok {
		return nil
	}

	return result
}

func (c *eventActivationConverter) convertToMap(in interface{}) map[string]interface{} {
	result, ok := in.(map[string]interface{})
	if !ok {
		return map[string]interface{}{}
	}

	return result
}
