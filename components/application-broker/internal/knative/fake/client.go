package fake

import (
	eventingv1alpha1 "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	messagingv1alpha1 "github.com/knative/eventing/pkg/apis/messaging/v1alpha1"
	"k8s.io/api/core/v1"

	"github.com/kyma-project/kyma/components/application-broker/internal/knative"
)

// KnClient is a fake Knative client used for testing.
type KnClient struct{}

func (k *KnClient) GetChannelByLabels(ns string, labels map[string]string) (*messagingv1alpha1.Channel, error) {
	panic("implement me")
}

func (k *KnClient) GetSubscriptionByLabels(ns string, labels map[string]string) (*eventingv1alpha1.Subscription, error) {
	panic("implement me")
}

func (k *KnClient) CreateSubscription(*eventingv1alpha1.Subscription) (*eventingv1alpha1.Subscription, error) {
	panic("implement me")
}

func (k *KnClient) UpdateSubscription(*eventingv1alpha1.Subscription) (*eventingv1alpha1.Subscription, error) {
	panic("implement me")
}

func (k *KnClient) DeleteSubscription(*eventingv1alpha1.Subscription) error {
	panic("implement me")
}

func (k *KnClient) GetDefaultBroker(ns string) (*eventingv1alpha1.Broker, error) {
	panic("implement me")
}

func (k *KnClient) DeleteBroker(*eventingv1alpha1.Broker) error {
	panic("implement me")
}

func (k *KnClient) GetNamespace(name string) (*v1.Namespace, error) {
	panic("implement me")
}

func (k *KnClient) UpdateNamespace(*v1.Namespace) (*v1.Namespace, error) {
	panic("implement me")
}

// compile time contract check
var _ knative.Client = &KnClient{}

// todo
func NewKnativeClient() knative.Client {
	// init the Knative client
	knativeClient := &KnClient{}
	return knativeClient
}
