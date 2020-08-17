package object

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
)

func TestNewChannel(t *testing.T) {
	ch := NewChannel(tNs, tName)

	expectCh := &messagingv1alpha1.Channel{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: tNs,
			Name:      tName,
		},
	}

	if d := cmp.Diff(expectCh, ch); d != "" {
		t.Errorf("Unexpected diff: (-:expect, +:got) %s", d)
	}
}

func TestApplyExistingChannelAttributes(t *testing.T) {
	existingChannel := &messagingv1alpha1.Channel{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       tNs,
			Name:            tName,
			ResourceVersion: "1",
			Annotations: map[string]string{
				knativeMessagingAnnotations[0]: "1",
				"another-annotation":           "some-value",
			},
		},
	}

	// Service with empty spec, status, annotations, ...
	channel := NewChannel(tNs, tName)
	ApplyExistingChannelAttributes(existingChannel, channel)

	expectedChannel := &messagingv1alpha1.Channel{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       tNs,
			Name:            tName,
			ResourceVersion: "1",
			Annotations: map[string]string{
				knativeMessagingAnnotations[0]: "1",
			},
		},
	}

	if d := cmp.Diff(expectedChannel, channel); d != "" {
		t.Errorf("Unexpected diff: (-:expect, +:got) %s", d)
	}
}
