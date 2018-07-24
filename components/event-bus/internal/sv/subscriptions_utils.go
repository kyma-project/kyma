package ea

import (
	"log"
	"time"

	subApis "github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma.cx/v1alpha1"
	subscriptionClientSet "github.com/kyma-project/kyma/components/event-bus/generated/push/clientset/versioned"
	eaApis "github.com/kyma-project/kyma/components/event-bus/internal/ea/apis/remoteenvironment.kyma.cx/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// find all the subscriptions having the same "namespace" as the "ea" namespace and the same "Source"
func getSubscriptionsForEventActivation(subClient *subscriptionClientSet.Clientset, eaObj *eaApis.EventActivation) ([]*subApis.Subscription, error) {
	eaNamespace := eaObj.GetNamespace()
	eaSource := eaObj.EventActivationSpec.Source

	subscrList, err := subClient.EventingV1alpha1().Subscriptions(eaNamespace).List(metav1.ListOptions{}) // TODO query on eaSource??
	if err != nil {
		log.Printf("Error: List Subscriptions call failed for event activatsion:\n    %v;\n    Error:%v\n", eaObj, err)
		return nil, err
	}
	var subs []*subApis.Subscription
	for _, s := range subscrList.Items {
		if eaSource.Environment == s.Source.SourceEnvironment &&
			eaSource.Namespace == s.Source.SourceNamespace &&
			eaSource.Type == s.Source.SourceType {
			subs = append(subs, &s)
		}
	}
	return subs, err
}

func writeSubscriptions(subClient *subscriptionClientSet.Clientset, namespace string, subs []*subApis.Subscription) {
	for _, u := range subs {
		_, err := subClient.EventingV1alpha1().Subscriptions(namespace).Update(u)
		if err != nil {
			log.Printf("Error: Update Subscription call failed for subscription:\n    %v,\n    %v\n", u, err)
		}
	}
}

func activateSubscriptions(subClient *subscriptionClientSet.Clientset, namespace string, subs []*subApis.Subscription) {
	updatedSubs := updateSubscriptionsStatus(subs, subApis.ConditionTrue)
	writeSubscriptions(subClient, namespace, updatedSubs)
}

func deactivateSubscriptions(subClient *subscriptionClientSet.Clientset, namespace string, subs []*subApis.Subscription) {
	updatedSubs := updateSubscriptionsStatus(subs, subApis.ConditionFalse)
	writeSubscriptions(subClient, namespace, updatedSubs)
}

func updateSubscriptionsStatus(subs []*subApis.Subscription, conditionStatus subApis.ConditionStatus) []*subApis.Subscription {
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
