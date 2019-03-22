package trace

import api "github.com/kyma-project/kyma/components/event-bus/api/publish"

const (
	eventID                 = "event-id"
	sourceID                = "source-id"
	eventType               = "event-type"
	eventTypeVersion        = "event-type-ver"
	SubscriptionName        = "sub-name"
	SubscriptionEnvironment = "sub-env"
)

func CreateTraceTagsFromCloudEvent(cloudEvent *api.CloudEvent) map[string]string {
	return map[string]string{
		eventID:          cloudEvent.EventID,
		sourceID:         cloudEvent.SourceID,
		eventType:        cloudEvent.EventType,
		eventTypeVersion: cloudEvent.EventTypeVersion,
	}
}

func CreateTraceTagsFromMessageHeader(headers map[string]string) map[string]string {
	return map[string]string{
		eventID:          headers[eventID],
		sourceID:         headers[sourceID],
		eventType:        headers[eventType],
		eventTypeVersion: headers[eventTypeVersion],
	}
}
