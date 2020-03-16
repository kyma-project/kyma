package helpers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/avast/retry-go"

	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	apiv1 "github.com/kyma-project/kyma/components/event-bus/apis/eventing/v1alpha1"
	subApis "github.com/kyma-project/kyma/components/event-bus/apis/eventing/v1alpha1"
	ebclientset "github.com/kyma-project/kyma/components/event-bus/client/generated/clientset/internalclientset"

	"knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	eventingclientv1alpha1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/eventing/v1alpha1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

const timeout = time.Second * 30

func SendEvent(target, eventType, eventTypeVersion string) error {
	log.Printf("Sending an event to target: %s with eventType: %s, eventTypeVersion: %s", target, eventType, eventTypeVersion)
	payload := fmt.Sprintf(
		`{"event-type":"%s","event-type-version": "%s","event-time":"2018-11-02T22:08:41+00:00","data":"foo"}`, eventType, eventTypeVersion)
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: timeout,
		}).DialContext,
	}
	client := http.Client{Transport: transport}
	res, err := client.Post(target,
		"application/json",
		strings.NewReader(payload))
	if err != nil {
		return errors.Wrap(err, "HTTP POST request failed in SendEvent() ")
	}

	if err := verifyStatusCode(res, 200); err != nil {
		return errors.Wrap(err, "HTTP POST request returned non-2xx failed in SendEvent() ")
	}

	return nil
}

func CheckEvent(target, eventType, eventTypeVersion string, retryOptions ...retry.Option) error {
	return retry.Do(func() error {
		transport := &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: timeout,
			}).DialContext,
		}
		client := http.Client{Transport: transport}
		res, err := client.Get(target)
		if err != nil {
			return errors.Wrap(err, "HTTP GET failed in CheckEvent()")
		}

		if err := verifyStatusCode(res, 200); err != nil {
			return errors.Wrap(err, "HTTP GET request returned non-2xx in CheckEvent()")
		}
		body, err := ioutil.ReadAll(res.Body)
		defer res.Body.Close()

		if err != nil {
			errors.Wrap(err, "failed ReadAll() in CheckEvent")
		}
		var resp http.Header
		err = json.Unmarshal(body, &resp)
		if err != nil {
			errors.Wrap(err, "failed Unmarshal() in CheckEvent")
		}

		err = res.Body.Close()
		if err != nil {
			errors.Wrap(err, "failed Close() CheckEvent")
		}

		if resp.Get("ce-type") != eventType {
			return fmt.Errorf("event type wrong: Got %s, Wanted: %s", resp.Get("ce-type"), eventType)
		}
		if resp.Get("ce-eventtypeversion") != eventTypeVersion {
			return fmt.Errorf("event type version wrong: Got %s, Wanted: %s", resp.Get("ce-eventtypeversion"), eventTypeVersion)
		}

		return nil
	}, retryOptions...)

	return nil
}

type TriggerOption func(trigger *v1alpha1.Trigger)

func WithURISubscriber(target string) TriggerOption {
	url, err := apis.ParseURL(target)
	if err != nil {
		// TODO(k15r): proper error handling here
		return func(trigger *v1alpha1.Trigger) {
			return
		}
	}
	destination := duckv1.Destination{
		URI: url,
	}
	return func(trigger *v1alpha1.Trigger) {
		trigger.Spec.Subscriber = destination
	}
}

func WithFilter(eventTypeVersion, eventType, source string) TriggerOption {
	filter := v1alpha1.TriggerFilter{
		Attributes: &v1alpha1.TriggerFilterAttributes{
			"eventtypeversion": eventTypeVersion,
			"type":             eventType,
			"source":           source,
		},
	}

	return func(trigger *v1alpha1.Trigger) {
		trigger.Spec.Filter = &filter
	}
}

