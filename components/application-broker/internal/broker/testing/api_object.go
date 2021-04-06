package testing

import (
	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	FakeChannelName = "fake-chan"
)

// redefine here to avoid cyclic dependency
const (
	integrationNamespace    = "kyma-integration"
	applicationNameLabelKey = "application-name"
)

func NewAppNamespace(name string) *corev1.Namespace {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	return ns
}

func NewEventActivation(ns, name string) *v1alpha1.EventActivation {
	return &v1alpha1.EventActivation{
		TypeMeta: metav1.TypeMeta{
			Kind:       "EventActivation",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: v1alpha1.EventActivationSpec{
			DisplayName: "DisplayName",
			SourceID:    "source-id",
		},
	}
}
