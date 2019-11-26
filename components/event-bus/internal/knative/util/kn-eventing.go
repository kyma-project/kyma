package util

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"reflect"
	"regexp"
	"sync"
	"time"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/rest"

	messagingV1Alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	evclientset "knative.dev/eventing/pkg/client/clientset/versioned"
	eventingv1alpha1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/eventing/v1alpha1"
	messagingv1alpha1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/messaging/v1alpha1"
	evinformers "knative.dev/eventing/pkg/client/informers/externalversions"
	evlistersv1alpha1 "knative.dev/eventing/pkg/client/listers/messaging/v1alpha1"
)

/*
//
// sample usage of KnativeLib
//
// get a KnativeLib object
k, err := GetKnativeLib()
if err != nil {
	log.Fatalf("Error while getting KnativeLibrary. %v", err)
}
// get a channel
var ch *eventingv1alpha1.Channel
if ch, err = k.GetChannel(channelName, namespace); err != nil && k8serrors.IsNotFound(err) {
	// channel doesn't exist, create it
	if ch, err = k.CreateChannel(provisionerName, channelName, namespace); err != nil {
		log.Printf("ERROR: createChannel() failed: %v", err)
		return
	}
} else if err != nil {
	log.Printf("ERROR: getChannel() failed: %v", err)
	return
}
// send a message to channel
var msg = "test-message"
if err := k.SendMessage(ch, "&msg); err != nil {
	log.Printf("ERROR: sendMessage() failed: %v", err)
	return
}
// create a subscription
var uri = "dnsName: hello-00001-service.default"
if err := k.CreateSubscription("my-sub", namespace, channelName, &uri); err != nil {
	log.Printf("ERROR: create subscription failed: %v", err)
	return
}
return
*/

const (
	generateNameSuffix         = "-"
	maxChannelNamePrefixLength = 10
	informerSyncTimeout        = time.Second * 5
)

var once sync.Once

// KnativeAccessLib encapsulates the Knative access lib behaviours.
type KnativeAccessLib interface {
	GetChannelByLabels(namespace string, labels map[string]string) (*messagingV1Alpha1.Channel, error)
	CreateChannel(prefix, namespace string, labels map[string]string, readyFn ...ChannelReadyFunc) (*messagingV1Alpha1.Channel, error)
	DeleteChannel(name string, namespace string) error
	CreateSubscription(name string, namespace string, channelName string, uri *string, labels map[string]string) error
	DeleteSubscription(name string, namespace string) error
	GetSubscription(name string, namespace string) (*messagingV1Alpha1.Subscription, error)
	UpdateSubscription(sub *messagingV1Alpha1.Subscription) (*messagingV1Alpha1.Subscription, error)
	SendMessage(channel *messagingV1Alpha1.Channel, headers *map[string][]string, message *string) error
	InjectClient(evClient eventingv1alpha1.EventingV1alpha1Interface, msgClient messagingv1alpha1.MessagingV1alpha1Interface) error
	MsgChannelClient() messagingv1alpha1.MessagingV1alpha1Interface
}

// ChannelReadyFunc is a function used to ensure that a Channel has become Ready.
type ChannelReadyFunc func(name string, lister evlistersv1alpha1.ChannelNamespaceLister) error

// WaitForChannelWithTimeout returns a function that waits until a Channel gets
// Ready or fails after a timeout.
func WaitForChannelWithTimeout(timeout time.Duration) ChannelReadyFunc {
	return func(name string, l evlistersv1alpha1.ChannelNamespaceLister) error {
		tout := time.After(timeout)
		tick := time.Tick(100 * time.Millisecond)

		var channel *messagingV1Alpha1.Channel
		var err error

		for ready := false; !ready; {
			select {
			case <-tout:
				return fmt.Errorf("timed out waiting for Channel readiness")
			case <-tick:
				channel, err = l.Get(name)
				if err != nil {
					log.Printf("ERROR: CreateChannel(): getting channel: %v", err)
					continue
				}
				ready = channel.Status.IsReady()
			}
		}

		return nil
	}
}

