package util

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	subApis "github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma-project.io/v1alpha1"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/event-bus/internal/ea/apis/applicationconnector.kyma-project.io/v1alpha1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Helper functions to check and remove string from a slice of strings.
func ContainsString(slice *[]string, s string) bool {
	for _, item := range *slice {
		if item == s {
			return true
		}
	}
	return false
}

func RemoveString(slice *[]string, s string) (result []string) {
	for _, item := range *slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

//
// Handle Kyma EventActivation
//
func UpdateEventActivation(ctx context.Context, client runtimeClient.Client, u *eventingv1alpha1.EventActivation) error {
	objectKey := runtimeClient.ObjectKey{Namespace: u.Namespace, Name: u.Name}
	ea := &eventingv1alpha1.EventActivation{}
	if err := client.Get(ctx, objectKey, ea); err != nil {
		return err
	}

	if !equality.Semantic.DeepEqual(ea.Finalizers, u.Finalizers) {
		ea.SetFinalizers(u.ObjectMeta.Finalizers)
		if err := client.Update(ctx, ea); err != nil {
			return err
		}
	}
	return nil
}

// CheckIfEventActivationExistForSubscription
func CheckIfEventActivationExistForSubscription(ctx context.Context, client runtimeClient.Client, sub *subApis.Subscription) bool {
	subNamespace := sub.GetNamespace()
	subSourceID := sub.SourceID

	eal := &eventingv1alpha1.EventActivationList{}
	lo := &runtimeClient.ListOptions{
		Namespace: subNamespace,
		Raw: &metav1.ListOptions{ // TODO this is here because the fake client needs it. Remove this when it's no longer needed.
			TypeMeta: metav1.TypeMeta{
				APIVersion: eventingv1alpha1.SchemeGroupVersion.String(),
				Kind:       "EventActivation",
			},
		},
	}
	if err := client.List(ctx, lo, eal); err != nil {
		return false
	}
	for _, ea := range eal.Items {
		if subSourceID == ea.SourceID && ea.DeletionTimestamp.IsZero() {
			return true
		}
	}
	return false
}

// GetSubscriptionsForEventActivation() gets all the subscriptions having the same "namespace" and the same "Source" as the "ea" object
func GetSubscriptionsForEventActivation(ctx context.Context, client runtimeClient.Client, ea *eventingv1alpha1.EventActivation) ([]*subApis.Subscription, error) {
	eaNamespace := ea.GetNamespace()
	eaSourceID := ea.EventActivationSpec.SourceID

	sl := &subApis.SubscriptionList{}
	lo := &runtimeClient.ListOptions{ // query using SourceID too?
		Namespace: eaNamespace,
		Raw: &metav1.ListOptions{ // TODO this is here because the fake client needs it. Remove this when it's no longer needed.
			TypeMeta: metav1.TypeMeta{
				APIVersion: subApis.SchemeGroupVersion.String(),
				Kind:       "Subscription",
			},
		},
	}
	if err := client.List(ctx, lo, sl); err != nil {
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

//
// Handle current time
//
type CurrentTime interface {
	GetCurrentTime() metav1.Time
}

type DefaultCurrentTime struct{}

func NewDefaultCurrentTime() CurrentTime {
	return new(DefaultCurrentTime)
}

func (t *DefaultCurrentTime) GetCurrentTime() metav1.Time {
	return metav1.NewTime(time.Now())
}

//
// Handle Kyma subscriptions
//
type SubscriptionWithError struct {
	Sub *subApis.Subscription
	Err error
}

func WriteSubscriptions(ctx context.Context, client runtimeClient.Client, subs []*subApis.Subscription) []SubscriptionWithError {
	var errorSubs []SubscriptionWithError
	for _, u := range subs {
		if err := WriteSubscription(ctx, client, u); err != nil {
			errorSubs = append(errorSubs, SubscriptionWithError{Sub: u, Err: err})
		}
	}
	return errorSubs
}

func WriteSubscription(ctx context.Context, client runtimeClient.Client, sub *subApis.Subscription) error {
	// update the subscription status subresource
	if err := client.Status().Update(ctx, sub.DeepCopy()); err != nil {
		return err
	}
	// update the subscription resource
	if err := client.Update(ctx, sub); err != nil {
		return err
	}
	return nil
}

func SetReadySubscription(ctx context.Context, client runtimeClient.Client, sub *subApis.Subscription, msg string, time CurrentTime) error {
	us := updateSubscriptionReadyStatus(sub, subApis.ConditionTrue, msg, time)
	return WriteSubscription(ctx, client, us)
}

func SetNotReadySubscription(ctx context.Context, client runtimeClient.Client, sub *subApis.Subscription, msg string, time CurrentTime) error {
	us := updateSubscriptionReadyStatus(sub, subApis.ConditionFalse, "", time)
	return WriteSubscription(ctx, client, us)
}

func IsSubscriptionActivated(sub *subApis.Subscription) bool {
	activatedCondition := subApis.SubscriptionCondition{Type: subApis.EventsActivated, Status: subApis.ConditionTrue}
	return sub.HasCondition(activatedCondition)

}

func ActivateSubscriptions(ctx context.Context, client runtimeClient.Client, subs []*subApis.Subscription, log logr.Logger, time CurrentTime) error {
	updatedSubs := updateSubscriptionsEventActivatedStatus(subs, subApis.ConditionTrue, time)
	return updateSubscriptions(ctx, client, updatedSubs, log, time)
}

func DeactivateSubscriptions(ctx context.Context, client runtimeClient.Client, subs []*subApis.Subscription, log logr.Logger, time CurrentTime) error {
	updatedSubs := updateSubscriptionsEventActivatedStatus(subs, subApis.ConditionFalse, time)
	return updateSubscriptions(ctx, client, updatedSubs, log, time)
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

func updateSubscriptions(ctx context.Context, client runtimeClient.Client, subs []*subApis.Subscription, log logr.Logger, time CurrentTime) error {
	if subsWithErrors := WriteSubscriptions(ctx, client, subs); len(subsWithErrors) != 0 {
		// try to set the "Ready" status to false
		for _, es := range subsWithErrors {
			log.Error(es.Err, "WriteSubscriptions() failed for this subscription:", "subscription", es.Sub)
			us := updateSubscriptionReadyStatus(es.Sub, subApis.ConditionFalse, es.Err.Error(), time)
			if err := WriteSubscription(ctx, client, us); err != nil {
				log.Error(err, "Update Ready status failed")
			}
		}
		return fmt.Errorf("WriteSubscriptions() failed, see the Ready status of each subscription")
	}
	return nil
}
