package testkit

import (
	"github.com/kyma-project/kyma/components/event-bus/apis/eventing/v1alpha1"
	subscriptions "github.com/kyma-project/kyma/components/event-bus/client/generated/clientset/internalclientset"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

const testSubscriptionName = "test-sub-dqwawshakjqmxifnc"

type SubscriptionsClient interface {
	Create(namespace, application, eventType string) error
	Delete(namespace string) error
}

type client struct {
	subscriptions *subscriptions.Clientset
}

func NewSubscriptionsClient() (SubscriptionsClient, error) {
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return initClient(k8sConfig)
}

func initClient(k8sConfig *rest.Config) (SubscriptionsClient, error) {
	subscriptionsClient, err := subscriptions.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	return &client{
		subscriptions: subscriptionsClient,
	}, nil
}

func (sc *client) Create(namespace, application, eventType string) error {
	subscription := &v1alpha1.Subscription{
		TypeMeta: v1.TypeMeta{
			Kind:       "Subscription",
			APIVersion: "eventing.kyma-project.io/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      testSubscriptionName,
			Namespace: namespace,
		},
		SubscriptionSpec: v1alpha1.SubscriptionSpec{
			Endpoint:                      "https://some.test.endpoint",
			IncludeSubscriptionNameHeader: true,
			EventType:                     eventType,
			EventTypeVersion:              "v1",
			SourceID:                      application,
		},
	}
	_, e := sc.subscriptions.EventingV1alpha1().Subscriptions(namespace).Create(subscription)
	return e
}

func (sc *client) Delete(namespace string) error {
	return sc.subscriptions.EventingV1alpha1().Subscriptions(namespace).Delete(testSubscriptionName, &v1.DeleteOptions{})
}