// NewKnativeLib returns an interface to KnativeLib, which can be mocked
func NewKnativeLib() (*KnativeLib, error) {
	return GetKnativeLib()
}

// KnativeLib represents the knative lib.
type KnativeLib struct {
	evClient   eventingv1alpha1.EventingV1alpha1Interface
	msgClient  messagingv1alpha1.MessagingV1alpha1Interface
	httpClient http.Client
	chLister   evlistersv1alpha1.ChannelLister
}

// Verify the struct KnativeLib implements KnativeLibInterface
var _ KnativeAccessLib = &KnativeLib{}

// GetKnativeLib returns the Knative/Eventing access layer
func GetKnativeLib() (*KnativeLib, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Printf("ERROR: GetChannel(): getting cluster config: %v", err)
		return nil, err
	}
	evClient, err := evclientset.NewForConfig(config)
	if err != nil {
		log.Printf("ERROR: GetChannel(): creating eventing client: %v", err)
		return nil, err
	}

	factory := evinformers.NewSharedInformerFactory(evClient, 0)

	k := &KnativeLib{
		evClient:  evClient.EventingV1alpha1(),
		msgClient: evClient.MessagingV1alpha1(),
		chLister:  factory.Messaging().V1alpha1().Channels().Lister(),
	}
	once.Do(func() {
		k.httpClient = http.Client{
			Transport: initHTTPTransport(),
		}
	})

	// TODO(antoineco): handle termination via k8s lifecycle events
	stop := make(chan struct{})
	factory.Start(stop)
	waitForInformersSyncOrDie(factory)

	return k, nil
}

// waitForInformersSyncOrDie blocks until all informer caches are synced, or panics after a timeout.
func waitForInformersSyncOrDie(f evinformers.SharedInformerFactory) {
	ctx, cancel := context.WithTimeout(context.Background(), informerSyncTimeout)
	defer cancel()

	err := hasSynced(ctx, f.WaitForCacheSync)
	if err != nil {
		log.Fatalf("Error waiting for caches sync: %s", err)
	}
}

type waitForCacheSyncFunc func(stopCh <-chan struct{}) map[reflect.Type]bool

// hasSynced blocks until the given informer sync waiting function completes. It returns an error if the passed context
// gets canceled.
func hasSynced(ctx context.Context, fn waitForCacheSyncFunc) error {
	// synced gets closed as soon as fn returns
	synced := make(chan struct{})

	// closing stopWait forces fn to return, which happens whenever ctx
	// gets canceled
	stopWait := make(chan struct{})
	defer close(stopWait)

	go func() {
		fn(stopWait)
		close(synced)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-synced:
	}

	return nil
}

// MsgChannelClient returns a clientset interface for messaging v1alpha1 API
func (k *KnativeLib) MsgChannelClient() messagingv1alpha1.MessagingV1alpha1Interface {
	return k.msgClient
}

// GetChannelByLabels return a knative channel fetched via label selectors
// so based on the labels, we assume that the list of channels should have only one item in it
// Hence, we'd be returning the item at 0th index.
func (k *KnativeLib) GetChannelByLabels(namespace string, labels map[string]string) (*messagingV1Alpha1.Channel, error) {
	if labels == nil {
		return nil, errors.New("no labels were passed to GetChannelByLabels()")
	}
	channelList, err := k.chLister.Channels(namespace).List(k8slabels.SelectorFromSet(labels))
	if err != nil {
		log.Printf("ERROR: GetChannelByLabels(): getting channels by labels: %v", err)
		return nil, err
	}

	log.Printf("knative channels fetched %v", channelList)

	// ChannelList length should exactly be equal to 1
	if channelListLength := len(channelList); channelListLength != 1 {
		if channelListLength == 0 {
			log.Printf("no channels with the %v labels were found in %v namespace", labels, namespace)
			return nil, k8serrors.NewNotFound(messagingV1Alpha1.Resource("channels"), "")
		}
		log.Printf("ERROR: GetChannelByLabels(): channel list has %d items", channelListLength)
		return nil, errors.New("length of channel list is not equal to 1")
	}

	return channelList[0], nil
}

