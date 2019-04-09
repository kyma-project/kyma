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
	evclientset "github.com/knative/eventing/pkg/client/clientset/versioned"
	eventingv1alpha1 "github.com/knative/eventing/pkg/client/clientset/versioned/typed/eventing/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
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

// Use it cause it should be mocked
type KnativeAccessLib interface {
	GetChannel(name string, namespace string) (*evapisv1alpha1.Channel, error)
	CreateChannel(provisioner string, name string, namespace string, labels *map[string]string,
		timeout time.Duration)(*evapisv1alpha1.Channel, error)
	DeleteChannel(name string, namespace string) error
	CreateSubscription(name string, namespace string, channelName string, uri *string) error
	DeleteSubscription(name string, namespace string) error
	GetSubscription(name string, namespace string) (*evapisv1alpha1.Subscription, error)
	UpdateSubscription(sub *evapisv1alpha1.Subscription) (*evapisv1alpha1.Subscription, error)
	SendMessage(channel *evapisv1alpha1.Channel, headers *map[string]string, message *string) error
	InjectClient(c eventingv1alpha1.EventingV1alpha1Interface) error
}

// NewKnativeLib returns an interface to KnativeLib, which can be mocked
func NewKnativeLib() (KnativeAccessLib, error) {
	return GetKnativeLib()
}

type KnativeLib struct {
	evClient eventingv1alpha1.EventingV1alpha1Interface
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
	k := &KnativeLib{
		evClient: evClient.EventingV1alpha1(),
	}
	return k, nil
}

// GetChannel returns an existing Knative/Eventing channel, if it exists.
// If the channel doesn't exist, the error returned can be checked using the
// standard K8S function: "k8serrors.IsNotFound(err) "
func (k *KnativeLib) GetChannel(name string, namespace string) (*evapisv1alpha1.Channel, error) {
	if channel, err := k.evClient.Channels(namespace).Get(name, metav1.GetOptions{}); err != nil {
		//log.Printf("ERROR: GetChannel(): getting channel: %v", err)
		return nil, err
	} else {
		if !channel.Status.IsReady() {
			return nil, fmt.Errorf("ERROR: GetChannel():channel NotReady")
		}
		return channel, nil
	}
}

// CreateChannel creates a Knative/Eventing channel controlled by the specified provisioner
func (k *KnativeLib) CreateChannel(provisioner string, name string, namespace string, labels *map[string]string,
	timeout time.Duration)(*evapisv1alpha1.Channel, error) {
	c := makeChannel(provisioner, name, namespace, labels)
	if channel, err := k.evClient.Channels(namespace).Create(c); err != nil && !k8serrors.IsAlreadyExists(err) {
		log.Printf("ERROR: CreateChannel(): creating channel: %v", err)
		return nil, err
	} else {
		isReady := channel.Status.IsReady()
		tout := time.After(timeout)
		tick := time.Tick(100 * time.Millisecond)
		for !isReady {
			select {
			case <-tout:
				return nil, errors.New("timed out")
			case <-tick:
				if channel, err = k.evClient.Channels(namespace).Get(name, metav1.GetOptions{}); err != nil {
					log.Printf("ERROR: CreateChannel(): geting channel: %v", err)
				} else {
					isReady = channel.Status.IsReady()
				}
			}
		}
		return channel, nil
	}
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
	sub := Subscription(name, namespace).ToChannel(channelName).ToUri(uri).EmptyReply().Build()
	if _, err := k.evClient.Subscriptions(namespace).Create(sub); err != nil {
		log.Printf("ERROR: CreateSubscription(): creating subscription: %v", err)
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
	if sub, err := k.evClient.Subscriptions(namespace).Get(name, metav1.GetOptions{}); err != nil {
		//log.Printf("ERROR: GetSubscription(): getting subscription: %v", err)
		return nil, err
	} else {
		return sub, nil
	}
}

// UpdateSubscription updates an existing subscription
func (k *KnativeLib) UpdateSubscription(sub *evapisv1alpha1.Subscription) (*evapisv1alpha1.Subscription, error) {
	if usub, err := k.evClient.Subscriptions(sub.Namespace).Update(sub); err != nil {
		log.Printf("ERROR: UpdateSubscription(): updating subscription: %v", err)
		return nil, err
	} else {
		return usub, nil
	}
}

// SendMessage sends a message to a channel
func (k *KnativeLib) SendMessage(channel *evapisv1alpha1.Channel, headers *map[string]string, payload *string) error {
	httpClient := &http.Client{
		Transport: initHTTPTransport(),
	}
	req, err := makeHttpRequest(channel, headers, payload)
	if err != nil {
		log.Printf("ERROR: SendMessage(): makeHttpRequest() failed: %v", err)
		return err
	}

	res, err := httpClient.Do(req)
	if err != nil {
		log.Printf("ERROR: SendMessage(): could not send HTTP request: %v", err)
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		// try to resend the mesasge only once
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

// InjectClient injects a client, useful for running tests.
func (r *KnativeLib) InjectClient(c eventingv1alpha1.EventingV1alpha1Interface) error {
	r.evClient = c
	return nil
}

func resendMessage(httpClient *http.Client, channel *evapisv1alpha1.Channel, headers *map[string]string, message *string) error {
	timeout := time.After(10 * time.Second)
	tick := time.Tick(200 * time.Millisecond)
	req, err := makeHttpRequest(channel, headers, message)
	if err != nil {
		log.Printf("ERROR: resendMessage(): makeHttpRequest() failed: %v", err)
		return err
	}
	res, err := httpClient.Do(req)
	if err != nil {
		log.Printf("ERROR: resendMessage(): could not send HTTP request: %v", err)
		return err
	}
	defer res.Body.Close()
	//dumpResponse(res)
	sc := res.StatusCode
	for sc == http.StatusNotFound {
		select {
		case <-timeout:
			log.Printf("ERROR: resendMessage(): timed out")
			return errors.New("ERROR: timed out")
		case <-tick:
			req, err := makeHttpRequest(channel, headers, message)
			if err != nil {
				log.Printf("ERROR: resendMessage(): makeHttpRequest() failed: %v", err)
				return err
			}
			res, err := httpClient.Do(req)
			if err != nil {
				log.Printf("ERROR: resendMessage(): could not resend HTTP request: %v", err)
				return err
			}
			defer res.Body.Close()
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

func makeChannel(provisioner string, name string, namespace string, labels *map[string]string) *evapisv1alpha1.Channel {
	c := &evapisv1alpha1.Channel{
		TypeMeta: metav1.TypeMeta{
			APIVersion: evapisv1alpha1.SchemeGroupVersion.String(),
			Kind:       "Channel",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			Labels:    *labels,
		},
		Spec: evapisv1alpha1.ChannelSpec{
			Provisioner: &corev1.ObjectReference{
				Name: provisioner,
			},
		},
	}
	return c
}

func makeHttpRequest(channel *evapisv1alpha1.Channel, headers *map[string]string, payload *string) (*http.Request, error) {
	var jsonStr = []byte(*payload)

	channelUri := "http://" + channel.Status.Address.Hostname
	req, err := http.NewRequest(http.MethodPost, channelUri, bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Printf("ERROR: makeHttpRequest(): could not create HTTP request: %v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	for k, v := range *headers {
		req.Header.Set(k, v)
	}

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
