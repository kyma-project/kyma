package publisher

import (
	"log"

	"github.com/kyma-project/kyma/components/event-bus/cmd/event-publish-service/metrics"
	"github.com/prometheus/client_golang/prometheus"

	messagingV1Alpha1 "github.com/knative/eventing/pkg/apis/messaging/v1alpha1"
	api "github.com/kyma-project/kyma/components/event-bus/api/publish"
	knative "github.com/kyma-project/kyma/components/event-bus/internal/knative/util"
)

const (
	// FAILED status label
	FAILED = "failed"
	// IGNORED_CHANNEL_MISSING status label
	IGNORED_CHANNEL_MISSING = "ignored_channel_missing"
	// IGNORED_CHANNEL_NOT_READY status label
	IGNORED_CHANNEL_NOT_READY = "ignored_channel_not_ready"
	// PUBLISHED status label
	PUBLISHED = "published"

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
		return api.ErrorResponseInternalServer(), FAILED, empty
	}

	// namespace should be present
	if namespace == nil || len(*namespace) == 0 {
		log.Println("namespace is missing")
		return api.ErrorResponseInternalServer(), FAILED, empty
	}

	// headers should be present
	if headers == nil || len(*headers) == 0 {
		log.Println("headers are missing")
		return api.ErrorResponseInternalServer(), FAILED, empty
	}

	// payload should be present
	if payload == nil || len(*payload) == 0 {
		log.Println("payload is missing")
		return api.ErrorResponseInternalServer(), FAILED, empty
	}

	if len(source) == 0 {
		log.Println("source is missing")
		return api.ErrorResponseInternalServer(), FAILED, empty
	}

	if len(eventType) == 0 {
		log.Println("eventType is missing")
		return api.ErrorResponseInternalServer(), FAILED, empty
	}

	if len(eventTypeVersion) == 0 {
		log.Println("eventTypeVersion is missing")
		return api.ErrorResponseInternalServer(), FAILED, empty
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
		updateMetrics(channel, IGNORED_CHANNEL_MISSING, *namespace)
		return nil, IGNORED_CHANNEL_MISSING, empty
	}

	// If Knative channel is not ready there is no point in pushing it to dispatcher hence ignored
	if !channel.Status.IsReady() {
		log.Printf("knative channel is not ready :: for source: '%v', event-type: '%v', event-type-version: '%v' in namespace '%v' error: '%v':", source, eventType, eventTypeVersion, *namespace, err)
		log.Println("incrementing ignored messages counter")
		updateMetrics(channel, IGNORED_CHANNEL_NOT_READY, *namespace)
		return nil, IGNORED_CHANNEL_NOT_READY, empty
	}

	return publisher.publishOnChannel(knativeLib, channel, namespace, headers, payload)
}

func (publisher *DefaultKnativePublisher) publishOnChannel(knativeLib *knative.KnativeLib, channel *messagingV1Alpha1.Channel, namespace *string, headers *map[string][]string,
	payload *[]byte) (*api.Error, string, string) {

	// send message to the knative channel
	messagePayload := string(*payload)
	err := knativeLib.SendMessage(channel, headers, &messagePayload)
	if err != nil {
		log.Printf("failed to send message to the knative channel '%v' in namespace '%v'", channel.Name, *namespace)
		return api.ErrorResponseInternalServer(), FAILED, channel.Name
	}
	updateMetrics(channel, PUBLISHED, *namespace)

	// publish to channel succeeded return nil error
	return nil, PUBLISHED, channel.Name
}

func updateMetrics(channel *messagingV1Alpha1.Channel, status, ns string) {
	metrics.TotalPublishedMessages.With(prometheus.Labels{
		metrics.Namespace:        ns,
		metrics.Status:           status,
		metrics.SourceID:         channel.Labels[subscriptionSourceID],
		metrics.EventType:        channel.Labels[subscriptionEventType],
		metrics.EventTypeVersion: channel.Labels[subscriptionEventTypeVersion]}).Inc()
}
