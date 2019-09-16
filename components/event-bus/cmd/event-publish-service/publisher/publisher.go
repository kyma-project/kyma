package publisher

import (
	"log"

	messagingV1Alpha1 "github.com/knative/eventing/pkg/apis/messaging/v1alpha1"
	api "github.com/kyma-project/kyma/components/event-bus/api/publish"
	"github.com/kyma-project/kyma/components/event-bus/cmd/event-publish-service/metrics"
	knative "github.com/kyma-project/kyma/components/event-bus/internal/knative/util"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// FAILED status label
	FAILED = "failed"
	// IGNORED status label
	IGNORED = "ignored"
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
		eventType string, eventTypeVersion string) (*api.Error, string, string)
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
	eventType string, eventTypeVersion string) (*api.Error,
	string, string) {

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

	if len(source) == 0 || len(eventType) == 0 || len(eventTypeVersion) == 0 {
		log.Println("one of the event source, type or version value is missing")
		return api.ErrorResponseInternalServer(), FAILED, empty
	}

	//Adding the event-metadata as channel labels
	knativeChannelLabels := make(map[string]string)
	knativeChannelLabels[subscriptionSourceID] = source
	knativeChannelLabels[subscriptionEventType] = eventType
	knativeChannelLabels[subscriptionEventTypeVersion] = eventTypeVersion

	//Fetch knative channel via label
	channel, err := knativeLib.GetChannelByLabels(*namespace, &knativeChannelLabels)
	if err != nil {
		log.Printf("an error occured while trying to get the knative channel for source: '%v', event-type: '%v', event-type-version: '%v' in namespace '%v'\n"+
			"error: '%v':", source, eventType, eventTypeVersion, *namespace, err)
		log.Println("incrementing ignored messages counter")
		metrics.TotalPublishedMessages.With(prometheus.Labels{
			metrics.Namespace:        *namespace,
			metrics.Status:           IGNORED,
			metrics.SourceID:         source,
			metrics.EventType:        eventType,
			metrics.EventTypeVersion: eventTypeVersion}).Inc()
		return nil, IGNORED, empty
	}

	return publisher.publishOnChannel(knativeLib, channel, namespace, headers, payload)
}

func (publisher *DefaultKnativePublisher) publishOnChannel(knativeLib *knative.KnativeLib, channel *messagingV1Alpha1.Channel, namespace *string, headers *map[string][]string,
	payload *[]byte) (*api.Error, string, string) {

	// knative Channel reference should not be nil
	if channel == nil {
		log.Println("knative channel reference is not passed")
		return api.ErrorResponseInternalServer(), FAILED, empty
	}

	// send message to the knative channel
	messagePayload := string(*payload)
	err := knativeLib.SendMessage(channel, headers, &messagePayload)
	if err != nil {
		log.Printf("failed to send message to the knative channel '%v' in namespace '%v'", channel.Name, *namespace)
		return api.ErrorResponseInternalServer(), FAILED, empty
	}

	// publish to channel succeeded return nil error
	return nil, PUBLISHED, channel.Name
}
