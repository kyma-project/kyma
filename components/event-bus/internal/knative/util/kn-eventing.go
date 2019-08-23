package util

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"time"

	evapisv1alpha1 "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	messagingV1Alpha1 "github.com/knative/eventing/pkg/apis/messaging/v1alpha1"
	clientsetChannel "github.com/knative/eventing/pkg/client/clientset/versioned"
	evclientset "github.com/knative/eventing/pkg/client/clientset/versioned"
	eventingv1alpha1 "github.com/knative/eventing/pkg/client/clientset/versioned/typed/eventing/v1alpha1"
	informersChannel "github.com/knative/eventing/pkg/client/informers/externalversions"
	informerChannel "github.com/knative/eventing/pkg/client/informers/externalversions/messaging/v1alpha1"

	messagingv1alpha1 "github.com/knative/eventing/pkg/client/clientset/versioned/typed/messaging/v1alpha1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
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
var uri = "dnsName: hello-00001-service.CreateChannel"
if err := k.CreateSubscription("my-sub", namespace, channelName, &uri); err != nil {
	log.Printf("ERROR: create subscription failed: %v", err)
	return
}
return
*/

// KnativeAccessLib encapsulates the Knative access lib behaviours.
type KnativeAccessLib interface {
	GetChannel(name string, namespace string) (*evapisv1alpha1.Channel, error)
	CreateChannel(provisioner string, name string, namespace string, labels *map[string]string,
		timeout time.Duration) (*messagingV1Alpha1.Channel, error)
	DeleteChannel(name string, namespace string) error
	CreateSubscription(name string, namespace string, channelName string, uri *string) error
	CreateNatssChannelSubscription(name string, namespace string, channelName string, uri *string) error
	CreateGPubSubChannelSubscription(name string, namespace string, channelName string, uri *string) error
	DeleteSubscription(name string, namespace string) error
	GetSubscription(name string, namespace string) (*evapisv1alpha1.Subscription, error)
	UpdateSubscription(sub *evapisv1alpha1.Subscription) (*evapisv1alpha1.Subscription, error)
	SendMessage(channel *evapisv1alpha1.Channel, headers *map[string][]string, message *string) error
	InjectClient(c eventingv1alpha1.EventingV1alpha1Interface) error
	// GetNatssChannel(name string, namespace string) (*v1alpha1.NatssChannel, error)
	GetMessagingChannel(name string, namespace string) (*messagingV1Alpha1.Channel, error)
	CreateMessagingChannel(name string, namespace string, labels *map[string]string,
		timeout time.Duration) (*messagingV1Alpha1.Channel, error)
}

// NewKnativeLib returns an interface to KnativeLib, which can be mocked
func NewKnativeLib() (KnativeAccessLib, error) {
	return GetKnativeLib()
}

// KnativeLib represents the knative lib.
type KnativeLib struct {
	evClient         eventingv1alpha1.EventingV1alpha1Interface
	messagingChannel messagingv1alpha1.MessagingV1alpha1Interface
}

// Verify the struct KnativeLib implements KnativeLibIntf
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
	// todo refactor
	// init natssChannelInformer ///////////////////////////////////////////////////////////////////////////////////////
	stopCh := make(<-chan struct{})
	resyncPeriod := 30 * time.Second
	var messagingChannelInformer informerChannel.ChannelInformer
	// var natssChannelInformer v1alpha1Natss.NatssChannelInformer
	// if natssClient := clientsetNatss.NewForConfigOrDie(config); natssClient != nil {
	// 	messagingInformerFactory := informersNatss.NewSharedInformerFactory(natssClient, resyncPeriod)
	// 	natssChannelInformer = messagingInformerFactory.Messaging().V1alpha1().NatssChannels()
	// 	// v1alpha1Natss.
	// 	if err := startInformers(stopCh, natssChannelInformer.Informer()); err != nil {
	// 		log.Printf("failed to start natssChannelInformer %v", err)
	// 		return nil, err
	// 	}
	// }
	if channelClientset := clientsetChannel.NewForConfigOrDie(config); channelClientset != nil {
		messagingInformerFactory := informersChannel.NewSharedInformerFactory(channelClientset, resyncPeriod)
		messagingChannelInformer = messagingInformerFactory.Messaging().V1alpha1().Channels()
		// v1alpha1Natss.
		if err := startInformers(stopCh, messagingChannelInformer.Informer()); err != nil {
			log.Printf("failed to start messaging channel informer %v", err)
			return nil, err
		}
	}
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	k := &KnativeLib{
		// natssChannelInformer: natssChannelInformer,
		evClient:         evClient.EventingV1alpha1(),
		messagingChannel: evClient.MessagingV1alpha1(),
	}
	return k, nil
}

