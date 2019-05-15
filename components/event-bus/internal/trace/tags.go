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

const (
	// push request headers to endpoint
	HeaderSourceID         = "Ce-Source-ID"
	HeaderEventType        = "Ce-Event-Type"
	HeaderEventTypeVersion = "Ce-Event-Type-Version"
	HeaderEventID          = "Ce-Event-ID"
	HeaderEventTime        = "Ce-Event-Time"
)

func CreateTraceTagsFromCloudEvent(cloudEvent *api.CloudEvent) map[string]string {
	return map[string]string{
		eventID:          cloudEvent.EventID,
		sourceID:         cloudEvent.SourceID,
		eventType:        cloudEvent.EventType,
		eventTypeVersion: cloudEvent.EventTypeVersion,
	}
}

func CreateTraceTagsFromMessageHeader(headers map[string][]string) map[string]string {
	return map[string]string{
		eventID:          headers[HeaderEventID][0],
		sourceID:         headers[HeaderSourceID][0],
		eventType:        headers[HeaderEventType][0],
		eventTypeVersion: headers[HeaderEventTypeVersion][0],
	}
}
