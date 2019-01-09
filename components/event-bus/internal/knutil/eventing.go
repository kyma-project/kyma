package knutil

import (
	"bytes"
	"errors"
	"crypto/tls"
	"log"
	"net/http"
	"net/http/httputil"
	"time"

	corev1 "k8s.io/api/core/v1"
	evclientset "github.com/knative/eventing/pkg/client/clientset/versioned"
	evapisv1alpha1 "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	eventingv1alpha1 "github.com/knative/eventing/pkg/client/clientset/versioned/typed/eventing/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"

)

/*
//
// sample usage of knutil
//

// get a channel
var ch *eventingv1alpha1.Channel
if ch, err = GetChannel(channelName, namespace); err != nil && k8serrors.IsNotFound(err) {
	// channel doesn't exist, create it
	if ch, err = CreateChannel(provisionerName, channelName, namespace); err != nil {
		log.Printf("ERROR: createChannel() failed: %v", err)
		return
	}
} else if err != nil {
	log.Printf("ERROR: getChannel() failed: %v", err)
	return
}

// send a message to channel
var msg = "test-message"
if err := SendMessage(ch, "&msg); err != nil {
	log.Printf("ERROR: sendMessage() failed: %v", err)
	return
}

// create a subscription
var uri = "dnsName: hello-00001-service.default"
if err := CreateSubscription("my-sub", namespace, channelName, &uri); err != nil {
	log.Printf("ERROR: create subscription failed: %v", err)
	return
}
return
*/


// GetChannel returns an existing Knative/Eventing channel, if it exists.
// If the channel doesn't exist, the error returned can be checked using the
// standard K8S function: "k8serrors.IsNotFound(err) "
// Sample usage:
// if ch, err = GetChannel(channelName, namespace); err != nil && k8serrors.IsNotFound(err) {
//    // channel doesn't exists, must be created first
// } else if err != nil {
//    // other errors
// }
func GetChannel(name string, namespace string) (*evapisv1alpha1.Channel, error) {
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
	if channel, err := evClient.EventingV1alpha1().Channels(namespace).Get(name, metav1.GetOptions{}); err != nil {
		log.Printf("ERROR: GetChannel(): geting channel: %v", err)
		return nil, err
	} else {
		return channel, nil
	}
}

// CreateChannel creates a Knative/Eventing channel controlled by the specified provisioner
func CreateChannel(provisioner string, name string, namespace string, timeout time.Duration) (*evapisv1alpha1.Channel, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Printf("ERROR: CreateChannel(): getting cluster config: %v", err)
		return nil, err
	}
	evClient, err := evclientset.NewForConfig(config)
	if err != nil {
		log.Printf("ERROR: CreateChannel(): creating eventing client: %v", err)
		return nil, err
	}
	return createChannel(evClient.EventingV1alpha1(), provisioner, name, namespace, timeout)
}

// CreateSubscription creates a subscription for the specified channel
func CreateSubscription(name string, namespace string, channelName string, uri *string) error {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Printf("ERROR: CreateSubscription(): getting cluster config: %v", err)
		return err
	}
	evClient, err := evclientset.NewForConfig(config)
	if err != nil {
		log.Printf("ERROR: CreateSubscription(): creating eventing client: %v", err)
		return err
	}
	sub := Subscription(name, namespace).ToChannel(channelName).ToUri(uri).EmptyReply().Build()
	if sub, err = evClient.EventingV1alpha1().Subscriptions(namespace).Create(sub); err != nil && !k8serrors.IsAlreadyExists(err) {
		log.Printf("ERROR: CreateSubscription(): create subscription: %v", err)
		return err
	} else if err != nil && k8serrors.IsAlreadyExists(err) {
		if sub, err = evClient.EventingV1alpha1().Subscriptions(namespace).Update(sub); err != nil {
			log.Printf("ERROR: CreateSubscription(): update subscription: %v", err)
			return err
		}
	}
	return nil
}

// SendMessage sends a message to a channel
func SendMessage(channel *evapisv1alpha1.Channel, message *string) error {
	httpClient := &http.Client{
		Transport: initHTTPTransport(),
	}
	req, err := makeHttpRequest(channel, message)
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
	//dumpResponse(res)
	if res.StatusCode == http.StatusNotFound {
		// try to resend the mesasge only once
		if err := resendMessage(httpClient, channel, message); err != nil {
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

func resendMessage(httpClient *http.Client, channel *evapisv1alpha1.Channel, message *string)  error {
	timeout := time.After(10 * time.Second)
	tick := time.Tick(200 * time.Millisecond)
	req, err := makeHttpRequest(channel, message)
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
	for ; sc == http.StatusNotFound; {
		select {
		case <-timeout:
			log.Printf("ERROR: resendMessage(): timed out")
			return errors.New("ERROR: timed out")
		case <-tick:
			req, err := makeHttpRequest(channel, message)
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

func createChannel(evClient eventingv1alpha1.EventingV1alpha1Interface, provisioner string, name string, namespace string, timeout time.Duration) (*evapisv1alpha1.Channel, error) {
	c := makeChannel(provisioner, name, namespace)
	if channel, err := evClient.Channels(namespace).Create(c); err != nil && !k8serrors.IsAlreadyExists(err) {
		log.Printf("ERROR: createChannel(): creating channel: %v", err)
		return nil, err
	} else {
		isReady := channel.Status.IsReady()
		tout := time.After(timeout)
		tick := time.Tick(100 * time.Millisecond)
		for ; !isReady; {
			select {
			case <-tout:
				return nil, errors.New("timed out")
			case <-tick:
				if channel, err = evClient.Channels(namespace).Get(name, metav1.GetOptions{}); err != nil {
					log.Printf("ERROR: createChannel(): geting channel: %v", err)
				} else {
					isReady = channel.Status.IsReady()
				}
			}
		}
		return channel, nil
	}
}

func makeChannel(provisioner string, name string, namespace string) *evapisv1alpha1.Channel {
	c := &evapisv1alpha1.Channel{
		TypeMeta: metav1.TypeMeta{
			APIVersion: evapisv1alpha1.SchemeGroupVersion.String(),
			Kind:       "Channel",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: evapisv1alpha1.ChannelSpec{
			Provisioner: &corev1.ObjectReference{
				Name: provisioner,
			},
		},
	}
	return c
}

func makeHttpRequest(channel *evapisv1alpha1.Channel, message *string) (*http.Request, error) {
	var jsonStr = []byte(`{"` + *message + `"}`)

	channelUri := "http://" + channel.GetName() + "-channel" + "." + channel.GetNamespace() + ".svc.cluster.local"
	req, err := http.NewRequest(http.MethodPost, channelUri, bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Printf("ERROR: makeHttpRequest(): could not create HTTP request: %v", err)
		return nil, err
	}
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