// GetChannel returns an existing Knative/Eventing channel, if it exists.
// If the channel doesn't exist, the error returned can be checked using the
// standard K8S function: "k8serrors.IsNotFound(err) "
func (k *KnativeLib) GetChannel(name string, namespace string) (*evapisv1alpha1.Channel, error) {
	channel, err := k.evClient.Channels(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		//log.Printf("ERROR: GetChannel(): getting channel: %v", err)
		return nil, err
	}
	if !channel.Status.IsReady() {
		return nil, fmt.Errorf("ERROR: GetChannel():channel NotReady")
	}
	return channel, nil
}

// GetNatssChannel todo
// func (k *KnativeLib) GetNatssChannel(name string, namespace string) (*v1alpha1.NatssChannel, error) {
// 	channel, err := k.natssChannelInformer.Lister().NatssChannels(namespace).Get(name)
// 	if err != nil {
// 		log.Printf("error: GetNatssChannel(): getting channel: %v", err)
// 		return nil, err
// 	}
// 	if !channel.Status.IsReady() {
// 		return nil, fmt.Errorf("error: GetNatssChannel():channel NotReady")
// 	}
// 	return channel, nil
// }

// GetMessagingChannel TODO
func (k *KnativeLib) GetMessagingChannel(name string, namespace string) (*messagingV1Alpha1.Channel, error) {
	channel, err := k.messagingChannel.Channels(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		log.Printf("error: Channel(): getting channel: %v", err)
		return nil, err
	}
	if !channel.Status.IsReady() {
		return nil, fmt.Errorf("error: GetMessagingChannel():channel NotReady")
	}
	return channel, nil
}

// CreateChannel creates a Knative/Eventing channel controlled by the specified provisioner
func (k *KnativeLib) CreateChannel(provisioner string, name string, namespace string, labels *map[string]string,
	timeout time.Duration) (*messagingV1Alpha1.Channel, error) {
	c := makeChannel(name, namespace, labels)
	channel, err := k.messagingChannel.Channels(namespace).Create(c)
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		log.Printf("ERROR: CreateChannel(): creating channel: %v", err)
		return nil, err
	}

	isReady := channel.Status.IsReady()
	tout := time.After(timeout)
	tick := time.Tick(100 * time.Millisecond)
	for !isReady {
		select {
		case <-tout:
			return nil, errors.New("timed out")
		case <-tick:
			if channel, err = k.messagingChannel.Channels(namespace).Get(name, metav1.GetOptions{}); err != nil {
				log.Printf("ERROR: CreateChannel(): geting channel: %v", err)
			} else {
				isReady = channel.Status.IsReady()
			}
		}
	}
	return channel, nil
}

// CreateMessagingChannel TODO
func (k *KnativeLib) CreateMessagingChannel(name string, namespace string, labels *map[string]string,
	timeout time.Duration) (*messagingV1Alpha1.Channel, error) {
	c := makeChannel(name, namespace, labels)
	channel, err := k.messagingChannel.Channels(namespace).Create(c)
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		log.Printf("ERROR: CreateChannel(): creating channel: %v", err)
		return nil, err
	}

	isReady := channel.Status.IsReady()
	tout := time.After(timeout)
	tick := time.Tick(100 * time.Millisecond)
	for !isReady {
		select {
		case <-tout:
			return nil, errors.New("timed out")
		case <-tick:
			if channel, err = k.messagingChannel.Channels(namespace).Get(name, metav1.GetOptions{}); err != nil {
				log.Printf("ERROR: CreateChannel(): geting channel: %v", err)
			} else {
				isReady = channel.Status.IsReady()
			}
		}
	}
	return channel, nil
}

// DeleteChannel deletes a Knative/Eventing channel
func (k *KnativeLib) DeleteChannel(name string, namespace string) error {
	if err := k.evClient.Channels(namespace).Delete(name, &metav1.DeleteOptions{}); err != nil {
		log.Printf("ERROR: DeleteChannel(): deleting channel: %v", err)
		return err
	}
	return nil
}

