package controllers

import (
	"log"

	subApis "github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma-project.io/v1alpha1"
	"github.com/kyma-project/kyma/components/event-bus/internal/common"
	"github.com/kyma-project/kyma/components/event-bus/internal/push/actors"
)

func getUpdateFnWithEventActivationCheck(supervisor actors.SubscriptionsSupervisorInterface) func(oldObj, newObj interface{}) {
	return func(oldObj, newObj interface{}) {
		if oldObj == newObj {
			return
		}
		if oldSub, newSub, ok := checkSubscriptions(oldObj, newObj); ok {
			if newSub.HasCondition(subApis.SubscriptionCondition{Type: subApis.EventsActivated, Status: subApis.ConditionFalse}) {
				log.Printf("Stop NATS Subscription %+v", newSub)
				supervisor.StopSubscriptionReq(newSub)
			} else {
				log.Printf("Stop old NATS Subscription %+v", oldSub)
				supervisor.StopSubscriptionReq(oldSub)
			}
			if newSub.HasCondition(subApis.SubscriptionCondition{Type: subApis.EventsActivated, Status: subApis.ConditionTrue}) {
				log.Printf("Start NATS Subscription %+v", newSub)
				supervisor.StartSubscriptionReq(newSub, common.DefaultRequestProvider)
			}
		}
	}
}

func getAddFnWithEventActivationCheck(supervisor actors.SubscriptionsSupervisorInterface) func(obj interface{}) {
	return func(obj interface{}) {
		if subscription, ok := checkSubscription(obj); ok {
			if subscription.HasCondition(subApis.SubscriptionCondition{Type: subApis.EventsActivated, Status: subApis.ConditionTrue}) {
				log.Printf("Subscription custom resource created %v", obj)
				supervisor.StartSubscriptionReq(subscription, common.DefaultRequestProvider)
			}
		}
	}
}
