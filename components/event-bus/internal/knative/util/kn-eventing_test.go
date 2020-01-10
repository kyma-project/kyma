package util

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"

	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck/v1alpha1"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"

	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	evclientset "knative.dev/eventing/pkg/client/clientset/versioned"
	evclientsetfake "knative.dev/eventing/pkg/client/clientset/versioned/fake"
	evinformers "knative.dev/eventing/pkg/client/informers/externalversions"
	evlistersv1alpha1 "knative.dev/eventing/pkg/client/listers/messaging/v1alpha1"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8stesting "k8s.io/client-go/testing"
)

const (
	channelName      = "test-channel"
	testNS           = "test-namespace"
	subscriptionName = "test-subscription"
)

var testChannel = &messagingv1alpha1.Channel{
	ObjectMeta: metav1.ObjectMeta{
		Namespace:    testNS,
		GenerateName: "testchann-", // as generated from channelName by makeChannel()
		Labels: map[string]string{
			"l1": "v1",
			"l2": "v2",
		},
	},
}

var labels = map[string]string{
	"l1": "v1",
	"l2": "v2",
}

var testChannelList = &messagingv1alpha1.ChannelList{
	Items: []messagingv1alpha1.Channel{*testChannel},
}

func Test_CreateChannel(t *testing.T) {
	client := evclientsetfake.NewSimpleClientset()

	k, stop := newKnativeLib(client, t)
	defer close(stop)

	ch, err := k.CreateChannel(channelName, testNS, labels)
	assert.Nil(t, err)
	log.Printf("Channel created: %v", ch)

	ignore := cmpopts.IgnoreTypes(apis.VolatileTime{})
	if diff := cmp.Diff(testChannel, ch, ignore); diff != "" {
		t.Errorf("(-want, +got) = %v", diff)
	}
}

func Test_GetChannelByLabels(t *testing.T) {
	log.Print("Creating Channel to fetch")

	client := evclientsetfake.NewSimpleClientset()

	/*
	   FIXME(antoineco): if the object was handled, fake.Invokes() returns
	   immediately and the object never makes it to the tracker. As a result, the
	   informer's cache doesn't get updated.
	   See https://github.com/kubernetes/client-go/issues/500
	   Fixed in client-go v11.0.0 (k8s 1.13) via https://github.com/kubernetes/kubernetes/pull/73601
	*/
	/*
		client.Fake.PrependReactor("create", "channels", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			obj := action.(k8stesting.CreateAction).GetObject()
			ch := obj.(*messagingv1alpha1.Channel)

			ch.Status = messagingv1alpha1.ChannelStatus{
				Status: duckv1beta1.Status{
					Conditions: duckv1beta1.Conditions{
						apis.Condition{
							Type:   messagingv1alpha1.ChannelConditionReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
			}

			return false, ch, nil
		})
	*/

	k, stop := newKnativeLib(client, t)
	defer close(stop)

	ch1, err1 := k.CreateChannel(channelName, testNS, labels)
	assert.Nil(t, err1)
	t.Logf("Channel created: %v", ch1)

	time.Sleep(100 * time.Millisecond)

	t.Log("Getting Channel by label")
	ch2, err2 := k.GetChannelByLabels(testNS, labels)

	assert.Nil(t, err2)

	ignore := cmpopts.IgnoreTypes(apis.VolatileTime{})
	if diff := cmp.Diff(ch1, ch2, ignore); diff != "" {
		t.Errorf("(-want, +got) = %v", diff)
	}
}

func Test_CreateChannelWithError(t *testing.T) {
	client := evclientsetfake.NewSimpleClientset()

	k, stop := newKnativeLib(client, t)
	defer close(stop)

	/*
	   FIXME(antoineco): replace with Reactor as soon as we update to client-go v11.0.0 (k8s 1.13)
	   See https://github.com/kubernetes/client-go/issues/500
	   and https://github.com/kubernetes/kubernetes/pull/73601
	*/
	setErrorWaitForChannelFunc := ChannelReadyFunc(func(name string, l evlistersv1alpha1.ChannelNamespaceLister) error {
		getFunc := k.msgClient.Channels(testNS).Get
		updFunc := k.msgClient.Channels(testNS).Update
		ch, _ := getFunc(name, metav1.GetOptions{})
		ch.Labels["l1"] = "not-matching"
		_, err := updFunc(ch)
		return err
	})

	ch, err := k.CreateChannel(channelName, testNS, labels, setErrorWaitForChannelFunc)
	assert.Nil(t, err)
	log.Printf("Channel created: %v", ch)

	time.Sleep(100 * time.Millisecond)

	t.Log("Getting Channel by label")
	_, err = k.GetChannelByLabels(testNS, labels)
	assert.EqualError(t, err, `channels.messaging.knative.dev "" not found`)

}

func Test_CreateChannelTimeout(t *testing.T) {
	client := evclientsetfake.NewSimpleClientset()

	k, stop := newKnativeLib(client, t)
	defer close(stop)

	_, err := k.CreateChannel(channelName, testNS, labels, WaitForChannelWithTimeout(10*time.Millisecond))

	assert.EqualError(t, err, "timed out waiting for Channel readiness")
}

func Test_SendMessage(t *testing.T) {
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
	ch, err := k.CreateChannel(channelName, testNS, labels)
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
	evClient := evclientsetfake.NewSimpleClientset().EventingV1alpha1()
	msgClient := evclientsetfake.NewSimpleClientset().MessagingV1alpha1()
	k := &KnativeLib{}
	err := k.InjectClient(evClient, msgClient)
	assert.Nil(t, err)
}