// CreateSubscription creates a Knative/Eventing subscription for the specified channel
func (k *KnativeLib) CreateSubscription(name string, namespace string, channelName string, uri *string) error {
	sub := Subscription(name, namespace).ToChannel(channelName).ToURI(uri).EmptyReply().Build()
	if _, err := k.evClient.Subscriptions(namespace).Create(sub); err != nil {
		log.Printf("ERROR: CreateSubscription(): creating subscription: %v", err)
		return err
	}
	return nil
}

// CreateNatssChannelSubscription todo
func (k *KnativeLib) CreateNatssChannelSubscription(name string, namespace string, channelName string, uri *string) error {
	sub := NatssChannelSubscription(name, namespace).ToNatssChannel(channelName).ToURI(uri).EmptyReply().Build()
	if _, err := k.evClient.Subscriptions(namespace).Create(sub); err != nil {
		log.Printf("ERROR: CreateNatssChannelSubscription(): creating subscription: %v", err)
		return err
	}
	return nil
}

// CreateGPubSubChannelSubscription todo
func (k *KnativeLib) CreateGPubSubChannelSubscription(name string, namespace string, channelName string, uri *string) error {
	sub := GPubSubChannelSubscription(name, namespace).ToGPubSubChannel(channelName).ToURI(uri).EmptyReply().Build()
	if _, err := k.evClient.Subscriptions(namespace).Create(sub); err != nil {
		log.Printf("ERROR: CreateGPubSubChannelSubscription(): creating subscription: %v", err)
		return err
	}
	return nil
}

// DeleteSubscription deletes a Knative/Eventing subscription
func (k *KnativeLib) DeleteSubscription(name string, namespace string) error {
	if err := k.evClient.Subscriptions(namespace).Delete(name, &metav1.DeleteOptions{}); err != nil {
		log.Printf("ERROR: DeleteSubscription(): deleting subscription: %v", err)
		return err
	}
	return nil
}

// GetSubscription gets a Knative/Eventing subscription
func (k *KnativeLib) GetSubscription(name string, namespace string) (*evapisv1alpha1.Subscription, error) {
	sub, err := k.evClient.Subscriptions(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		//log.Printf("ERROR: GetSubscription(): getting subscription: %v", err)
		return nil, err
	}
	return sub, nil
}

// UpdateSubscription updates an existing subscription
func (k *KnativeLib) UpdateSubscription(sub *evapisv1alpha1.Subscription) (*evapisv1alpha1.Subscription, error) {
	usub, err := k.evClient.Subscriptions(sub.Namespace).Update(sub)
	if err != nil {
		log.Printf("ERROR: UpdateSubscription(): updating subscription: %v", err)
		return nil, err
	}
	return usub, nil
}

