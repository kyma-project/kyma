package util

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	evapisv1alpha1 "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	evclientsetfake "github.com/knative/eventing/pkg/client/clientset/versioned/fake"
	"github.com/knative/pkg/apis"
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8stesting "k8s.io/client-go/testing"
)

const (
	channelName      = "test-channel"
	testNS           = "test-namespace"
	provisioner      = "test-provisioner"
	subscriptionName = "test-subscription"
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
			Labels: map[string]string{
				"l1": "v1",
				"l2": "v2",
			},
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
	labels = map[string]string{
		"l1": "v1",
		"l2": "v2",
	}
	labels2 = map[string]string{
		"l1": "v13",
		"l2": "v23",
	}
)

func Test_CreateChannel(t *testing.T) {
	log.Print("Test_CreateChannel")
	client := evclientsetfake.NewSimpleClientset()
	client.Fake.ReactionChain = nil
	client.Fake.AddReactor("create", "channels", func(action k8stesting.Action) (handled bool,
		ret runtime.Object, err error) {
		return true, testChannel, nil
	})

	k := &KnativeLib{
		evClient: client.EventingV1alpha1(),
	}
	ch, err := k.CreateChannel(provisioner, channelName, testNS, &labels, 10*time.Second)
	assert.Nil(t, err)
	log.Printf("Channel created: %v", ch)

	ignore := cmpopts.IgnoreTypes(apis.VolatileTime{})
	if diff := cmp.Diff(testChannel, ch, ignore); diff != "" {
		t.Errorf("%s (-want, +got) = %v", "Test_CreateChannel", diff)
	}
}

func Test_CreateChannelWithError(t *testing.T) {
	log.Print("Test_CreateChannel")
	client := evclientsetfake.NewSimpleClientset()
	client.Fake.ReactionChain = nil
	client.Fake.AddReactor("create", "channels", func(action k8stesting.Action) (handled bool,
		ret runtime.Object, err error) {
		tc := testChannel.DeepCopy()
		tc.Labels = labels2
		return true, tc, nil
	})

	k := &KnativeLib{
		evClient: client.EventingV1alpha1(),
	}
	ch, err := k.CreateChannel(provisioner, channelName, testNS, &labels, 10*time.Second)
	assert.Nil(t, err)
	log.Printf("Channel created: %v", ch)

	ignore := cmpopts.IgnoreTypes(apis.VolatileTime{})
	if diff := cmp.Diff(testChannel, ch, ignore); diff != "" {
		t.Logf("%s (-want, +got) = %v;\n want should be: %v;\n got should be: %v", "Test_CreateChannel",
			diff, labels, labels2)
	} else {
		t.Error("Test_CreateChannelWithError should return different labels")
	}
}

func Test_CreateChannelTimeout(t *testing.T) {
	log.Print("Test_CreateChannelTimeout")
	client := evclientsetfake.NewSimpleClientset()
	client.Fake.ReactionChain = nil
	client.Fake.AddReactor("create", "channels", func(action k8stesting.Action) (handled bool,
		ret runtime.Object, err error) {
		notReadyCondition := duckv1alpha1.Condition{
			Type: evapisv1alpha1.ChannelConditionReady, Status: corev1.ConditionFalse}
		tc := testChannel.DeepCopy()
		tc.Status.Conditions[0] = notReadyCondition
		return true, tc, nil
	})

	k := &KnativeLib{
		evClient: client.EventingV1alpha1(),
	}
	_, err := k.CreateChannel(provisioner, channelName, testNS, &labels, 1*time.Second)
	assert.NotNil(t, err)
	log.Printf("Test_CreateChannelTimeout: %v", err)
}

