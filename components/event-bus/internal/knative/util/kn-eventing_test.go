package util

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"knative.dev/pkg/apis/duck/v1alpha1"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	messagingv1alpha1 "github.com/knative/eventing/pkg/apis/messaging/v1alpha1"
	evclientsetfake "github.com/knative/eventing/pkg/client/clientset/versioned/fake"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8stesting "k8s.io/client-go/testing"
	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

const (
	channelName      = "test-channel"
	testNS           = "test-namespace"
	subscriptionName = "test-subscription"
)

var (
	testChannel = &messagingv1alpha1.Channel{
		TypeMeta: metav1.TypeMeta{
			APIVersion: messagingv1alpha1.SchemeGroupVersion.String(),
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
		Status: messagingv1alpha1.ChannelStatus{
			Status: duckv1beta1.Status{
				Conditions: duckv1beta1.Conditions{
					apis.Condition{
						Type:   apis.ConditionReady,
						Status: corev1.ConditionTrue,
					},
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
	testChannelList = &messagingv1alpha1.ChannelList{
		Items: []messagingv1alpha1.Channel{*testChannel},
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
		evClient:         client.EventingV1alpha1(),
		messagingChannel: client.MessagingV1alpha1(),
	}
	ch, err := k.CreateChannel(channelName, testNS, labels, 10*time.Second)
	assert.Nil(t, err)
	log.Printf("Channel created: %v", ch)

	ignore := cmpopts.IgnoreTypes(apis.VolatileTime{})
	if diff := cmp.Diff(testChannel, ch, ignore); diff != "" {
		t.Errorf("%s (-want, +got) = %v", "Test_CreateChannel", diff)
	}
}

func Test_GetChannelByLabels(t *testing.T) {
	log.Print("Test_GetChannelByLabels")
	log.Print("Creating Channel to fetch")
	client := evclientsetfake.NewSimpleClientset()
	client.Fake.ReactionChain = nil
	client.Fake.AddReactor("create", "channels", func(action k8stesting.Action) (handled bool,
		ret runtime.Object, err error) {
		return true, testChannel, nil
	})
	client.Fake.AddReactor("list", "channels", func(action k8stesting.Action) (handled bool,
		ret runtime.Object, err error) {
		return true, testChannelList, nil
	})

	k := &KnativeLib{
		evClient:         client.EventingV1alpha1(),
		messagingChannel: client.MessagingV1alpha1(),
	}
	ch1, err1 := k.CreateChannel(channelName, testNS, labels, 10*time.Second)
	assert.Nil(t, err1)
	log.Printf("Channel created: %v", ch1)
	log.Println("Getting Channel By label")

	ch2, err2 := k.GetChannelByLabels(testNS, labels)
	assert.Nil(t, err2)

	ignore := cmpopts.IgnoreTypes(apis.VolatileTime{})
	if diff := cmp.Diff(ch1, ch2, ignore); diff != "" {
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
		evClient:         client.EventingV1alpha1(),
		messagingChannel: client.MessagingV1alpha1(),
	}
	ch, err := k.CreateChannel(channelName, testNS, labels, 10*time.Second)
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
		notReadyCondition := apis.Condition{
			Type: messagingv1alpha1.ChannelConditionReady, Status: corev1.ConditionFalse}
		tc := testChannel.DeepCopy()
		tc.Status.Conditions[0] = notReadyCondition
		return true, tc, nil
	})

	k := &KnativeLib{
		evClient:         client.EventingV1alpha1(),
		messagingChannel: client.MessagingV1alpha1(),
	}
	_, err := k.CreateChannel(channelName, testNS, labels, 1*time.Second)
	assert.NotNil(t, err)
	log.Printf("Test_CreateChannelTimeout: %v", err)
}

func Test_SendMessage(t *testing.T) {
	log.Print("Test_SendMessage")
	// create the test http server
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		defer func() {
			_ = r.Body.Close()
		}()

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
	e := k.InjectClient(client.EventingV1alpha1(), client.MessagingV1alpha1())
	assert.Nil(t, e)
	ch, err := k.CreateChannel(channelName, testNS, labels, 10*time.Second)
	assert.Nil(t, err)
	u, err := url.Parse(srv.URL)
	assert.Nil(t, err)
	url := &apis.URL{Scheme: "http", Host: fmt.Sprintf("%s:%s", u.Hostname(), u.Port())}
	address := &v1alpha1.Addressable{
		Addressable: duckv1beta1.Addressable{URL: url},
		Hostname:    "u.Hostname()",
	}
	ch.Status.SetAddress(address)

	// send a message to the channel
	p := "message 1"
	h := make(map[string][]string)
	h["test"] = []string{"test"}
	err = k.SendMessage(ch, &h, &p)
	assert.Nil(t, err)
}

func Test_InjectClient(t *testing.T) {
	log.Print("Test_InjectClient")
	evClient := evclientsetfake.NewSimpleClientset().EventingV1alpha1()
	msgClient := evclientsetfake.NewSimpleClientset().MessagingV1alpha1()
	k := &KnativeLib{}
	err := k.InjectClient(evClient, msgClient)
	assert.Nil(t, err)
}

func Test_DeleteInexistentChannel(t *testing.T) {
	log.Print("Test_DeleteInexistentChannel")
	evClient := evclientsetfake.NewSimpleClientset().EventingV1alpha1()
	msgClient := evclientsetfake.NewSimpleClientset().MessagingV1alpha1()
	k := &KnativeLib{}
	e := k.InjectClient(evClient, msgClient)
	assert.Nil(t, e)
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
	e := k.InjectClient(client.EventingV1alpha1(), client.MessagingV1alpha1())
	assert.Nil(t, e)
	ch, err := k.CreateChannel(channelName, testNS, labels, 1*time.Second)
	assert.Nil(t, err)
	err = k.DeleteChannel(ch.Name, ch.Namespace)
	assert.Nil(t, err)
}

func Test_CreateSubscription(t *testing.T) {
	log.Print("Test_CreateSubscription")
	k := &KnativeLib{}
	e := k.InjectClient(evclientsetfake.NewSimpleClientset().EventingV1alpha1(), evclientsetfake.NewSimpleClientset().MessagingV1alpha1())
	assert.Nil(t, e)
	var uri = "dnsName: hello-00001-service.default"
	err := k.CreateSubscription(subscriptionName, testNS, channelName, &uri)
	assert.Nil(t, err)
}

func Test_DeleteInexistentSubscription(t *testing.T) {
	log.Print("Test_DeleteInexistentSubscription")
	k := &KnativeLib{}
	e := k.InjectClient(evclientsetfake.NewSimpleClientset().EventingV1alpha1(), evclientsetfake.NewSimpleClientset().MessagingV1alpha1())
	assert.Nil(t, e)
	err := k.DeleteSubscription(subscriptionName, testNS)
	assert.True(t, k8serrors.IsNotFound(err))
}

func Test_CreateDeleteSubscription(t *testing.T) {
	log.Print("Test_CreateDeleteSubscription")
	k := &KnativeLib{}
	e := k.InjectClient(evclientsetfake.NewSimpleClientset().EventingV1alpha1(), evclientsetfake.NewSimpleClientset().MessagingV1alpha1())
	assert.Nil(t, e)
	var uri = "dnsName: hello-00001-service.default"
	err := k.CreateSubscription(subscriptionName, testNS, channelName, &uri)
	assert.Nil(t, err)
	err = k.DeleteSubscription(subscriptionName, testNS)
	assert.Nil(t, err)
}

func Test_CreateSubscriptionAgain(t *testing.T) {
	log.Print("Test_CreateSubscriptionAgain")
	k := &KnativeLib{}
	e := k.InjectClient(evclientsetfake.NewSimpleClientset().EventingV1alpha1(), evclientsetfake.NewSimpleClientset().MessagingV1alpha1())
	assert.Nil(t, e)
	var uri = "dnsName: hello-00001-service.default"
	err := k.CreateSubscription(subscriptionName, testNS, channelName, &uri)
	assert.Nil(t, err)
	err = k.CreateSubscription(subscriptionName, testNS, channelName, &uri)
	assert.True(t, k8serrors.IsAlreadyExists(err))
}

func Test_makeHttpRequest(t *testing.T) {
	headers := make(map[string][]string)
	headers["ce-test"] = []string{"test-ce"}
	headers["not-ce-test"] = []string{"test-not-ce"}
	payload := ""
	req, _ := makeHTTPRequest(testChannel, &headers, &payload)
	assert.Equal(t, req.Header["Content-Type"][0], "application/json")
	assert.Equal(t, req.Header["ce-test"][0], "test-ce")
	for k := range req.Header {
		log.Printf("Request Header: %s", k)
	}
	assert.Len(t, req.Header, 3, "Headers map should have exactly 3 keys")
}

func Test_MakeChannelWithPrefix(t *testing.T) {
	prefix := "order.created"
	a := makeChannel(prefix, testNS, labels)

	// makeChannel should rmove all the special characters from the prefix string
	assert.False(t, strings.Contains(a.GenerateName, "."))

	// makeChannel should add hyphen at the end if not present
	assert.True(t, strings.HasSuffix(a.GenerateName, "-"))

	prefix = "order.created-"
	a = makeChannel(prefix, testNS, labels)

	// makeChannel should not add double hyphens if already present
	assert.False(t, strings.HasSuffix(a.GenerateName, "--"))

	//Test prefix length is truncated if the event-type is too big
	prefix = "order.created.on.some.big.enterprise.soljution"
	a = makeChannel(prefix, testNS, labels)
	assert.Equal(t, len(a.GenerateName), 10)

	// Check if prefix is added with a "-" at the end.
	prefix = "order"
	a = makeChannel(prefix, testNS, labels)
	assert.Equal(t, a.GenerateName, "order-")
}