func Test_DeleteInexistentChannel(t *testing.T) {
	evClient := evclientsetfake.NewSimpleClientset().EventingV1alpha1()
	msgClient := evclientsetfake.NewSimpleClientset().MessagingV1alpha1()
	k := &KnativeLib{}
	e := k.InjectClient(evClient, msgClient)
	assert.Nil(t, e)
	err := k.DeleteChannel(channelName, testNS)
	assert.True(t, k8serrors.IsNotFound(err))
}

func Test_CreateDeleteChannel(t *testing.T) {
	client := evclientsetfake.NewSimpleClientset()
	client.Fake.ReactionChain = nil
	client.Fake.AddReactor("create", "channels", func(action k8stesting.Action) (handled bool,
		ret runtime.Object, err error) {
		return true, testChannel, nil
	})
	k := &KnativeLib{}
	e := k.InjectClient(client.EventingV1alpha1(), client.MessagingV1alpha1())
	assert.Nil(t, e)
	ch, err := k.CreateChannel(channelName, testNS, labels)
	assert.Nil(t, err)
	err = k.DeleteChannel(ch.Name, ch.Namespace)
	assert.Nil(t, err)
}

func Test_CreateSubscription(t *testing.T) {
	k := &KnativeLib{}
	e := k.InjectClient(evclientsetfake.NewSimpleClientset().EventingV1alpha1(), evclientsetfake.NewSimpleClientset().MessagingV1alpha1())
	assert.Nil(t, e)
	var uri = "dnsName: hello-00001-service.default"
	err := k.CreateSubscription(subscriptionName, testNS, channelName, &uri, labels)
	assert.Nil(t, err)
}

func Test_DeleteInexistentSubscription(t *testing.T) {
	k := &KnativeLib{}
	e := k.InjectClient(evclientsetfake.NewSimpleClientset().EventingV1alpha1(), evclientsetfake.NewSimpleClientset().MessagingV1alpha1())
	assert.Nil(t, e)
	err := k.DeleteSubscription(subscriptionName, testNS)
	assert.True(t, k8serrors.IsNotFound(err))
}

func Test_CreateDeleteSubscription(t *testing.T) {
	k := &KnativeLib{}
	e := k.InjectClient(evclientsetfake.NewSimpleClientset().EventingV1alpha1(), evclientsetfake.NewSimpleClientset().MessagingV1alpha1())
	assert.Nil(t, e)
	var uri = "dnsName: hello-00001-service.default"
	err := k.CreateSubscription(subscriptionName, testNS, channelName, &uri, labels)
	assert.Nil(t, err)
	err = k.DeleteSubscription(subscriptionName, testNS)
	assert.Nil(t, err)
}

func Test_CreateSubscriptionAgain(t *testing.T) {
	k := &KnativeLib{}
	e := k.InjectClient(evclientsetfake.NewSimpleClientset().EventingV1alpha1(), evclientsetfake.NewSimpleClientset().MessagingV1alpha1())
	assert.Nil(t, e)
	var uri = "dnsName: hello-00001-service.default"
	err := k.CreateSubscription(subscriptionName, testNS, channelName, &uri, labels)
	assert.Nil(t, err)
	err = k.CreateSubscription(subscriptionName, testNS, channelName, &uri, labels)
	assert.True(t, k8serrors.IsAlreadyExists(err))
}

func Test_MakeHttpRequest(t *testing.T) {
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

	// makeChannel should remove all the special characters from the prefix string
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

func newKnativeLib(client evclientset.Interface, t *testing.T) (*KnativeLib, chan struct{}) {
	t.Helper()

	factory := evinformers.NewSharedInformerFactory(client, 0)

	factory.Messaging().V1alpha1().Channels().Informer()
	kl := &KnativeLib{
		evClient:  client.EventingV1alpha1(),
		msgClient: client.MessagingV1alpha1(),
		chLister:  factory.Messaging().V1alpha1().Channels().Lister(),
	}

	stopCh := make(chan struct{})
	factory.Start(stopCh)
	factory.WaitForCacheSync(stopCh)

	return kl, stopCh
}

func TestHasSynced(t *testing.T) {
	testCases := map[string]struct {
		syncFunc  waitForCacheSyncFunc
		timeout   time.Duration
		expectErr bool
	}{
		"Succeeds with real informers": {
			syncFunc:  newFakeFactory(t).WaitForCacheSync,
			timeout:   time.Second * 2,
			expectErr: false,
		},
		"Fails when timeout is exceeded": {
			// dummy function that blocks forever to ensure we time
			// out and fail
			syncFunc:  func(<-chan struct{}) map[reflect.Type]bool { select {} },
			timeout:   time.Millisecond * 10,
			expectErr: true,
		},
	}

	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			defer cancel()

			err := hasSynced(ctx, tc.syncFunc)

			if tc.expectErr && err == nil {
				t.Error("Expected an error but got none")
			}
			if !tc.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// newFakeFactory returns an informer factory initialized with a fake client.
func newFakeFactory(t *testing.T) evinformers.SharedInformerFactory {
	t.Helper()

	factory := evinformers.NewSharedInformerFactory(evclientsetfake.NewSimpleClientset(), 0)
	// request channels informer
	factory.Messaging().V1alpha1().Channels().Informer()

	return factory
}
