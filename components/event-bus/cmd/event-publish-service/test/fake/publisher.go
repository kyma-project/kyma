package fake

import (
	api "github.com/kyma-project/kyma/components/event-bus/api/publish"
	"github.com/kyma-project/kyma/components/event-bus/cmd/event-bus-publish-knative/publisher"
	knative "github.com/kyma-project/kyma/components/event-bus/internal/knative/util"
)

// MockKnativePublisher to mock the knative publisher for testing purposes.
type MockKnativePublisher struct{}

// NewMockKnativePublisher creates a new KnativePublisher instance.
func NewMockKnativePublisher() publisher.KnativePublisher {
	mockPublisher := new(MockKnativePublisher)
	return mockPublisher
}

// Publish for mocking the KnativePublisher.Publish behaviour.
func (m *MockKnativePublisher) Publish(knativeLib *knative.KnativeLib, channelName *string, namespace *string,
	headers *map[string][]string, payload *[]byte, source string, eventType string,
	eventTypeVersion string) (*api.Error, string) {
	return nil, publisher.PUBLISHED
}
