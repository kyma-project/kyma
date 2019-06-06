package fake

import (
	api "github.com/kyma-project/kyma/components/event-bus/api/publish"
	"github.com/kyma-project/kyma/components/event-bus/cmd/event-bus-publish-knative/publisher"
	knative "github.com/kyma-project/kyma/components/event-bus/internal/knative/util"
)

type MockKnativePublisher struct{}

func NewMockKnativePublisher() publisher.KnativePublisher {
	mockPublisher := new(MockKnativePublisher)
	return mockPublisher
}

func (m *MockKnativePublisher) Publish(knativeLib *knative.KnativeLib, channelName *string, namespace *string,
	headers *map[string][]string, payload *[]byte, publishRequest *api.PublishRequest) (*api.Error, string) {
	return nil, publisher.PUBLISHED
}
