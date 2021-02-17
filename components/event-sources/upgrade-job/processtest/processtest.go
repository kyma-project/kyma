package processtest

// processtest package provides utilities for Process testing.

import (
	"fmt"

	kymaeventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"

	corev1 "k8s.io/api/core/v1"

	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	knativeapis "knative.dev/pkg/apis"

	eventsourcesv1alpha1 "github.com/kyma-project/kyma/components/event-sources/apis/sources/v1alpha1"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	applicationconnectorv1alpha1 "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	eventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
)

const (
	KymaIntegrationNamespace = "kyma-integration"
)

func NewApp(name string) applicationconnectorv1alpha1.Application {
	return applicationconnectorv1alpha1.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Application",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: applicationconnectorv1alpha1.ApplicationSpec{},
	}
}

func NewApps() *applicationconnectorv1alpha1.ApplicationList {
	app1 := NewApp("app1")
	return &applicationconnectorv1alpha1.ApplicationList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ApplicationList",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		Items: []applicationconnectorv1alpha1.Application{
			app1,
		},
	}
}

func NewValidators() *appsv1.DeploymentList {
	validator := NewValidator("app1")
	return &appsv1.DeploymentList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DeploymentList",
			APIVersion: "apps/v1",
		},
		Items: []appsv1.Deployment{
			validator,
		},
	}
}

func NewValidator(appName string) appsv1.Deployment {
	name := fmt.Sprintf("%s-connectivity-validator", appName)
	return appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: KymaIntegrationNamespace,
			Annotations: map[string]string{
				"meta.helm.sh/release-name": appName,
			},
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: fmt.Sprintf("%s-connectivity-validator", appName),
						},
					},
				},
			},
		},
	}
}

func NewEventServices() *appsv1.DeploymentList {
	validator := NewEventService("app1")
	return &appsv1.DeploymentList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DeploymentList",
			APIVersion: "apps/v1",
		},
		Items: []appsv1.Deployment{
			*validator,
		},
	}
}

func NewEventService(appName string) *appsv1.Deployment {
	name := fmt.Sprintf("%s-event-service", appName)
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: KymaIntegrationNamespace,
			Labels: map[string]string{
				"app": name,
			},
		},
	}
}

func NewTriggers() *eventingv1alpha1.TriggerList {
	trigger1 := NewTrigger("trigger1", "test1", "type1", "source1", "v1")
	trigger2 := NewTrigger("trigger2", "test1", "type2", "source2", "v1")
	return &eventingv1alpha1.TriggerList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "TriggerList",
			APIVersion: "eventing.knative.dev/v1alpha1",
		},
		Items: []eventingv1alpha1.Trigger{
			trigger1,
			trigger2,
		},
	}
}

func NewTrigger(name, namespace, eventType, source, version string) eventingv1alpha1.Trigger {
	return eventingv1alpha1.Trigger{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Trigger",
			APIVersion: "eventing.knative.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: eventingv1alpha1.TriggerSpec{
			Broker: "default",
			Filter: &eventingv1alpha1.TriggerFilter{
				Attributes: &eventingv1alpha1.TriggerFilterAttributes{
					"type":             eventType,
					"source":           source,
					"eventtypeversion": version,
				},
			},
			Subscriber: duckv1.Destination{
				URI: &knativeapis.URL{
					Scheme: "http",
					Host:   "host.com",
				},
			},
		},
	}
}

