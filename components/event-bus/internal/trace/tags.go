package trace

import api "github.com/kyma-project/kyma/components/event-bus/api/publish"

const (
	eventID                 = "event-id"
	sourceNamespace         = "source-ns"
	sourceType              = "source-type"
	sourceEnvironment       = "source-env"
	eventType               = "event-type"
	eventTypeVersion        = "event-type-ver"
	SubscriptionName        = "sub-name"
	SubscriptionEnvironment = "sub-env"
)

func CreateTraceTagsFromCloudEvent(cloudEvent *api.CloudEvent) map[string]string {
	return map[string]string{
		eventID:           cloudEvent.EventID,
		sourceNamespace:   cloudEvent.Source.SourceNamespace,
		sourceType:        cloudEvent.Source.SourceType,
		sourceEnvironment: cloudEvent.Source.SourceEnvironment,
		eventType:         cloudEvent.EventType,
		eventTypeVersion:  cloudEvent.EventTypeVersion,
	}
}
