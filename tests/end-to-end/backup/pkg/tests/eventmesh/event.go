package eventmesh

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/avast/retry-go"
	cloudevents "github.com/cloudevents/sdk-go"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	messaging "knative.dev/eventing/pkg/client/clientset/versioned"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

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
	destination := &duckv1.Destination{
		URI: url,
	}
	return func(trigger *v1alpha1.Trigger) {
		trigger.Spec.Subscriber = destination
	}
}

func WithBroker(broker string) TriggerOption {
	return func(trigger *v1alpha1.Trigger) {
		trigger.Spec.Broker = broker
	}
}

func WithRefSubscriber(ref *corev1.ObjectReference) TriggerOption {
	destination := &duckv1.Destination{
		Ref: ref,
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

func CreateTrigger(messaging messaging.Interface, name, namespace string, triggeroptions ...TriggerOption) error {
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
			Subscriber: &duckv1.Destination{
				Ref: nil,
				URI: nil,
			},
		},
	}

	for _, option := range triggeroptions {
		option(trigger)
	}

	_, err := messaging.EventingV1alpha1().Triggers(namespace).Create(trigger)
	return err
}

func WaitForTrigger(messaging messaging.Interface, name, namespace string, retryOptions ...retry.Option) error {
	return retry.Do(
		func() error {
			trigger, err := messaging.EventingV1alpha1().Triggers(namespace).Get(name, metav1.GetOptions{})
			if err != nil {
				return err
			}

			if !trigger.Status.IsReady() {
				return fmt.Errorf("trigger %s not ready: %v", name, trigger.Status)
			}
			return nil
		}, retryOptions...)
}

func WaitForBroker(messaging messaging.Interface, name, namespace string, retryOptions ...retry.Option) error {
	return retry.Do(
		func() error {
			broker, err := messaging.EventingV1alpha1().Brokers(namespace).Get(name, metav1.GetOptions{})
			if err != nil {
				return err
			}

			if !broker.Status.IsReady() {
				return fmt.Errorf("broker %s not ready: %v", name, broker.Status)
			}
			return nil
		}, retryOptions...)
}