func Test_SendMessage(t *testing.T) {
	log.Print("Test_SendMessage")
	// create the test http server
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()

		log.Printf("Message received: %v", fmt.Sprintf("%s", fmt.Sprintf("%s", body)))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
	})
	srv := httptest.NewServer(handler)
	log.Printf("test srv URL: %v", srv.URL)

	// create a KN channel to connect to test http server
	client := evclientsetfake.NewSimpleClientset()
	client.Fake.ReactionChain = nil
	client.Fake.AddReactor("create", "channels", func(action k8stesting.Action) (handled bool,
		ret runtime.Object, err error) {
		return true, testChannel, nil
	})
	k := &KnativeLib{}
	k.InjectClient(client.EventingV1alpha1())
	ch, err := k.CreateChannel(provisioner, channelName, testNS, &labels, 10*time.Second)
	assert.Nil(t, err)
	u, err := url.Parse(srv.URL)
	assert.Nil(t, err)
	ch.Status.SetAddress(u.Hostname() + ":" + u.Port())

	// send a message to the channel
	p := "message 1"
	h := make(map[string]string)
	h["test"] = "test"
	err = k.SendMessage(ch, &h, &p)
	assert.Nil(t, err)
}

func Test_InjectClient(t *testing.T) {
	log.Print("Test_InjectClient")
	client := evclientsetfake.NewSimpleClientset().EventingV1alpha1()
	k := &KnativeLib{}
	err := k.InjectClient(client)
	assert.Nil(t, err)
}

func Test_DeleteInexistentChannel(t *testing.T) {
	log.Print("Test_DeleteInexistentChannel")
	client := evclientsetfake.NewSimpleClientset().EventingV1alpha1()
	k := &KnativeLib{}
	k.InjectClient(client)
	err := k.DeleteChannel(channelName, testNS)
	assert.True(t, k8serrors.IsNotFound(err))
}

func Test_CreateDeleteChannel(t *testing.T) {
	log.Print("Test_CreateDeleteChannel")
	client := evclientsetfake.NewSimpleClientset()
	client.Fake.ReactionChain = nil
	client.Fake.AddReactor("create", "channels", func(action k8stesting.Action) (handled bool,
		ret runtime.Object, err error) {
		return true, testChannel, nil
	})
	k := &KnativeLib{}
	k.InjectClient(client.EventingV1alpha1())
	ch, err := k.CreateChannel(provisioner, channelName, testNS, &labels, 1*time.Second)
	assert.Nil(t, err)
	err = k.DeleteChannel(ch.Name, ch.Namespace)
	assert.Nil(t, err)
}

func Test_CreateSubscription(t *testing.T) {
	log.Print("Test_CreateSubscription")
	k := &KnativeLib{}
	k.InjectClient(evclientsetfake.NewSimpleClientset().EventingV1alpha1())
	var uri = "dnsName: hello-00001-service.default"
	err := k.CreateSubscription(subscriptionName, testNS, channelName, &uri)
	assert.Nil(t, err)
}

func Test_DeleteInexistentSubscription(t *testing.T) {
	log.Print("Test_DeleteInexistentSubscription")
	k := &KnativeLib{}
	k.InjectClient(evclientsetfake.NewSimpleClientset().EventingV1alpha1())
	err := k.DeleteSubscription(subscriptionName, testNS)
	assert.True(t, k8serrors.IsNotFound(err))
}

func Test_CreateDeleteSubscription(t *testing.T) {
	log.Print("Test_CreateDeleteSubscription")
	k := &KnativeLib{}
	k.InjectClient(evclientsetfake.NewSimpleClientset().EventingV1alpha1())
	var uri = "dnsName: hello-00001-service.default"
	err := k.CreateSubscription(subscriptionName, testNS, channelName, &uri)
	assert.Nil(t, err)
	err = k.DeleteSubscription(subscriptionName, testNS)
	assert.Nil(t, err)
}

func Test_CreateSubscriptionAgain(t *testing.T) {
	log.Print("Test_CreateSubscriptionAgain")
	k := &KnativeLib{}
	k.InjectClient(evclientsetfake.NewSimpleClientset().EventingV1alpha1())
	var uri = "dnsName: hello-00001-service.default"
	err := k.CreateSubscription(subscriptionName, testNS, channelName, &uri)
	assert.Nil(t, err)
	err = k.CreateSubscription(subscriptionName, testNS, channelName, &uri)
	assert.True(t, k8serrors.IsAlreadyExists(err))
}
