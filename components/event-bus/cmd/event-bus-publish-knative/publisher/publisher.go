package publisher

import (
	"log"

	api "github.com/kyma-project/kyma/components/event-bus/api/publish"
	knative "github.com/kyma-project/kyma/components/event-bus/internal/knative/util"
)

type KnativePublisher interface {
	Publish(knativeLib *knative.KnativeLib, channelName *string, namespace *string, headers *map[string]string, payload *[]byte) *api.Error
}

type DefaultKnativePublisher struct{}

func NewKnativePublisher() KnativePublisher {
	publisher := new(DefaultKnativePublisher)
	return publisher
}

func (publisher *DefaultKnativePublisher) Publish(knativeLib *knative.KnativeLib, channelName *string, namespace *string, headers *map[string]string, payload *[]byte) *api.Error {
	// knativelib should not be nil
	if knativeLib == nil {
		log.Println("knative-lib is nil")
		return api.ErrorResponseInternalServer()
	}

	// channelName should be present
	if channelName == nil || len(*channelName) == 0 {
		log.Println("channelName is missing")
		return api.ErrorResponseInternalServer()
	}

	// namespace should be present
	if namespace == nil || len(*namespace) == 0 {
		log.Println("namespace is missing")
		return api.ErrorResponseInternalServer()
	}

	// headers should be present
	if headers == nil || len(*headers) == 0 {
		log.Println("headers are missing")
		return api.ErrorResponseInternalServer()
	}

	// payload should be present
	if payload == nil || len(*payload) == 0 {
		log.Println("payload is missing")
		return api.ErrorResponseInternalServer()
	}

	// get the knative channel
	channel, err := knativeLib.GetChannel(*channelName, *namespace)
	if err != nil || channel == nil {
		log.Printf("cannot find the knative channel '%v' in namespace '%v'", *channelName, *namespace)
		return api.ErrorResponseInternalServer()
	}

	// send message to the knative channel
	messagePayload := string(*payload)
	err = knativeLib.SendMessage(channel, headers, &messagePayload)
	if err != nil {
		log.Printf("failed to send message to the knative channel '%v' in namespace '%v'", *channelName, *namespace)
		return api.ErrorResponseInternalServer()
	}

	// publish to channel succeeded return nil error
	return nil
}
