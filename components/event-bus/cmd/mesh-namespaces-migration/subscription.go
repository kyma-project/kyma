package main

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"

	kneventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	kneventingclientset "knative.dev/eventing/pkg/client/clientset/versioned"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	kymaeventingv1alpha1 "github.com/kyma-project/kyma/components/event-bus/apis/eventing/v1alpha1"
	kymaeventingclientset "github.com/kyma-project/kyma/components/event-bus/client/generated/clientset/internalclientset"
)

// CloudEvents attributes
const (
	sourceAttr           = "source"
	eventTypeAttr        = "type"
	eventTypeVersionAttr = "eventtypeversion"
)

const (
	defaultBrokerName = "default"
	brokerLabel       = "eventing.knative.dev/broker"
)

type triggersByNamespaceMap map[string][]kneventingv1alpha1.Trigger
type subscriptionsList []kymaeventingv1alpha1.Subscription

// subscriptionMigrator performs migrations of Kyma Subscriptions to Knative Triggers.
type subscriptionMigrator struct {
	kymaClient    kymaeventingclientset.Interface
	knativeClient kneventingclientset.Interface

	subscriptions       subscriptionsList
	triggersByNamespace triggersByNamespaceMap
}

// newSubscriptionMigrator creates and initializes a subscriptionMigrator.
func newSubscriptionMigrator(kymaClient kymaeventingclientset.Interface,
	knativeClient kneventingclientset.Interface, namespaces []string) (*subscriptionMigrator, error) {

	m := &subscriptionMigrator{
		kymaClient:    kymaClient,
		knativeClient: knativeClient,
	}

	if err := m.populateSubscriptions(namespaces); err != nil {
		return nil, err
	}
	if err := m.populateTriggers(namespaces); err != nil {
		return nil, err
	}

	return m, nil
}

// populateSubscription populates the local list of Kyma Subscriptions.
func (m *subscriptionMigrator) populateSubscriptions(namespaces []string) error {
	var kymaSubscriptions []kymaeventingv1alpha1.Subscription

	for _, ns := range namespaces {
		subs, err := m.kymaClient.EventingV1alpha1().Subscriptions(ns).List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrapf(err, "listing Subscriptions in namespace %s", ns)
		}
		kymaSubscriptions = append(kymaSubscriptions, subs.Items...)
	}

	m.subscriptions = kymaSubscriptions

	return nil
}

// populateTriggers populates the local triggersByNamespace map.
func (m *subscriptionMigrator) populateTriggers(namespaces []string) error {
	triggersByNamespace := make(triggersByNamespaceMap)

	for _, ns := range namespaces {
		triggers, err := m.knativeClient.EventingV1alpha1().Triggers(ns).List(metav1.ListOptions{})
		if err != nil {
			return errors.Wrapf(err, "listing Triggers in namespace %s", ns)
		}
		triggersByNamespace[ns] = triggers.Items
	}

	m.triggersByNamespace = triggersByNamespace

	return nil
}

// migrateAll migrates all Kyma Subscriptions listed in the subscriptionMigrator.
func (m *subscriptionMigrator) migrateAll() error {
	log.Printf("Starting migration of %d Subscriptions", len(m.subscriptions))

	for _, sub := range m.subscriptions {
		subKey := fmt.Sprintf("%s/%s", sub.Namespace, sub.Name)

		if err := m.migrateSubscription(sub); err != nil {
			return errors.Wrapf(err, "migrating Subscription %q", subKey)
		}

		log.Printf("+ Deleting Subscription %q", subKey)

		if err := m.deleteSubscriptionWithRetry(sub.Namespace, sub.Name); err != nil {
			return errors.Wrapf(err, "deleting Subscription %q", subKey)
		}
	}

	return nil
}

// deleteSubscriptionWithRetry deletes a Subscription and retries in case of failure.
func (m *subscriptionMigrator) deleteSubscriptionWithRetry(ns, name string) error {
	var expectSuccessfulSubscriptionDeletion wait.ConditionFunc = func() (bool, error) {
		err := m.kymaClient.EventingV1alpha1().Subscriptions(ns).Delete(name, &metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return false, nil
		}
		return true, nil
	}

	return wait.PollImmediateUntil(5*time.Second, expectSuccessfulSubscriptionDeletion, newTimeoutChannel())
}