// SendMessage sends a message to a channel
func (k *KnativeLib) SendMessage(channel *evapisv1alpha1.Channel, headers *map[string][]string, payload *string) error {
	httpClient := &http.Client{
		Transport: initHTTPTransport(),
	}
	req, err := makeHTTPRequest(channel, headers, payload)
	if err != nil {
		log.Printf("ERROR: SendMessage(): makeHTTPRequest() failed: %v", err)
		return err
	}

	res, err := httpClient.Do(req)
	if err != nil {
		log.Printf("ERROR: SendMessage(): could not send HTTP request: %v", err)
		return err
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode == http.StatusNotFound {
		// try to resend the message only once
		if err := resendMessage(httpClient, channel, headers, payload); err != nil {
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

// SendMessageToNatssChannel todo
// func (k *KnativeLib) SendMessageToNatssChannel(channel *v1alpha1.NatssChannel, headers *map[string][]string, payload *string) error {
// 	httpClient := &http.Client{
// 		Transport: initHTTPTransport(),
// 	}
// 	req, err := makeHTTPRequestToNatssChannel(channel, headers, payload)
// 	if err != nil {
// 		log.Printf("ERROR: SendMessage(): makeHTTPRequest() failed: %v", err)
// 		return err
// 	}

// 	res, err := httpClient.Do(req)
// 	if err != nil {
// 		log.Printf("ERROR: SendMessage(): could not send HTTP request: %v", err)
// 		return err
// 	}
// 	defer func() {
// 		_ = res.Body.Close()
// 	}()

// 	if res.StatusCode == http.StatusNotFound {
// 		// try to resend the message only once
// 		if err := resendMessageToNatssChannel(httpClient, channel, headers, payload); err != nil {
// 			log.Printf("ERROR: SendMessage(): resendMessage() failed: %v", err)
// 			return err
// 		}
// 	} else if res.StatusCode != http.StatusAccepted {
// 		log.Printf("ERROR: SendMessage(): %s", res.Status)
// 		return errors.New(res.Status)
// 	}
// 	// ok
// 	return nil
// }

// SendMessageToChannel TODO
func (k *KnativeLib) SendMessageToChannel(channel *messagingV1Alpha1.Channel, headers *map[string][]string, payload *string) error {
	httpClient := &http.Client{
		Transport: initHTTPTransport(),
	}
	req, err := makeHTTPRequestToMessagingChannel(channel, headers, payload)
	if err != nil {
		log.Printf("ERROR: SendMessage(): makeHTTPRequest() failed: %v", err)
		return err
	}

	res, err := httpClient.Do(req)
	if err != nil {
		log.Printf("ERROR: SendMessage(): could not send HTTP request: %v", err)
		return err
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode == http.StatusNotFound {
		// try to resend the message only once
		if err := resendMessageToNatssChannel(httpClient, channel, headers, payload); err != nil {
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
func (k *KnativeLib) InjectClient(c eventingv1alpha1.EventingV1alpha1Interface) error {
	k.evClient = c
	return nil
}

type informer interface {
	Run(<-chan struct{})
	HasSynced() bool
}

func startInformers(stopCh <-chan struct{}, informers ...informer) error {
	for _, informer := range informers {
		informer := informer
		go informer.Run(stopCh)
	}

	for i, informer := range informers {
		if ok := cache.WaitForCacheSync(stopCh, informer.HasSynced); !ok {
			return fmt.Errorf("failed to wait for cache at index %d to sync", i)
		}
	}
	return nil
}

func resendMessage(httpClient *http.Client, channel *evapisv1alpha1.Channel, headers *map[string][]string, message *string) error {
	timeout := time.After(10 * time.Second)
	tick := time.Tick(200 * time.Millisecond)
	req, err := makeHTTPRequest(channel, headers, message)
	if err != nil {
		log.Printf("ERROR: resendMessage(): makeHTTPRequest() failed: %v", err)
		return err
	}
	res, err := httpClient.Do(req)
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
			if err != nil {
				log.Printf("ERROR: resendMessage(): could not resend HTTP request: %v", err)
				return err
			}
			defer func() { _ = res.Body.Close() }()
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

func resendMessageToNatssChannel(httpClient *http.Client, channel *messagingV1Alpha1.Channel, headers *map[string][]string, message *string) error {
	timeout := time.After(10 * time.Second)
	tick := time.Tick(200 * time.Millisecond)
	req, err := makeHTTPRequestToMessagingChannel(channel, headers, message)
	if err != nil {
		log.Printf("ERROR: resendMessage(): makeHTTPRequest() failed: %v", err)
		return err
	}
	res, err := httpClient.Do(req)
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
			req, err := makeHTTPRequestToMessagingChannel(channel, headers, message)
			if err != nil {
				log.Printf("ERROR: resendMessage(): makeHTTPRequest() failed: %v", err)
				return err
			}
			res, err := httpClient.Do(req)
			if err != nil {
				log.Printf("ERROR: resendMessage(): could not resend HTTP request: %v", err)
				return err
			}
			defer func() { _ = res.Body.Close() }()
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

func makeChannel(name string, namespace string, labels *map[string]string) *messagingV1Alpha1.Channel {
	c := &messagingV1Alpha1.Channel{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			Labels:    *labels,
		},
	}
	return c
}

func makeHTTPRequest(channel *evapisv1alpha1.Channel, headers *map[string][]string, payload *string) (*http.Request, error) {
	var jsonStr = []byte(*payload)

	channelURI := "http://" + channel.Status.Address.Hostname
	req, err := http.NewRequest(http.MethodPost, channelURI, bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Printf("ERROR: makeHTTPRequest(): could not create HTTP request: %v", err)
		return nil, err
	}
	req.Header = *headers
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func makeHTTPRequestToMessagingChannel(channel *messagingV1Alpha1.Channel, headers *map[string][]string, payload *string) (*http.Request, error) {
	var jsonStr = []byte(*payload)

	channelURI := "http://" + channel.Status.Address.Hostname
	req, err := http.NewRequest(http.MethodPost, channelURI, bytes.NewBuffer(jsonStr))
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