// CreateChannel creates a Knative/Eventing channel controlled by the specified provisioner
func (k *KnativeLib) CreateChannel(prefix, namespace string, labels map[string]string,
	readyFn ...ChannelReadyFunc) (*messagingV1Alpha1.Channel, error) {

	c := makeChannel(prefix, namespace, labels)
	channel, err := k.msgClient.Channels(namespace).Create(c)
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		log.Printf("ERROR: CreateChannel(): creating channel: %v", err)
		return nil, err
	}

	if readyFn != nil && len(readyFn) == 1 {
		if err := readyFn[0](channel.Name, k.chLister.Channels(namespace)); err != nil {
			return nil, err
		}
	}

	return channel, nil
}

// DeleteChannel deletes a Knative/Eventing channel
func (k *KnativeLib) DeleteChannel(name string, namespace string) error {
	if err := k.msgClient.Channels(namespace).Delete(name, &metav1.DeleteOptions{}); err != nil {
		log.Printf("ERROR: DeleteChannel(): deleting channel: %v", err)
		return err
	}
	return nil
}

// CreateSubscription creates a Knative/Eventing subscription for the specified channel
func (k *KnativeLib) CreateSubscription(name string, namespace string, channelName string, uri *string, labels map[string]string) error {
	sub := Subscription(name, namespace, labels).ToChannel(channelName).ToURI(uri).Build()
	if _, err := k.msgClient.Subscriptions(namespace).Create(sub); err != nil {
		log.Printf("ERROR: CreateSubscription(): creating subscription: %v", err)
		return err
	}
	return nil
}

// DeleteSubscription deletes a Knative/Eventing subscription
func (k *KnativeLib) DeleteSubscription(name string, namespace string) error {
	if err := k.msgClient.Subscriptions(namespace).Delete(name, &metav1.DeleteOptions{}); err != nil {
		log.Printf("ERROR: DeleteSubscription(): deleting subscription: %v", err)
		return err
	}
	return nil
}

// GetSubscription gets a Knative/Eventing subscription
func (k *KnativeLib) GetSubscription(name string, namespace string) (*messagingV1Alpha1.Subscription, error) {
	sub, err := k.msgClient.Subscriptions(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		//log.Printf("ERROR: GetSubscription(): getting subscription: %v", err)
		return nil, err
	}
	return sub, nil
}

// UpdateSubscription updates an existing subscription
func (k *KnativeLib) UpdateSubscription(sub *messagingV1Alpha1.Subscription) (*messagingV1Alpha1.Subscription, error) {
	usub, err := k.msgClient.Subscriptions(sub.Namespace).Update(sub)
	if err != nil {
		log.Printf("ERROR: UpdateSubscription(): updating subscription: %v", err)
		return nil, err
	}
	return usub, nil
}

// SendMessage sends a message to a channel
func (k *KnativeLib) SendMessage(channel *messagingV1Alpha1.Channel, headers *map[string][]string, payload *string) error {

	req, err := makeHTTPRequest(channel, headers, payload)
	if err != nil {
		log.Printf("ERROR: SendMessage(): makeHTTPRequest() failed: %v", err)
		return err
	}

	res, err := k.httpClient.Do(req)
	defer func() {
		if tran, ok := k.httpClient.Transport.(*http.Transport); ok {
			tran.CloseIdleConnections()
		}
	}()
	if err != nil {
		log.Printf("ERROR: SendMessage(): could not send HTTP request: %v", err)
		return err
	}
	defer func() {
		_ = req.Body.Close()
		_ = res.Body.Close()
	}()

	if res.StatusCode == http.StatusNotFound {
		// try to resend the message only once
		if err := resendMessage(&k.httpClient, channel, headers, payload); err != nil {
			log.Printf("ERROR: SendMessage(): resendMessage() failed: %v", err)
			return err
		}
	} else if res.StatusCode != http.StatusAccepted {
		log.Printf("ERROR: SendMessage(): %s", res.Status)
		return errors.New(res.Status)
	}
	// ok
	return nil
}

