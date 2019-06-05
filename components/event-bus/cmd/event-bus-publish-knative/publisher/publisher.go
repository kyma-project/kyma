package publisher

import (
	api "github.com/kyma-project/kyma/components/event-bus/api/publish"
	"github.com/kyma-project/kyma/components/event-bus/cmd/event-bus-publish-knative/metrics"
	knative "github.com/kyma-project/kyma/components/event-bus/internal/knative/util"
	"github.com/prometheus/client_golang/prometheus"
	"log"
)

const (
	FAILED    = "failed"
	IGNORED   = "ignored"
	PUBLISHED = "published"
)

type KnativePublisher interface {
	Publish(knativeLib *knative.KnativeLib, channelName *string, namespace *string, headers *map[string][]string,
		payload *[]byte, publishRequest *api.PublishRequest) (*api.Error, string)
}

type DefaultKnativePublisher struct{}

func NewKnativePublisher() KnativePublisher {
	publisher := new(DefaultKnativePublisher)
	return publisher
}

func (publisher *DefaultKnativePublisher) Publish(knativeLib *knative.KnativeLib, channelName *string,
	namespace *string, headers *map[string][]string, payload *[]byte, publishRequest *api.PublishRequest) (*api.Error,
	string) {
	// knativelib should not be nil
	if knativeLib == nil {
		log.Println("knative-lib is nil")
		return api.ErrorResponseInternalServer(), FAILED
	}

	// channelName should be present
	if channelName == nil || len(*channelName) == 0 {
		log.Println("channelName is missing")
		return api.ErrorResponseInternalServer(), FAILED
	}

	// namespace should be present
	if namespace == nil || len(*namespace) == 0 {
		log.Println("namespace is missing")
		return api.ErrorResponseInternalServer(), FAILED
	}

	// headers should be present
	if headers == nil || len(*headers) == 0 {
		log.Println("headers are missing")
		return api.ErrorResponseInternalServer(), FAILED
	}

	// payload should be present
	if payload == nil || len(*payload) == 0 {
		log.Println("payload is missing")
		return api.ErrorResponseInternalServer(), FAILED
	}

	// get the knative channel
	channel, err := knativeLib.GetChannel(*channelName, *namespace)
	if err != nil {
		log.Printf("cannot find the knative channel '%v' in namespace '%v'", *channelName, *namespace)
		log.Println("incrementing ignored messages counter")
		metrics.TotalPublishedMessages.With(prometheus.Labels{
			metrics.Namespace:        *namespace,
			metrics.Status:           IGNORED,
			metrics.SourceID:         publishRequest.SourceID,
			metrics.EventType:        publishRequest.EventType,
			metrics.EventTypeVersion: publishRequest.EventTypeVersion}).Inc()
		return nil, IGNORED
	}

	// send message to the knative channel
	messagePayload := string(*payload)
	err = knativeLib.SendMessage(channel, headers, &messagePayload)
	if err != nil {
		log.Printf("failed to send message to the knative channel '%v' in namespace '%v'", *channelName, *namespace)
		return api.ErrorResponseInternalServer(), FAILED
	}

	// publish to channel succeeded return nil error
	return nil, PUBLISHED
}
