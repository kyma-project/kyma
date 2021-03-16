package helpers

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/pkg/errors"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"

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
		return errors.Wrap(err, "HTTP POST request failed in SendEvent()")
	}

	if err := verifyStatusCode(res, 200); err != nil {
		return errors.Wrap(err, "HTTP POST request returned non-2xx failed in SendEvent()")
	}

	return nil
}

func CheckEvent(target string, statusCode int, retryOptions ...retry.Option) error {
	return retry.Do(func() error {
		transport := &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: timeout,
			}).DialContext,
		}
		client := http.Client{Transport: transport}
		res, err := client.Get(target)

		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("HTTP GET failed in CheckEvent() for target: %v", target))
		}

		if err := verifyStatusCode(res, statusCode); err != nil {
			return errors.Wrap(err, fmt.Sprintf("HTTP GET request returned non-200 in CheckEvent() for target: %v", target))
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
		return func(trigger *v1alpha1.Trigger) {}
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

func RemoveBrokerInjectionLabel(k8sInf kubernetes.Interface, namespace string, retryOptions ...retry.Option) error {
	return retry.Do(
		func() error {
			ns, err := k8sInf.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{})
			if err != nil {
				return err
			}
			delete(ns.Labels, "knative-eventing-injection")
			_, err = k8sInf.CoreV1().Namespaces().Update(ns)
			if err != nil {
				return err
			}
			return nil
		}, retryOptions...)
}

func DeleteBroker(eventingInf eventingclientv1alpha1.EventingV1alpha1Interface, name, namespace string, retryOptions ...retry.Option) error {
	err := retry.Do(
		func() error {
			err := eventingInf.Brokers(namespace).Delete(name, &metav1.DeleteOptions{})
			if err != nil {
				return err
			}
			return nil
		}, retryOptions...)

	if err != nil {
		return err
	}

	return retry.Do(
		func() error {
			broker, err := eventingInf.Brokers(namespace).Get(name, metav1.GetOptions{})
			if k8serrors.IsNotFound(err) {
				return nil
			}
			if err != nil {
				return err
			}
			if broker.Name == name {
				return fmt.Errorf("broker: %s still exists in ns: %s", name, namespace)
			}
			return nil
		}, retryOptions...)
}

// Verify that the http response has the given status code and return an error if not
func verifyStatusCode(res *http.Response, expectedStatusCode int) error {
	if res.StatusCode != expectedStatusCode {
		return fmt.Errorf("status code is wrong, have: %d, want: %d", res.StatusCode, expectedStatusCode)
	}
	return nil
}
