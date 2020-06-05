package application

import (
	"strings"

	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/spec"
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
	for k, channel := range in.Data.Channels {
		if !c.isSubscribeEvent(channel) {
			continue
		}

		eventType, version := c.getEventVersionedType(k)
		events = append(events, gqlschema.EventActivationEvent{
			EventType:   eventType,
			Version:     version,
			Description: c.getSummary(channel),
			Schema:      c.getPayload(channel),
		})
	}

	return events
}

func (c *eventActivationConverter) getEventVersionedType(in string) (string, string) {
	versionedType := strings.Replace(in, "/", ".", -1)
	lastDotIndex := strings.LastIndex(versionedType, ".")
	if lastDotIndex < 0 {
		return versionedType, ""
	}

	eventType := versionedType[:lastDotIndex]
	version := versionedType[(lastDotIndex + 1):]

	return eventType, version
}

func (c *eventActivationConverter) isSubscribeEvent(in interface{}) bool {
	_, exists := c.convertToMap(in)["subscribe"]
	return exists
}

func (c *eventActivationConverter) getSummary(in interface{}) string {
	message := c.getMessage(in)
	if message == nil {
		return ""
	}

	summary, exists := message["summary"]
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
	message := c.getMessage(in)
	if message == nil {
		return nil
	}

	payload, exists := message["payload"]
	if !exists {
		return nil
	}

	return c.convertToMap(payload)
}

func (c *eventActivationConverter) getMessage(in interface{}) map[string]interface{} {
	subscribe, exists := c.convertToMap(in)["subscribe"]
	if !exists {
		return map[string]interface{}{}
	}

	message, exists := c.convertToMap(subscribe)["message"]
	if !exists {
		return map[string]interface{}{}
	}
	return c.convertToMap(message)
}

func (c *eventActivationConverter) convertToMap(in interface{}) map[string]interface{} {
	result, ok := in.(map[string]interface{})
	if !ok {
		return map[string]interface{}{}
	}

	return result
}