// InjectClient injects a client, useful for running tests.
func (k *KnativeLib) InjectClient(evClient eventingv1alpha1.EventingV1alpha1Interface, msgClient messagingv1alpha1.MessagingV1alpha1Interface) error {
	k.evClient = evClient
	k.msgClient = msgClient
	return nil
}

func resendMessage(httpClient *http.Client, channel *messagingV1Alpha1.Channel, headers *map[string][]string, message *string) error {
	timeout := time.After(10 * time.Second)
	tick := time.Tick(200 * time.Millisecond)
	req, err := makeHTTPRequest(channel, headers, message)
	if err != nil {
		log.Printf("ERROR: resendMessage(): makeHTTPRequest() failed: %v", err)
		return err
	}
	res, err := httpClient.Do(req)
	defer func() {
		if tran, ok := httpClient.Transport.(*http.Transport); ok {
			tran.CloseIdleConnections()
		}
	}()
	if err != nil {
		log.Printf("ERROR: resendMessage(): could not send HTTP request: %v", err)
		return err
	}
	defer func() {
		_ = res.Body.Close()
	}()
	//dumpResponse(res)
	sc := res.StatusCode
	for sc == http.StatusNotFound {
		select {
		case <-timeout:
			log.Printf("ERROR: resendMessage(): timed out")
			return errors.New("ERROR: timed out")
		case <-tick:
			req, err := makeHTTPRequest(channel, headers, message)
			if err != nil {
				log.Printf("ERROR: resendMessage(): makeHTTPRequest() failed: %v", err)
				return err
			}
			res, err := httpClient.Do(req)
			defer func() {
				if tran, ok := httpClient.Transport.(*http.Transport); ok {
					tran.CloseIdleConnections()
				}
			}()
			if err != nil {
				log.Printf("ERROR: resendMessage(): could not resend HTTP request: %v", err)
				return err
			}
			defer func() {
				_ = res.Body.Close()
			}()
			dumpResponse(res)
			sc = res.StatusCode
		}
	}
	if sc != http.StatusAccepted {
		log.Printf("ERROR: resendMessage(): %v", sc)
		return errors.New(string(sc))
	}
	return nil
}

func makeChannel(prefix, namespace string, labels map[string]string) *messagingV1Alpha1.Channel {
	// Remove all the special characters from the prefix string
	reg, err := regexp.Compile("[^a-z0-9]+")
	if err != nil {
		log.Fatal(err)
	}

	if len(prefix) > maxChannelNamePrefixLength {
		prefix = prefix[:maxChannelNamePrefixLength]
	}
	prefix = reg.ReplaceAllString(prefix, "") + generateNameSuffix

	return &messagingV1Alpha1.Channel{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    namespace,
			GenerateName: prefix,
			Labels:       labels,
		},
	}
}

func makeHTTPRequest(channel *messagingV1Alpha1.Channel, headers *map[string][]string, payload *string) (*http.Request, error) {
	var jsonStr = []byte(*payload)
	channelURI := channel.Status.Address.URL
	req, err := http.NewRequest(http.MethodPost, channelURI.String(), bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Printf("ERROR: makeHTTPRequest(): could not create HTTP request: %v", err)
		return nil, err
	}
	req.Header = *headers
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func initHTTPTransport() *http.Transport {
	return &http.Transport{
		DisableCompression: true,
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
	}
}

func dumpResponse(res *http.Response) {
	dump, err := httputil.DumpResponse(res, true)
	if err != nil {
		log.Printf("ERROR: dumpResponse(): %v", err)
	}
	log.Printf("\n\ndump res1:%s", dump)
}

func getLabelSelectorsAsString(labels map[string]string) string {
	return k8slabels.SelectorFromSet(labels).String()
}
