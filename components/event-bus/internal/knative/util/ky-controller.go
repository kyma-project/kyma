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
	lo := &runtimeClient.ListOptions{Namespace: subNamespace}
	if err := client.List(ctx, lo, eal); err != nil {
		return false
	}
	for _, ea := range eal.Items {
		if subSourceID == ea.SourceID && ea.DeletionTimestamp.IsZero()  {
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
	lo := &runtimeClient.ListOptions{Namespace: eaNamespace} // query using SourceID too?
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
	if err := client.Update(ctx, sub); err != nil {
		return err
	}
	return nil
}

func UpdateSubscriptionsEventActivatedStatus(subs []*subApis.Subscription, conditionStatus subApis.ConditionStatus) []*subApis.Subscription {
	t := metav1.NewTime(time.Now())
	var newCondition subApis.SubscriptionCondition
	if conditionStatus == subApis.ConditionTrue {
		newCondition = subApis.SubscriptionCondition{Type: subApis.EventsActivated, Status: subApis.ConditionTrue, LastTransitionTime: t}
	} else {
		newCondition = subApis.SubscriptionCondition{Type: subApis.EventsActivated, Status: subApis.ConditionFalse, LastTransitionTime: t}
	}

	var updatedSubs []*subApis.Subscription
	for _, s := range subs {
		if !s.HasCondition(newCondition) {
			if len(s.Status.Conditions) == 0 {
				s.Status.Conditions = []subApis.SubscriptionCondition{newCondition}
				updatedSubs = append(updatedSubs, s)
			} else {
				for i, cond := range s.Status.Conditions {
					if cond.Type == subApis.EventsActivated && cond.Status != conditionStatus {
						s.Status.Conditions[i] = newCondition
						updatedSubs = append(updatedSubs, s)
						break
					}
				}
			}
		}
	}
	return updatedSubs
}

func UpdateSubscriptionReadyStatus(sub *subApis.Subscription, conditionStatus subApis.ConditionStatus, msg string) *subApis.Subscription {
	t := metav1.NewTime(time.Now())
	var newCondition subApis.SubscriptionCondition
	if conditionStatus == subApis.ConditionTrue {
		newCondition = subApis.SubscriptionCondition{Type: subApis.Ready, Status: subApis.ConditionTrue, LastTransitionTime: t, Message: msg}
	} else {
		newCondition = subApis.SubscriptionCondition{Type: subApis.Ready, Status: subApis.ConditionFalse, LastTransitionTime: t, Message: msg}
	}
	if !sub.HasCondition(newCondition) {
		if len(sub.Status.Conditions) == 0 {
			sub.Status.Conditions = []subApis.SubscriptionCondition{newCondition}
		} else {
			for i, cond := range sub.Status.Conditions {
				if cond.Type == subApis.Ready && cond.Status != conditionStatus {
					sub.Status.Conditions[i] = newCondition
					break
				}
			}
		}
	}
	return sub
}

func ActivateSubscriptions(ctx context.Context, client runtimeClient.Client, subs []*subApis.Subscription, log logr.Logger) error {
	updatedSubs := UpdateSubscriptionsEventActivatedStatus(subs, subApis.ConditionTrue)
	return updateSubscriptions(ctx, client, updatedSubs, log)
}

func DeactivateSubscriptions(ctx context.Context, client runtimeClient.Client, subs []*subApis.Subscription, log logr.Logger) error {
	updatedSubs := UpdateSubscriptionsEventActivatedStatus(subs, subApis.ConditionFalse)
	return updateSubscriptions(ctx, client, updatedSubs, log)
}

func updateSubscriptions(ctx context.Context, client runtimeClient.Client, subs []*subApis.Subscription, log logr.Logger) error {
	if subsWithErrors := WriteSubscriptions(ctx, client, subs); len(subsWithErrors) != 0 {
		// try to set the "Ready" status to false
		for _, es := range subsWithErrors {
			log.Error(es.Err, "WriteSubscriptions() failed for this subscription:", "subscription", es.Sub)
			us := UpdateSubscriptionReadyStatus(es.Sub, subApis.ConditionFalse, es.Err.Error())
			if err := WriteSubscription(ctx, client, us); err != nil {
				log.Error(err, "Update Ready status failed")
			}
		}
		return fmt.Errorf("WriteSubscriptions() failed, see the Ready status of each subscription")
	}
	return nil
}