func NewTriggerWithoutFilter(name, namespace string) eventingv1alpha1.Trigger {
	return eventingv1alpha1.Trigger{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Trigger",
			APIVersion: "eventing.knative.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}

func NewHTTPSources() *eventsourcesv1alpha1.HTTPSourceList {
	httpSource1 := NewHTTPSource("app1")
	return &eventsourcesv1alpha1.HTTPSourceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		Items: []eventsourcesv1alpha1.HTTPSource{
			*httpSource1,
		},
	}
}

func NewHTTPSource(appName string) *eventsourcesv1alpha1.HTTPSource {
	return &eventsourcesv1alpha1.HTTPSource{
		ObjectMeta: metav1.ObjectMeta{
			Name: appName,
		},
		Spec: eventsourcesv1alpha1.HTTPSourceSpec{
			Source: appName,
		},
	}
}

func NewAppSubscriptions() *messagingv1alpha1.SubscriptionList {
	sub1 := NewAppSubscription("app1")
	return &messagingv1alpha1.SubscriptionList{
		Items: []messagingv1alpha1.Subscription{*sub1},
	}
}

func NewAppSubscription(appName string) *messagingv1alpha1.Subscription {
	return &messagingv1alpha1.Subscription{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Subscription",
			APIVersion: "messaging.knative.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: "kyma-integration",
			Labels: map[string]string{
				"application-name": appName,
			},
		},
		Spec: messagingv1alpha1.SubscriptionSpec{
			Channel: corev1.ObjectReference{
				APIVersion: "messaging.knative.dev/v1alpha1",
				Kind:       "Channel",
				Name:       appName,
			},
		},
	}
}

func NewAppChannels() *messagingv1alpha1.ChannelList {
	channel1 := NewAppChannel("app1")
	return &messagingv1alpha1.ChannelList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ChannelList",
			APIVersion: "messaging.knative.dev/v1alpha1",
		},
		Items: []messagingv1alpha1.Channel{*channel1},
	}
}

func NewAppChannel(appName string) *messagingv1alpha1.Channel {
	return &messagingv1alpha1.Channel{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Channel",
			APIVersion: "messaging.knative.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: "kyma-integration",
			Labels: map[string]string{
				"application-name": appName,
			},
		},
	}
}

func NewBrokers() *eventingv1alpha1.BrokerList {
	broker1 := NewBroker("test1")
	return &eventingv1alpha1.BrokerList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "BrokerList",
			APIVersion: "eventing.knative.dev/v1alpha1",
		},
		Items: []eventingv1alpha1.Broker{
			*broker1,
		},
	}
}

func NewBroker(namespace string) *eventingv1alpha1.Broker {
	return &eventingv1alpha1.Broker{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Broker",
			APIVersion: "eventing.knative.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: namespace,
		},
	}
}

func NewNamespaces() *corev1.NamespaceList {
	test1 := NewNamespace("test1")
	return &corev1.NamespaceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "NamespaceList",
			APIVersion: "v1",
		},
		Items: []corev1.Namespace{*test1},
	}
}

func NewNamespace(name string) *corev1.Namespace {
	return &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"knative-eventing-injection": "enabled",
			},
		},
	}
}

type SubOpts func(subscription *kymaeventingv1alpha1.Subscription)

func NewKymaSubscription(name, namespace string, opts ...SubOpts) kymaeventingv1alpha1.Subscription {
	sub := kymaeventingv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	for _, opt := range opts {
		opt(&sub)
	}
	return sub
}

func WithSink(sub *kymaeventingv1alpha1.Subscription) {
	sub.Spec.Sink = "http://host.com"
}

func WithDefaultProtocolSetting(sub *kymaeventingv1alpha1.Subscription) {
	sub.Spec.ProtocolSettings = new(kymaeventingv1alpha1.ProtocolSettings)
	sub.Spec.Protocol = ""
}

func WithFilters(eventType, namespace, version, appName string, sub *kymaeventingv1alpha1.Subscription) {
	prefix := "prefix"
	sourceFilter := kymaeventingv1alpha1.BebFilter{
		EventSource: &kymaeventingv1alpha1.Filter{
			Type:     "exact",
			Property: "source",
			Value:    namespace,
		},
		EventType: &kymaeventingv1alpha1.Filter{
			Type:     "exact",
			Property: "type",
			Value:    fmt.Sprintf("%s.%s.%s.%s", prefix, appName, eventType, version),
		},
	}
	sub.Spec.Filter = &kymaeventingv1alpha1.BebFilters{
		Filters: []*kymaeventingv1alpha1.BebFilter{
			&sourceFilter,
		},
	}
}