// migrateSubscription migrates a single Kyma Subscription.
func (m *subscriptionMigrator) migrateSubscription(sub kymaeventingv1alpha1.Subscription) error {
	subKey := fmt.Sprintf("%s/%s", sub.Namespace, sub.Name)

	log.Printf("+ Checking Subscription %q", subKey)
	if m.findTriggerForSubscription(sub) != nil {
		log.Printf("+ Trigger already exists for Subscription %q, skipping", subKey)
		return nil
	}

	log.Printf("+ Trigger not found for Subscription %q", subKey)
	trigger, err := m.createTriggerForSubscription(sub)
	if err != nil {
		return errors.Wrapf(err, "creating Trigger for Subscription %q", subKey)
	}
	log.Printf("+ Trigger \"%s/%s\" created for Subscription %q", trigger.Namespace, trigger.Name, subKey)

	return nil
}

// findTriggerForSubscription returns a Trigger containing the same attributes as the given Kyma Subscription if it
// exists in the subscriptionMigrator.
func (m *subscriptionMigrator) findTriggerForSubscription(sub kymaeventingv1alpha1.Subscription) *kneventingv1alpha1.Trigger {
	triggersInNamespace := m.triggersByNamespace[sub.Namespace]
	if triggersInNamespace == nil {
		return nil
	}

	for _, tr := range triggersInNamespace {
		if tr.Spec.Filter.Attributes == nil {
			continue
		}
		trAttributes := *tr.Spec.Filter.Attributes

		if sub.Endpoint == tr.Spec.Subscriber.URI.String() &&
			sub.SourceID == trAttributes[sourceAttr] &&
			sub.EventType == trAttributes[eventTypeAttr] &&
			sub.EventTypeVersion == trAttributes[eventTypeVersionAttr] {

			return &tr
		}
	}

	return nil
}

// createTriggerForSubscription creates a Knative Trigger equivalent to the given Kyma Subscription.
func (m *subscriptionMigrator) createTriggerForSubscription(sub kymaeventingv1alpha1.Subscription) (*kneventingv1alpha1.Trigger, error) {
	subKey := fmt.Sprintf("%s/%s", sub.Namespace, sub.Name)

	tr, err := newTriggerForSubscription(sub)
	if err != nil {
		return nil, errors.Wrapf(err, "generating Trigger object for Subscription %q", subKey)
	}

	if tr, err = m.knativeClient.EventingV1alpha1().Triggers(sub.Namespace).Create(tr); err != nil {
		return nil, err
	}

	return tr, nil
}

// newTriggerForSubscription returns a Knative Trigger object equivalent to the given Kyma Subscription.
func newTriggerForSubscription(sub kymaeventingv1alpha1.Subscription) (*kneventingv1alpha1.Trigger, error) {
	endpointURL, err := apis.ParseURL(sub.Endpoint)
	if err != nil {
		return nil, errors.Wrapf(err, "parsing URL %v", sub.Endpoint)
	}

	trLabels := make(labels.Set, len(sub.Labels)+1)
	for k, v := range sub.Labels {
		trLabels[k] = v
	}
	trLabels[brokerLabel] = defaultBrokerName

	trAttributes := kneventingv1alpha1.TriggerFilterAttributes{
		sourceAttr:           sub.SourceID,
		eventTypeAttr:        sub.EventType,
		eventTypeVersionAttr: sub.EventTypeVersion,
	}

	return &kneventingv1alpha1.Trigger{
		ObjectMeta: metav1.ObjectMeta{
			Name:      uuid.New().String(),
			Namespace: sub.Namespace,
			Labels:    trLabels,
		},
		Spec: kneventingv1alpha1.TriggerSpec{
			Broker: defaultBrokerName,
			Filter: &kneventingv1alpha1.TriggerFilter{
				Attributes: &trAttributes,
			},
			Subscriber: duckv1.Destination{
				URI: endpointURL,
			},
		},
	}, nil
}
