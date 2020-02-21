package test

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	eventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	eventingclientv1alpha1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/eventing/v1alpha1"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/v1alpha1"
)

const testSubscriptionName = "test-sub-dqwawshakjqmxifnc"

type TriggerClient interface {
	Create(namespace, application, eventType string) error
	Delete(namespace string) error
}

type client struct {
	knEventingClient *eventingclientv1alpha1.EventingV1alpha1Client
}

func NewTriggerClient() (TriggerClient, error) {
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return initClient(k8sConfig)
}

func initClient(k8sConfig *rest.Config) (TriggerClient, error) {
	kneventingClient, err := eventingclientv1alpha1.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}
	return &client{
		knEventingClient: kneventingClient,
	}, nil
}

func (tc *client) Create(namespace, application, eventType string) error {
	url, err := apis.ParseURL("http://some-host.ns:8080/")
	if err != nil {
		return err
	}

	t := &eventingv1alpha1.Trigger{
		ObjectMeta: v1.ObjectMeta{
			Name:      testSubscriptionName,
			Namespace: namespace,
		},
		Spec: eventingv1alpha1.TriggerSpec{
			Broker: "default",
			Filter: &eventingv1alpha1.TriggerFilter{
				Attributes: &eventingv1alpha1.TriggerFilterAttributes{
					"source":           application,
					"type":             eventType,
					"eventtypeversion": "v1",
				},
			},
			Subscriber: &v1alpha1.Destination{
				URI: url,
			},
		},
	}
	_, e := tc.knEventingClient.Triggers(namespace).Create(t)
	return e
}

func (tc *client) Delete(namespace string) error {
	return tc.knEventingClient.Triggers(namespace).Delete(testSubscriptionName, &v1.DeleteOptions{})
}
