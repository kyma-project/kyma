package knutil

import (
	"log"
	"net/http"
	"net/http/httputil"
	"time"

	eventingv1alpha1 "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"

	evclientset "github.com/knative/eventing/pkg/client/clientset/versioned"
	"bytes"
	"errors"
	"crypto/tls"
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



func GetChannel(name string, namespace string) (*eventingv1alpha1.Channel, error) {
	log.Printf("GetChannel() name: %v; namespace: %v", name, namespace)
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Printf("ERROR: getting cluster config: %v", err)
		return nil, err
	}
	evClient, err := evclientset.NewForConfig(config)
	if err != nil {
		log.Printf("ERROR: creating eventing client: %v", err)
		return nil, err
	}
	if channel, err := evClient.EventingV1alpha1().Channels(namespace).Get(name, metav1.GetOptions{}); err != nil {
		log.Printf("ERROR: geting channel: %v", err)
		return nil, err
	} else {
		return channel, nil
	}
}

func CreateChannel(provisioner string, name string, namespace string, timeout time.Duration) (*eventingv1alpha1.Channel, error) {
	log.Printf("createChannel() provisioner: %v; name: %v; namespace: %v; timeout: %v", provisioner, name, namespace, timeout)
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Printf("ERROR: getting cluster config: %v", err)
		return nil, err
	}
	evClient, err := evclientset.NewForConfig(config)
	if err != nil {
		log.Printf("ERROR: creating eventing client: %v", err)
		return nil, err
	}
	// create the channel object
	c := makeChannel(provisioner, name, namespace)
	log.Printf("channel c: %+v", c)
	if channel, err := evClient.EventingV1alpha1().Channels(namespace).Create(c); err != nil && !k8serrors.IsAlreadyExists(err) {
		log.Printf("ERROR: creating channel: %v", err)
		return nil, err
	} else {
		isReady := channel.Status.IsReady()
		tout := time.After(timeout) // 5 * time.Second
		tick := time.Tick(100 * time.Millisecond)
		for ; !isReady; {
			select {
			case <-tout:
				return nil, errors.New("timed out")
			case <-tick:
				if channel, err = evClient.EventingV1alpha1().Channels(namespace).Get(name, metav1.GetOptions{}); err != nil {
					log.Printf("ERROR: geting channel: %v", err)
				} else {
					isReady = channel.Status.IsReady()
				}
			}
		}
		return channel, nil
	}
}

// create subscription only for test now...
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
	log.Printf("Subscription created: %v", sub)
	if sub, err = evClient.EventingV1alpha1().Subscriptions(namespace).Create(sub); err != nil && !k8serrors.IsAlreadyExists(err) {
		log.Printf("ERROR: CreateSubscription(): create subscription: %v", err)
		return err
	} else if err != nil && k8serrors.IsAlreadyExists(err) {
		if sub, err = evClient.EventingV1alpha1().Subscriptions(namespace).Update(sub); err != nil {
			log.Printf("ERROR: CreateSubscription(): update subscription: %v", err)
			return err
		}
	}

	log.Printf("Subscription created: %v", sub)
	return nil
}


func SendMessage(channel *eventingv1alpha1.Channel, message *string) error {
	log.Printf("SendMessage() channel: %v; message: %v", channel, message)
	httpClient := &http.Client{
		Transport: initHTTPTransport(),
	}
	req, err := makeHttpRequest(channel, message)
	if err != nil {
		log.Printf("ERROR: makeHttpRequest() failed: %v", err)
		return err
	}
	res, err := httpClient.Do(req)
	if err != nil {
		log.Printf("ERROR: SendMessage() could not send HTTP request: %v", err)
		return err
	}
	defer res.Body.Close()
	dumpResponse(res)

	// try to resend the message to the channel, if necessary
	if res.StatusCode == http.StatusNotFound {
		// try to resend the mesasge
		if err := resendMessage(httpClient, channel, message); err != nil {
			log.Printf("ERROR: SendMessage(): resendMessage() failed: %v", err)
			return err
		}
	} else if res.StatusCode != http.StatusAccepted {
		log.Printf("ERROR: %s", res.Status)
		return errors.New(res.Status)
	}
	// ok
	return nil
}

func resendMessage(httpClient *http.Client, channel *eventingv1alpha1.Channel, message *string)  error {
	timeout := time.After(10 * time.Second)
	tick := time.Tick(200 * time.Millisecond)

	req, err := makeHttpRequest(channel, message)
	if err != nil {
		log.Printf("ERROR: makeHttpRequest() failed: %v", err)
		return err
	}
	res, err := httpClient.Do(req)
	if err != nil {
		log.Printf("ERROR: resendMessage() could not send HTTP request: %v", err)
		return err
	}
	defer res.Body.Close()
	dumpResponse(res)

	sc := res.StatusCode
	for ; sc == http.StatusNotFound; {
		select {
		case <-timeout:
			log.Printf("ERROR: timed out")
			return errors.New("ERROR: timed out")
		case <-tick:
			req, err := makeHttpRequest(channel, message)
			if err != nil {
				log.Printf("ERROR: makeHttpRequest() failed: %v", err)
				return err
			}
			res, err := httpClient.Do(req)
			if err != nil {
				log.Printf("ERROR: resendMessage() could not resend HTTP request: %v", err)
				return err
			}
			defer res.Body.Close()
			dumpResponse(res)
			sc = res.StatusCode
		}
	}
	if sc != http.StatusAccepted {
		log.Printf("ERROR: %v", sc)
		return errors.New(string(sc))
	}
	return nil
}

func makeChannel(provisioner string, name string, namespace string) *eventingv1alpha1.Channel {
	c := &eventingv1alpha1.Channel{
		TypeMeta: metav1.TypeMeta{
			APIVersion: eventingv1alpha1.SchemeGroupVersion.String(),
			Kind:       "Channel",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: eventingv1alpha1.ChannelSpec{
			Provisioner: &corev1.ObjectReference{
				Name: provisioner,
			},
		},
	}
	return c
}

func makeHttpRequest(channel *eventingv1alpha1.Channel, message *string) (*http.Request, error) {
	var jsonStr = []byte(`{"` + *message + `"}`)

	channelUri := "http://" + channel.GetName() + "-channel" + "." + channel.GetNamespace() + ".svc.cluster.local"
	req, err := http.NewRequest(http.MethodPost, channelUri, bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Printf("ERROR: SendMessage() could not create HTTP request: %v", err)
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
