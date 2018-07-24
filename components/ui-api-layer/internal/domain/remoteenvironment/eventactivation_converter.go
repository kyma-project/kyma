package remoteenvironment

import (
	"strings"

	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/remoteenvironment/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/storage"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
)

type eventActivationConverter struct{}

func (c *eventActivationConverter) ToGQL(in *v1alpha1.EventActivation) *gqlschema.EventActivation {
	if in == nil {
		return nil
	}

	return &gqlschema.EventActivation{
		Name:        in.Name,
		DisplayName: in.Spec.DisplayName,
		Source:      c.toGQLSource(*in),
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

func (c *eventActivationConverter) toGQLSource(in v1alpha1.EventActivation) gqlschema.EventActivationSource {
	return gqlschema.EventActivationSource{
		Environment: in.Spec.Source.Environment,
		Namespace:   in.Spec.Source.Namespace,
		Type:        in.Spec.Source.Type,
	}
}

func (c *eventActivationConverter) ToGQLEvents(in *storage.AsyncApiSpec) []gqlschema.EventActivationEvent {
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

func (c *eventActivationConverter) convertToMap(in interface{}) map[string]interface{} {
	result, ok := in.(map[string]interface{})
	if !ok {
		return map[string]interface{}{}
	}

	return result
}
