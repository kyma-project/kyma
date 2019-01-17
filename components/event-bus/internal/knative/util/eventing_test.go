package util

import (
	"log"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/knative/pkg/apis"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	evapisv1alpha1 "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	evclientsetfake "github.com/knative/eventing/pkg/client/clientset/versioned/fake"
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	k8stesting "k8s.io/client-go/testing"
)

const (
	channelName = "test-channel"
	testNS      = "test-namespace"
	provisioner = "test-provisioner"
)

var (
	testChannel = &evapisv1alpha1.Channel{
		TypeMeta: metav1.TypeMeta{
			APIVersion: evapisv1alpha1.SchemeGroupVersion.String(),
			Kind:       "Channel",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNS,
			Name:      channelName,
		},
		Spec: evapisv1alpha1.ChannelSpec{
			Provisioner: &corev1.ObjectReference{
				Name: provisioner,
			},
		},
		Status: evapisv1alpha1.ChannelStatus{
			Conditions: []duckv1alpha1.Condition{
				{
					Type:   evapisv1alpha1.ChannelConditionReady,
					Status: corev1.ConditionTrue,
				},
			},
		},
	}
)

func Test_CreateChannel(t *testing.T) {
	log.Print("Test_CreateChannel")
	client := evclientsetfake.NewSimpleClientset()
	client.Fake.ReactionChain = nil
	client.Fake.AddReactor("create","channels", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, testChannel, nil
	})

	k := &KnativeLib {
		evClient: client.EventingV1alpha1(),
	}
	ch, err := k.CreateChannel(provisioner, channelName, testNS, 10 * time.Second)
	assert.Nil(t, err)
	log.Printf("Channel created: %v", ch)

	ignore := cmpopts.IgnoreTypes(apis.VolatileTime{})
	if diff := cmp.Diff(testChannel, ch, ignore); diff != "" {
		t.Errorf("%s (-want, +got) = %v", "Test_CreateChannel", diff)
	}
}

func Test_CreateChannelTimeout(t *testing.T) {
	log.Print("Test_CreateChannelTimeout")
	client := evclientsetfake.NewSimpleClientset()
	client.Fake.ReactionChain = nil
	client.Fake.AddReactor("create","channels", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		notReadyCondition := duckv1alpha1.Condition{Type: evapisv1alpha1.ChannelConditionReady, Status: corev1.ConditionFalse}
		testChannel.Status.Conditions[0] = notReadyCondition
		return true, testChannel, nil
	})

	k := &KnativeLib {
		evClient: client.EventingV1alpha1(),
	}
	_, err := k.CreateChannel(provisioner, channelName, testNS, 1 * time.Second)
	assert.NotNil(t, err)
	log.Printf("Test_CreateChannelTimeout: %v", err)
}

func Test_InjectClient(t *testing.T) {
	log.Print("Test_InjectClient")
	client := evclientsetfake.NewSimpleClientset().EventingV1alpha1()
	k := &KnativeLib {}
	err := k.InjectClient(client)
	assert.Nil(t, err)
}