func CreateTrigger(eventingCli eventingclientv1alpha1.EventingV1alpha1Interface, name, namespace string, triggeroptions ...TriggerOption) error {
	labelSelector := map[string]string{
		"function": name,
	}
	trigger := &v1alpha1.Trigger{
		TypeMeta: metav1.TypeMeta{
			Kind:       "trigger",
			APIVersion: "eventing.knative.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labelSelector,
		},
		Spec: v1alpha1.TriggerSpec{
			Broker: "default",
			Filter: &v1alpha1.TriggerFilter{
				DeprecatedSourceAndType: nil,
				Attributes:              nil,
			},
			Subscriber: duckv1.Destination{
				Ref: nil,
				URI: nil,
			},
		},
	}

	for _, option := range triggeroptions {
		option(trigger)
	}

	_, err := eventingCli.Triggers(namespace).Create(trigger)
	return err
}

func WaitForTrigger(eventingCli eventingclientv1alpha1.EventingV1alpha1Interface, subName, namespace string, retryOptions ...retry.Option) error {
	return retry.Do(
		func() error {
			labelSelector := map[string]string{
				"function": subName,
			}
			listOptions := metav1.ListOptions{LabelSelector: labels.SelectorFromSet(labelSelector).String()}
			triggers, err := eventingCli.Triggers(namespace).List(listOptions)
			if err != nil {
				return err
			}
			if len(triggers.Items) == 0 {
				return fmt.Errorf("trigger with labels: %+v  not found", labelSelector)
			}

			trigger := triggers.Items[0]
			if !trigger.Status.IsReady() {
				return fmt.Errorf("trigger %s not ready: %v", trigger.Name, trigger.Status)
			}
			return nil
		}, retryOptions...)
}

func WaitForBroker(eventingCli eventingclientv1alpha1.EventingV1alpha1Interface, name, namespace string, retryOptions ...retry.Option) error {
	return retry.Do(
		func() error {
			broker, err := eventingCli.Brokers(namespace).Get(name, metav1.GetOptions{})
			if err != nil {
				return err
			}

			if !broker.Status.IsReady() {
				return fmt.Errorf("broker %s not ready: %v", name, broker.Status)
			}
			return nil
		}, retryOptions...)
}

func CreateSubscription(ebCli ebclientset.Interface, name, namespace, eventType, eventVersion, srcID string, retryOptions ...retry.Option) error {
	subscriberEventEndpointURL := "http://" + name + "." + namespace + ".svc.cluster.local:9000/v3/events"
	return retry.Do(func() error {
		if _, err := ebCli.EventingV1alpha1().Subscriptions(namespace).Create(NewSubscription(name, namespace, subscriberEventEndpointURL, eventType, eventVersion, srcID)); err != nil {
			if !strings.Contains(err.Error(), "already exists") {
				return fmt.Errorf("error in creating subscription: %v", err)
			}
		}
		return nil
	}, retryOptions...)
}

func NewSubscription(name string, namespace string, subscriberEventEndpointURL string, eventType string, eventTypeVersion string,
	sourceID string) *apiv1.Subscription {
	labelsSelector := map[string]string{
		"function": name,
	}
	return &apiv1.Subscription{
		TypeMeta: metav1.TypeMeta{APIVersion: apiv1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labelsSelector,
		},

		SubscriptionSpec: apiv1.SubscriptionSpec{
			Endpoint:                      subscriberEventEndpointURL,
			IncludeSubscriptionNameHeader: false,
			SourceID:                      sourceID,
			EventType:                     eventType,
			EventTypeVersion:              eventTypeVersion,
		},
	}
}

func CheckSubscriptionReady(ebCli ebclientset.Interface, name, namespace string, retryOptions ...retry.Option) error {
	activatedCondition := subApis.SubscriptionCondition{Type: subApis.Ready, Status: subApis.ConditionTrue}
	return retry.Do(func() error {
		kySub, err := ebCli.EventingV1alpha1().Subscriptions(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("cannot get Kyma subscription, name: %v; namespace: %v", name, namespace)
		}
		if kySub.HasCondition(activatedCondition) {
			return nil
		}
		return fmt.Errorf("subscription %v does not have condition %+v", kySub, activatedCondition)
	}, retryOptions...)
}

// Verify that the http response has the given status code and return an error if not
func verifyStatusCode(res *http.Response, expectedStatusCode int) error {
	if res.StatusCode != expectedStatusCode {
		return fmt.Errorf("status code is wrong, have: %d, want: %d", res.StatusCode, expectedStatusCode)
	}
	return nil
}
