package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/labels"

	apiv1 "github.com/kyma-project/kyma/components/event-bus/apis/eventing/v1alpha1"

	subApis "github.com/kyma-project/kyma/components/event-bus/apis/eventing/v1alpha1"
	ebClientSet "github.com/kyma-project/kyma/components/event-bus/client/generated/clientset/internalclientset"

	"github.com/avast/retry-go"
	cloudevents "github.com/cloudevents/sdk-go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	eventingclientv1alpha1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/eventing/v1alpha1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

const timeout = time.Second * 30

func SendEvent(target, payload, eventType, eventTypeVersion string) error {
	ctx := context.Background()
	event := cloudevents.NewEvent(cloudevents.VersionV1)
	event.SetType(eventType)
	event.SetDataContentType("text/plain")
	if err := event.SetData(payload); err != nil {
		return err
	}
	event.SetExtension("eventtypeversion", eventTypeVersion)
	event.SetSource("i.will.be.replaced")

	t, err := cloudevents.NewHTTPTransport(cloudevents.WithTarget(target), cloudevents.WithStructuredEncoding())
	if err != nil {
		return err
	}
	c, err := cloudevents.NewClient(t, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	if err != nil {
		return err
	}
	_, _, err = c.Send(ctx, event)
	if err != nil {
		return err
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	request, err := http.NewRequest(http.MethodPost, target, bytes.NewReader(body))
	if err != nil {
		return err
	}

	request.Header.Add("Content-Type", "application/json")

	client := &http.Client{
		Timeout: timeout,
	}

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("send event failed: %v\nrequest: %v\nresponse: %v", response.StatusCode, request, response)
	}

	return nil
}

func CheckEvent(target, eventType, eventTypeVersion string, retryOptions ...retry.Option) error {
	return retry.Do(func() error {
		res, err := http.Get(target)
		if err != nil {
			return fmt.Errorf("get request failed: %v", err)
		}

		if res.StatusCode != 200 {
			return fmt.Errorf("GET request failed: %v", res.StatusCode)
		}
		body, err := ioutil.ReadAll(res.Body)
		defer res.Body.Close()

		if err != nil {
			return err
		}
		var resp http.Header
		err = json.Unmarshal(body, &resp)
		if err != nil {
			return err
		}

		err = res.Body.Close()
		if err != nil {
			return err
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

//func WithBroker(broker string) TriggerOption {
//	return func(trigger *v1alpha1.Trigger) {
//		trigger.Spec.Broker = broker
//	}
//}
//
//func WithRefSubscriber(ref *corev1.ObjectReference) TriggerOption {
//	destination := duckv1.Destination{
//		Ref: ref,
//	}
//	return func(trigger *v1alpha1.Trigger) {
//		trigger.Spec.Subscriber = destination
//	}
//}

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
	trigger := &v1alpha1.Trigger{
		TypeMeta: metav1.TypeMeta{
			Kind:       "trigger",
			APIVersion: "eventing.knative.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
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
				return fmt.Errorf("Trigger with labels: %+v  not found", labelSelector)
			}

			trigger := triggers.Items[0]
			if !trigger.Status.IsReady() {
				return fmt.Errorf("trigger %s n ot ready: %v", trigger.Name, trigger.Status)
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

func CreateSubscription(ebCli ebClientSet.Interface, name, namespace, eventType, srcID string, retryOptions ...retry.Option) error {
	subscriberEventEndpointURL := "http://" + name + "." + namespace + ":9000/v1/events"
	return retry.Do(func() error {
		if _, err := ebCli.EventingV1alpha1().Subscriptions(namespace).Create(NewSubscription(name, namespace, subscriberEventEndpointURL, eventType, "v1", srcID)); err != nil {
			if !strings.Contains(err.Error(), "already exists") {
				return fmt.Errorf("error in creating subscription: %v", err)
			}
		}
		return nil
	}, retryOptions...)
}

func NewSubscription(name string, namespace string, subscriberEventEndpointURL string, eventType string, eventTypeVersion string,
	sourceID string) *apiv1.Subscription {
	//uid, err := uuid.NewV4()
	//if err != nil {
	//	log.Fatalf("Error while generating UID: %v", err)
	//}
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

func CheckSubscriptionReady(ebCli ebClientSet.Interface, name, namespace string, retryOptions ...retry.Option) error {
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
