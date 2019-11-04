package testing

import (
	eventingv1alpha1 "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const FakeSubscriptionName = "fake-sub"

// redefine here to avoid cyclic dependency
const (
	integrationNamespace                      = "kyma-integration"
	applicationNameLabelKey                   = "applicationName"
	brokerNamespaceLabelKey                   = "brokerNamespace"
	knativeEventingInjectionLabelKey          = "knative-eventing-injection"
	knativeEventingInjectionLabelValueEnabled = "enabled"
)

func NewAppSubscription(appNs, appName string) *eventingv1alpha1.Subscription {
	return &eventingv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:      FakeSubscriptionName,
			Namespace: integrationNamespace,
			Labels: map[string]string{
				brokerNamespaceLabelKey: appNs,
				applicationNameLabelKey: appName,
			},
		},
	}
}

func NewAppNamespace(name string, brokerInjection bool) *corev1.Namespace {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	if brokerInjection {
		ns.Labels = map[string]string{
			knativeEventingInjectionLabelKey: knativeEventingInjectionLabelValueEnabled,
		}
	}
	return ns
}

func NewDefaultBroker(ns string) *eventingv1alpha1.Broker {
	return &eventingv1alpha1.Broker{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: ns,
		},
	}
}
