package processtest

// processtest package provides utilities for Process testing.

import (
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
)

const (
	KymaSystemNamespace = "kyma-system"
	testingShootName    = "testing"
)

func NewEventingControllers() *appsv1.DeploymentList {
	validator := NewEventingController("eventing-controller")
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

func NewEventingController(name string) *appsv1.Deployment {
	var replicaCount int32 = 1
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: KymaSystemNamespace,
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicaCount,
		},
	}
}

func NewEventingPublishers() *appsv1.DeploymentList {
	validator := NewEventingPublisher("eventing-publisher-proxy")
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

func NewEventingPublisher(name string) *appsv1.Deployment {
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: KymaSystemNamespace,
			Labels: map[string]string{
				"app": name,
			},
		},
	}
}

func NewEventingBackends() *eventingv1alpha1.EventingBackendList {
	validator := NewEventingBackend("eventing-backend", true)
	return &eventingv1alpha1.EventingBackendList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "EventingBackend",
			APIVersion: "eventing.kyma-project.io/v1alpha1",
		},
		Items: []eventingv1alpha1.EventingBackend{
			*validator,
		},
	}
}

func NewEventingBackend(name string, isBebEnabled bool) *eventingv1alpha1.EventingBackend {
	backendType := eventingv1alpha1.NatsBackendType
	if isBebEnabled {
		backendType = eventingv1alpha1.BebBackendType
	}

	return &eventingv1alpha1.EventingBackend{
		TypeMeta: metav1.TypeMeta{
			Kind:       "EventingBackend",
			APIVersion: "eventing.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: KymaSystemNamespace,
			Labels: map[string]string{
				"kyma-project.io/eventing": "backend",
			},
		},
		Status: eventingv1alpha1.EventingBackendStatus{
			Backend: backendType,
		},
	}
}

func NewSecrets() *corev1.SecretList {
	validator := NewSecret("eventing-backend")
	return &corev1.SecretList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SecretList",
			APIVersion: "v1",
		},
		Items: []corev1.Secret{
			*validator,
		},
	}
}

func NewSecret(name string) *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: KymaSystemNamespace,
			Labels: map[string]string{
				"app": name,
			},
		},
	}
}

func NewConfigMaps() *corev1.ConfigMapList {
	validator := ConfigMap("shoot-info")
	return &corev1.ConfigMapList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMapList",
			APIVersion: "v1",
		},
		Items: []corev1.ConfigMap{
			*validator,
		},
	}
}

func ConfigMap(name string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: KymaSystemNamespace,
			Labels: map[string]string{
				"app": name,
			},
		},
		Data: map[string]string{
			"shootName": testingShootName,
		},
	}
}

func NewKymaSubscriptions() *eventingv1alpha1.SubscriptionList {
	sub1 := NewKymaSubscription("sub1", true, false)
	sub2 := NewKymaSubscription("sub2", true, false)
	sub3 := NewKymaSubscription("sub3", true, true)
	sub4 := NewKymaSubscription("sub4", false, false)

	return &eventingv1alpha1.SubscriptionList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Subscription",
			APIVersion: "eventing.kyma-project.io/v1alpha1",
		},
		Items: []eventingv1alpha1.Subscription{
			*sub1, *sub2, *sub3, *sub4,
		},
	}
}

func NewKymaSubscription(appName string, addConditions bool, includeBebMessageInCondition bool) *eventingv1alpha1.Subscription {
	subscription := &eventingv1alpha1.Subscription{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Subscription",
			APIVersion: "eventing.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: KymaSystemNamespace,
			Labels: map[string]string{
				"foo": "sub",
			},
		},
		Spec: eventingv1alpha1.SubscriptionSpec{
			ProtocolSettings: &eventingv1alpha1.ProtocolSettings{},
			Sink:             "http://test.test.svc.cluster.local",
		},
		Status: eventingv1alpha1.SubscriptionStatus{},
	}

	// Define conditions
	condition1 := eventingv1alpha1.Condition{
		Type:               eventingv1alpha1.ConditionSubscriptionActive,
		Reason:             "BEB Subscription active",
		Message:            "",
		Status:             "True",
		LastTransitionTime: metav1.Now(),
	}
	if includeBebMessageInCondition {
		nameMapper := handlers.NewBebSubscriptionNameMapper(testingShootName, handlers.MaxBEBSubscriptionNameLength)
		newBebSubscriptionName := nameMapper.MapSubscriptionName(subscription)
		condition1.Message = eventingv1alpha1.CreateMessageForConditionReasonSubscriptionCreated(newBebSubscriptionName)
	}

	condition2 := eventingv1alpha1.Condition{
		Type:               eventingv1alpha1.ConditionSubscribed,
		Reason:             "BEB Subscription creation failed",
		Message:            "",
		Status:             "False",
		LastTransitionTime: metav1.Now(),
	}
	condition3 := eventingv1alpha1.Condition{
		Type:               eventingv1alpha1.ConditionAPIRuleStatus,
		Reason:             "APIRule status ready",
		Message:            "",
		Status:             "True",
		LastTransitionTime: metav1.Now(),
	}

	if addConditions {
		subscription.Status.Conditions = []eventingv1alpha1.Condition{condition1, condition2, condition3}
	}

	return subscription
}
