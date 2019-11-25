package publisher

import (
	"log"

	"github.com/kyma-project/kyma/components/event-bus/cmd/event-publish-service/metrics"
	"github.com/prometheus/client_golang/prometheus"

	api "github.com/kyma-project/kyma/components/event-bus/api/publish"
	knative "github.com/kyma-project/kyma/components/event-bus/internal/knative/util"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
)

const (
	// Failed status label
	Failed = "failed"
	// IgnoredChannelMissing status label
	IgnoredChannelMissing = "ignored_channel_missing"
	// IgnoredChannelNotReady status label
	IgnoredChannelNotReady = "ignored_channel_not_ready"
	// Published status label
	Published = "published"

	empty = ""

	subscriptionSourceID         = "kyma-source-id"
	subscriptionEventType        = "kyma-event-type"
	subscriptionEventTypeVersion = "kyma-event-type-version"
)

// KnativePublisher encapsulates the publish behaviour.
type KnativePublisher interface {
	Publish(knativeLib *knative.KnativeLib, namespace *string, headers *map[string][]string,
		payload *[]byte, source string,
		eventType string, eventTypeVersion string) (error *api.Error, status string, channelName string)
}

// DefaultKnativePublisher is the default KnativePublisher instance.
type DefaultKnativePublisher struct{}

// NewKnativePublisher creates a new KnativePublisher instance.
func NewKnativePublisher() KnativePublisher {
	publisher := new(DefaultKnativePublisher)
	return publisher
}

// Publish events using the KnativeLib
func (publisher *DefaultKnativePublisher) Publish(knativeLib *knative.KnativeLib,
	namespace *string, headers *map[string][]string, payload *[]byte, source string,
	eventType string, eventTypeVersion string) (error *api.Error, status string, channelName string) {

	// knativelib should not be nil
	if knativeLib == nil {
		log.Println("knative-lib is nil")
		return api.ErrorResponseInternalServer(), Failed, empty
	}

	// namespace should be present
	if namespace == nil || len(*namespace) == 0 {
		log.Println("namespace is missing")
		return api.ErrorResponseInternalServer(), Failed, empty
	}

	// headers should be present
	if headers == nil || len(*headers) == 0 {
		log.Println("headers are missing")
		return api.ErrorResponseInternalServer(), Failed, empty
	}

	// payload should be present
	if payload == nil || len(*payload) == 0 {
		log.Println("payload is missing")
		return api.ErrorResponseInternalServer(), Failed, empty
	}

	if len(source) == 0 {
		log.Println("source is missing")
		return api.ErrorResponseInternalServer(), Failed, empty
	}

	if len(eventType) == 0 {
		log.Println("eventType is missing")
		return api.ErrorResponseInternalServer(), Failed, empty
	}

	if len(eventTypeVersion) == 0 {
		log.Println("eventTypeVersion is missing")
		return api.ErrorResponseInternalServer(), Failed, empty
	}
	//Adding the event-metadata as channel labels
	knativeChannelLabels := make(map[string]string)
	knativeChannelLabels[subscriptionSourceID] = source
	knativeChannelLabels[subscriptionEventType] = eventType
	knativeChannelLabels[subscriptionEventTypeVersion] = eventTypeVersion

	//Fetch knative channel via label
	channel, err := knativeLib.GetChannelByLabels(*namespace, knativeChannelLabels)
	if err != nil {
		log.Printf("failed to get the knative channel for source: '%v', event-type: '%v', event-type-version: '%v' in namespace '%v'\n"+
			"error: '%v':", source, eventType, eventTypeVersion, *namespace, err)
		log.Println("incrementing ignored messages counter")
		updateMetrics(knativeChannelLabels, IgnoredChannelMissing, *namespace)
		return nil, IgnoredChannelMissing, empty
	}

	// If Knative channel is not ready there is no point in pushing it to dispatcher hence ignored
	if !channel.Status.IsReady() {
		log.Printf("knative channel is not ready :: for source: '%v', event-type: '%v', event-type-version: '%v' in namespace '%v' error: '%v':", source, eventType, eventTypeVersion, *namespace, err)
		log.Println("incrementing ignored messages counter")
		updateMetrics(knativeChannelLabels, IgnoredChannelNotReady, *namespace)
		return nil, IgnoredChannelNotReady, empty
	}

	return publisher.publishOnChannel(knativeLib, channel, namespace, headers, payload, knativeChannelLabels)
}

func (publisher *DefaultKnativePublisher) publishOnChannel(knativeLib *knative.KnativeLib, channel *messagingv1alpha1.Channel, namespace *string, headers *map[string][]string,
	payload *[]byte, knativeChannelLabels map[string]string) (*api.Error, string, string) {
	// send message to the knative channel
	messagePayload := string(*payload)
	err := knativeLib.SendMessage(channel, headers, &messagePayload)
	if err != nil {
		log.Printf("failed to send message to the knative channel '%v' in namespace '%v'", channel.Name, *namespace)
		return api.ErrorResponseInternalServer(), Failed, channel.Name
	}
	updateMetrics(knativeChannelLabels, Published, *namespace)

	// publish to channel succeeded return nil error
	return nil, Published, channel.Name
}

func updateMetrics(knativeChannelLabels map[string]string, status, ns string) {
	metrics.TotalPublishedMessages.With(prometheus.Labels{
		metrics.Namespace:        ns,
		metrics.Status:           status,
		metrics.SourceID:         knativeChannelLabels[subscriptionSourceID],
		metrics.EventType:        knativeChannelLabels[subscriptionEventType],
		metrics.EventTypeVersion: knativeChannelLabels[subscriptionEventTypeVersion]}).Inc()
}
