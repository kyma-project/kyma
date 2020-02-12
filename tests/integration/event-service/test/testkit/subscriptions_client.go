package testkit

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	v1alpha12 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	eventingv1alpha1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/eventing/v1alpha1"
	"knative.dev/pkg/apis"
	v1alpha13 "knative.dev/pkg/apis/v1alpha1"
)

const testSubscriptionName = "test-sub-dqwawshakjqmxifnc"

type TriggerClient interface {
	Create(namespace, application, eventType string) error
	Delete(namespace string) error
}

type client struct {
	knEventingClient *eventingv1alpha1.EventingV1alpha1Client
}

func NewSubscriptionsClient() (TriggerClient, error) {
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return initClient(k8sConfig)
}

func initClient(k8sConfig *rest.Config) (TriggerClient, error) {
	kneventingClient, err := eventingv1alpha1.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}
	return &client{
		knEventingClient: kneventingClient,
	}, nil
}

func (tc *client) Create(namespace, application, eventType string) error {
	t := &v1alpha12.Trigger{
		ObjectMeta: v1.ObjectMeta{
			Name:      testSubscriptionName,
			Namespace: namespace,
		},
		Spec: v1alpha12.TriggerSpec{
			Broker: "default",
			Filter: &v1alpha12.TriggerFilter{
				Attributes: &v1alpha12.TriggerFilterAttributes{
					"source":           application,
					"type":             eventType,
					"eventtypeversion": "v1",
				},
			},
			Subscriber: &v1alpha13.Destination{
				URI: &apis.URL{
					Host: "https://some.test.endpoint",
				},
			},
		},
	}
	_, e := tc.knEventingClient.Triggers(namespace).Create(t)
	return e
}

func (tc *client) Delete(namespace string) error {
	return tc.knEventingClient.Triggers(namespace).Delete(testSubscriptionName, &v1.DeleteOptions{})
}
