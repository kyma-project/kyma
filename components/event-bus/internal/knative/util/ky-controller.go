package util

import (
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"go.uber.org/zap"

	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiLabels "k8s.io/apimachinery/pkg/labels"

	evapisv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	messagingv1alpha1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/messaging/v1alpha1"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/event-bus/apis/applicationconnector/v1alpha1"
	subApis "github.com/kyma-project/kyma/components/event-bus/apis/eventing/v1alpha1"
	applicationconnectorclientv1alpha1 "github.com/kyma-project/kyma/components/event-bus/client/generated/clientset/internalclientset/typed/applicationconnector/v1alpha1"
	eventingclientv1alpha1 "github.com/kyma-project/kyma/components/event-bus/client/generated/clientset/internalclientset/typed/eventing/v1alpha1"
	applicationconnectorlistersv1alpha1 "github.com/kyma-project/kyma/components/event-bus/client/generated/lister/applicationconnector/v1alpha1"
)

// ContainsString returns true if the string exists in the array.
func ContainsString(slice *[]string, s string) bool {
	for _, item := range *slice {
		if item == s {
			return true
		}
	}
	return false
}

// RemoveString removes the string from in the array and returns a new array instance.
func RemoveString(slice *[]string, s string) (result []string) {
	for _, item := range *slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

// UpdateEventActivation handles Kyma EventActivation
func UpdateEventActivation(client applicationconnectorclientv1alpha1.ApplicationconnectorV1alpha1Interface, ea *eventingv1alpha1.EventActivation) error {
	currentEA, err := client.EventActivations(ea.Namespace).Get(ea.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if !equality.Semantic.DeepEqual(currentEA.Finalizers, ea.Finalizers) {
		ea.SetFinalizers(ea.ObjectMeta.Finalizers)
		if _, err := client.EventActivations(ea.Namespace).Update(ea); err != nil {
			return err
		}
	}
	return nil
}

// UpdateKnativeSubscription updates Knative subscription on change in Finalizer
func UpdateKnativeSubscription(client messagingv1alpha1.MessagingV1alpha1Interface, subscription *evapisv1alpha1.Subscription) error {
	//objectKey := runtimeClient.ObjectKey{Namespace: subscription.Namespace, Name: subscription.Name}
	//sub := &evapisv1alpha1.Subscription{}
	currentSub, err := client.Subscriptions(subscription.Namespace).Get(subscription.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if !equality.Semantic.DeepEqual(currentSub.Finalizers, subscription.Finalizers) {
		subscription.SetFinalizers(subscription.ObjectMeta.Finalizers)
		if _, err := client.Subscriptions(subscription.Namespace).Update(subscription); err != nil {
			return err
		}
	}
	return nil
}

// GetKymaSubscriptionForSubscription gets Kyma Subscription for a particular Knative Subscription
func GetKymaSubscriptionForSubscription(client eventingclientv1alpha1.EventingV1alpha1Interface, knSub *evapisv1alpha1.Subscription) (*subApis.Subscription, error) {
	var chNamespace string
	if _, ok := knSub.Labels[SubNs]; ok {
		chNamespace = knSub.Labels[SubNs]
	}
	sl, err := client.Subscriptions(chNamespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var kymaSub *subApis.Subscription
	for _, s := range sl.Items {
		if doesSubscriptionMatchLabels(&s, knSub.Labels) {
			kymaSub = &s
			break
		}
	}

	return kymaSub, nil
}

// CheckIfEventActivationExistForSubscription returns a boolean value indicating if there is an EventActivation for
// the Subscription or not.
func CheckIfEventActivationExistForSubscription(eventActivationLister applicationconnectorlistersv1alpha1.EventActivationLister, sub *subApis.Subscription) bool {
	subNamespace := sub.GetNamespace()
	subSourceID := sub.SourceID

	//eal := &eventingv1alpha1.EventActivationList{}
	//lo := &runtimeClient.ListOptions{
	//	Namespace: subNamespace,
	//	Raw: &metav1.ListOptions{ // TODO this is here because the fake client needs it. Remove this when it's no longer needed.
	//		TypeMeta: metav1.TypeMeta{
	//			APIVersion: eventingv1alpha1.SchemeGroupVersion.String(),
	//			Kind:       "EventActivation",
	//		},
	//	},
	//}
	//lo := metav1.ListOptions{ // TODO this is here because the fake client needs it. Remove this when it's no longer needed.
	//	TypeMeta: metav1.TypeMeta{
	//		APIVersion: eventingv1alpha1.SchemeGroupVersion.String(),
	//		Kind:       "EventActivation",
	//	},
	//}
	eal, err := eventActivationLister.EventActivations(subNamespace).List(apiLabels.Everything())
	if err != nil {
		return false
	}
	for _, ea := range eal {
		if subSourceID == ea.SourceID && ea.DeletionTimestamp.IsZero() {
			return true
		}
	}
	return false
}

// GetSubscriptionsForEventActivation gets the "ea" object of all the subscriptions having
// the same "namespace" and the same "Source"
func GetSubscriptionsForEventActivation(client eventingclientv1alpha1.EventingV1alpha1Interface, ea *eventingv1alpha1.EventActivation) ([]*subApis.Subscription, error) {
	eaNamespace := ea.GetNamespace()
	eaSourceID := ea.EventActivationSpec.SourceID

	sl, err := client.Subscriptions(eaNamespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var subs []*subApis.Subscription
	for _, s := range sl.Items {
		if eaSourceID == s.SourceID {
			subs = append(subs, &s)
		}
	}
	return subs, nil
}

// CurrentTime handles current time
type CurrentTime interface {
	GetCurrentTime() metav1.Time
}

// DefaultCurrentTime represents the default current time
type DefaultCurrentTime struct{}

// NewDefaultCurrentTime returns a new CurrentTime instance
func NewDefaultCurrentTime() CurrentTime {
	return new(DefaultCurrentTime)
}

// GetCurrentTime returns the current time.
func (t *DefaultCurrentTime) GetCurrentTime() metav1.Time {
	return metav1.NewTime(time.Now())
}

// SubscriptionWithError handles Kyma subscriptions
type SubscriptionWithError struct {
	Sub *subApis.Subscription
	Err error
}

// WriteSubscriptions writes subscriptions.
func WriteSubscriptions(client eventingclientv1alpha1.EventingV1alpha1Interface, subs []*subApis.Subscription) []SubscriptionWithError {
	var errorSubs []SubscriptionWithError
	for _, u := range subs {
		if err := WriteSubscription(client, u); err != nil {
			errorSubs = append(errorSubs, SubscriptionWithError{Sub: u, Err: err})
		}
	}
	return errorSubs
}

// WriteSubscription writes a subscription.
func WriteSubscription(client eventingclientv1alpha1.EventingV1alpha1Interface, sub *subApis.Subscription) error {
	var err error

	// update the subscription status sub-resource
	_, err = client.Subscriptions(sub.Namespace).UpdateStatus(sub)
	if err != nil {
		return err
	}

	// update the subscription resource
	_, err = client.Subscriptions(sub.Namespace).Update(sub)
	if err != nil {
		return err
	}

	return nil
}

// SetReadySubscription set subscription as ready.
func SetReadySubscription(client eventingclientv1alpha1.EventingV1alpha1Interface, sub *subApis.Subscription, msg string, time CurrentTime) error {
	us := updateSubscriptionReadyStatus(sub, subApis.ConditionTrue, msg, time)
	return WriteSubscription(client, us)
}

// SetNotReadySubscription set subscription as not ready.
func SetNotReadySubscription(client eventingclientv1alpha1.EventingV1alpha1Interface, sub *subApis.Subscription, time CurrentTime) error {
	us := updateSubscriptionReadyStatus(sub, subApis.ConditionFalse, "", time)
	return WriteSubscription(client, us)
}

// IsSubscriptionActivated checks if the subscription is active or not.
func IsSubscriptionActivated(sub *subApis.Subscription) bool {
	eventActivatedCondition := subApis.SubscriptionCondition{Type: subApis.EventsActivated, Status: subApis.ConditionTrue}
	knSubReadyCondition := subApis.SubscriptionCondition{Type: subApis.SubscriptionReady, Status: subApis.ConditionTrue}
	return sub.HasCondition(eventActivatedCondition) && sub.HasCondition(knSubReadyCondition)

}

// ActivateSubscriptions activates subscriptions.
func ActivateSubscriptions(client eventingclientv1alpha1.EventingV1alpha1Interface, subs []*subApis.Subscription, log *zap.Logger, time CurrentTime) error {
	updatedSubs := updateSubscriptionsEventActivatedStatus(subs, subApis.ConditionTrue, time)
	return updateSubscriptions(client, updatedSubs, log, time)
}

// ActivateSubscriptionForKnSubscription activates a Kyma Subscription when Kn Subscription is ready
func ActivateSubscriptionForKnSubscription(client eventingclientv1alpha1.EventingV1alpha1Interface, sub *subApis.Subscription, log logr.Logger, time CurrentTime) error {
	updatedSub := updateSubscriptionKnSubscriptionStatus(sub, subApis.ConditionTrue, time)
	return updateSubscription(client, updatedSub, log)
}

// DeactivateSubscriptions deactivate subscriptions.
func DeactivateSubscriptions(client eventingclientv1alpha1.EventingV1alpha1Interface, subs []*subApis.Subscription, log *zap.Logger, time CurrentTime) error {
	updatedSubs := updateSubscriptionsEventActivatedStatus(subs, subApis.ConditionFalse, time)
	return updateSubscriptions(client, updatedSubs, log, time)
}

// DeactivateSubscriptionForKnSubscription  deactivates a Kyma Subscription when Kn Subscription is not ready
func DeactivateSubscriptionForKnSubscription(client eventingclientv1alpha1.EventingV1alpha1Interface, sub *subApis.Subscription, log logr.Logger, time CurrentTime) error {
	updatedSub := updateSubscriptionKnSubscriptionStatus(sub, subApis.ConditionFalse, time)
	return updateSubscription(client, updatedSub, log)
}

func doesSubscriptionMatchLabels(sub *subApis.Subscription, labels map[string]string) bool {
	eventTypeMatched := false
	eventTypeVersionMatched := false
	sourceIDMatched := false
	if eventType, ok := labels[SubscriptionEventType]; ok {
		if sub.SubscriptionSpec.EventType == eventType {
			eventTypeMatched = true
		}
	}
	if eventTypeVersion, ok := labels[SubscriptionEventTypeVersion]; ok {
		if sub.SubscriptionSpec.EventTypeVersion == eventTypeVersion {
			eventTypeVersionMatched = true
		}
	}
	if sourceID, ok := labels[SubscriptionSourceID]; ok {
		if sub.SubscriptionSpec.SourceID == sourceID {
			sourceIDMatched = true
		}
	}
	if eventTypeMatched && eventTypeVersionMatched && sourceIDMatched {
		return true
	}
	return false
}

func updateSubscriptionsEventActivatedStatus(subs []*subApis.Subscription, conditionStatus subApis.ConditionStatus, time CurrentTime) []*subApis.Subscription {
	t := time.GetCurrentTime()
	var newCondition subApis.SubscriptionCondition
	if conditionStatus == subApis.ConditionTrue {
		newCondition = subApis.SubscriptionCondition{Type: subApis.EventsActivated, Status: subApis.ConditionTrue, LastTransitionTime: t}
	} else {
		newCondition = subApis.SubscriptionCondition{Type: subApis.EventsActivated, Status: subApis.ConditionFalse, LastTransitionTime: t}
	}

	var updatedSubs []*subApis.Subscription
	for _, s := range subs {
		if !s.HasCondition(newCondition) {
			s = updateSubscriptionStatus(s, subApis.EventsActivated, conditionStatus, "", time)
			updatedSubs = append(updatedSubs, s)
		}
	}
	return updatedSubs
}

func updateSubscriptionKnSubscriptionStatus(sub *subApis.Subscription, conditionStatus subApis.ConditionStatus, time CurrentTime) *subApis.Subscription {
	t := time.GetCurrentTime()
	var newCondition subApis.SubscriptionCondition
	if conditionStatus == subApis.ConditionTrue {
		newCondition = subApis.SubscriptionCondition{Type: subApis.SubscriptionReady, Status: subApis.ConditionTrue, LastTransitionTime: t}
	} else {
		newCondition = subApis.SubscriptionCondition{Type: subApis.SubscriptionReady, Status: subApis.ConditionFalse, LastTransitionTime: t}
	}

	if !sub.HasCondition(newCondition) {
		sub = updateSubscriptionStatus(sub, subApis.SubscriptionReady, conditionStatus, "", time)
	}
	return sub
}

func updateSubscriptionReadyStatus(sub *subApis.Subscription, conditionStatus subApis.ConditionStatus, msg string, time CurrentTime) *subApis.Subscription {
	return updateSubscriptionStatus(sub, subApis.Ready, conditionStatus, msg, time)
}

func updateSubscriptionStatus(sub *subApis.Subscription, conditionType subApis.SubscriptionConditionType,
	conditionStatus subApis.ConditionStatus, msg string, time CurrentTime) *subApis.Subscription {
	t := time.GetCurrentTime()
	newCondition := subApis.SubscriptionCondition{Type: conditionType, Status: conditionStatus, LastTransitionTime: t, Message: msg}
	if !sub.HasCondition(newCondition) {
		if len(sub.Status.Conditions) == 0 {
			sub.Status.Conditions = []subApis.SubscriptionCondition{newCondition}
		} else {
			var found bool
			for i, cond := range sub.Status.Conditions {
				if cond.Type == conditionType && cond.Status != conditionStatus {
					sub.Status.Conditions[i] = newCondition
					found = true
					break
				}
			}
			if !found {
				sub.Status.Conditions = append(sub.Status.Conditions, newCondition)
			}
		}
	}
	return sub
}

func updateSubscriptions(client eventingclientv1alpha1.EventingV1alpha1Interface, subs []*subApis.Subscription, log *zap.Logger, time CurrentTime) error {
	if subsWithErrors := WriteSubscriptions(client, subs); len(subsWithErrors) != 0 {
		// try to set the "Ready" status to false
		for _, es := range subsWithErrors {
			log.Error("WriteSubscriptions() failed for this subscription:", zap.String("subscription", es.Sub.Name), zap.Error(es.Err))
			us := updateSubscriptionReadyStatus(es.Sub, subApis.ConditionFalse, es.Err.Error(), time)
			if err := WriteSubscription(client, us); err != nil {
				log.Error("Update Ready status failed", zap.Error(err))
			}
		}
		return fmt.Errorf("WriteSubscriptions() failed, see the Ready status of each subscription")
	}
	return nil
}

func updateSubscription(client eventingclientv1alpha1.EventingV1alpha1Interface, sub *subApis.Subscription, log logr.Logger) error {
	err := WriteSubscription(client, sub)
	if err != nil {
		log.Error(err, "Update Ready status failed")
		return err
	}
	return nil
}
