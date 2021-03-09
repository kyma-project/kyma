package process

import (
	"fmt"
	"strings"

	knativeapis "knative.dev/pkg/apis"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	kymaeventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	eventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
)

var _ Step = &ConvertTriggersToKymaSubscriptions{}

const (
	eventVersionKey    = "eventtypeversion"
	eventSourceKey     = "source"
	eventTypeKey       = "type"
	clusterLocalSuffix = "svc.cluster.local"
)

type ConvertTriggersToKymaSubscriptions struct {
	name    string
	process *Process
}

func NewConvertTriggersToKymaSubscriptions(p *Process) ConvertTriggersToKymaSubscriptions {
	return ConvertTriggersToKymaSubscriptions{
		name:    "Convert Triggers to Kyma Subscriptions",
		process: p,
	}
}

func (s ConvertTriggersToKymaSubscriptions) Do() error {
	for _, trigger := range s.process.State.Triggers.Items {
		// Generate equivalent subscription for this trigger
		sub := s.NewSubscription(&trigger)

		_, err := s.process.Clients.KymaSubscription.Create(sub)
		if err != nil {
			if k8serrors.IsAlreadyExists(err) {
				// Nothing to do!
				continue
			}
			return errors.Wrapf(err, "failed to create Kyma subscription %s/%s", sub.Namespace, sub.Name)
		}
		s.process.Logger.Infof("Step: %s, converted trigger: %s/%s to subscription", s.ToString(), trigger.Namespace, trigger.Name)
	}
	return nil
}

func (s ConvertTriggersToKymaSubscriptions) NewSubscription(trigger *eventingv1alpha1.Trigger) *kymaeventingv1alpha1.Subscription {
	var bebFilters *kymaeventingv1alpha1.BebFilters
	protocolSettings := new(kymaeventingv1alpha1.ProtocolSettings)
	sink := ""

	var attributes eventingv1alpha1.TriggerFilterAttributes
	var triggerEventTypeVersion, triggerEventSource, triggerEventType string
	if trigger.Spec.Filter != nil {
		attributes = *trigger.Spec.Filter.Attributes
		triggerEventTypeVersion = attributes[eventVersionKey]
		triggerEventSource = attributes[eventSourceKey]
		triggerEventType = attributes[eventTypeKey]
	}
	eventName := build(s.process.EventTypePrefix, triggerEventSource, triggerEventType, triggerEventTypeVersion)

	if trigger.Spec.Filter != nil {
		eventSource := &kymaeventingv1alpha1.Filter{
			Type:     "exact",
			Property: "source",
			Value:    s.process.BEBNamespace,
		}
		eventType := &kymaeventingv1alpha1.Filter{
			Type:     "exact",
			Property: "type",
			Value:    eventName,
		}
		bebFilter := &kymaeventingv1alpha1.BebFilter{
			EventSource: eventSource,
			EventType:   eventType,
		}
		bebFilters = &kymaeventingv1alpha1.BebFilters{
			Dialect: "",
			Filters: []*kymaeventingv1alpha1.BebFilter{
				bebFilter,
			},
		}
	}

	if trigger.Spec.Subscriber.URI != nil {

		sink = rewriteSink(trigger.Spec.Subscriber.URI, trigger.Namespace)
	}
	subscription := &kymaeventingv1alpha1.Subscription{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Subscription",
			APIVersion: "eventing.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            trigger.Name,
			Namespace:       trigger.Namespace,
			Labels:          trigger.Labels,
			OwnerReferences: trigger.OwnerReferences,
		},
		Spec: kymaeventingv1alpha1.SubscriptionSpec{
			Sink:             sink,
			Filter:           bebFilters,
			Protocol:         "",
			ProtocolSettings: protocolSettings,
		},
	}
	return subscription
}

func (s ConvertTriggersToKymaSubscriptions) ToString() string {
	return s.name
}

func build(prefix, applicationName, event, version string) string {
	return fmt.Sprintf("%s.%s.%s.%s", prefix, applicationName, event, version)
}

func rewriteSink(sinkURL *knativeapis.URL, namespace string) string {
	finalURL := *sinkURL
	if sinkURL.Scheme == "" {
		return sinkURL.String()
	}

	baseURLWithoutPort := sinkURL.Host
	portFromBaseURL := ""
	portFromBaseURLSegments := strings.Split(sinkURL.Host, ":")
	baseURLSegments := strings.Split(sinkURL.Host, ".")
	if len(portFromBaseURLSegments) == 2 {
		baseURLWithoutPort = portFromBaseURLSegments[0]
		portFromBaseURL = portFromBaseURLSegments[1]
		baseURLSegments = strings.Split(baseURLWithoutPort, ".")
	}

	if len(baseURLSegments) == 2 {
		ns := baseURLSegments[1]
		if ns == namespace {
			finalURL.Host = fmt.Sprintf("%s.%s", baseURLWithoutPort, clusterLocalSuffix)
			if portFromBaseURL != "" {
				finalURL.Host = fmt.Sprintf("%s:%s", finalURL.Host, portFromBaseURL)
			}
		}
	}

	return finalURL.String()
}
