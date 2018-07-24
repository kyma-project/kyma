package controllers

import "log"
import (
	subApis "github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma.cx/v1alpha1"
	"github.com/kyma-project/kyma/components/event-bus/internal/push/actors"
)

func getUpdateFnWithEventActivationCheck(supervisor actors.SubscriptionsSupervisorInterface) func(oldObj, newObj interface{}) {
	return func(oldObj, newObj interface{}) {
		if oldObj == newObj {
			return
		}

		_, oldSubOk := oldObj.(*subApis.Subscription)
		newSub, newSubOK := newObj.(*subApis.Subscription)

		if !oldSubOk || !newSubOK {
			log.Printf("unknown object type either updated %+v or original +%v", newObj, oldObj)
			return
		}

		if newSub.HasCondition(subApis.SubscriptionCondition{Type: subApis.EventsActivated, Status: subApis.ConditionFalse}) {
			log.Printf("Stop NATS Subscription %+v", newSub)
			supervisor.StopSubscriptionReq(newSub)
		}

		if newSub.HasCondition(subApis.SubscriptionCondition{Type: subApis.EventsActivated, Status: subApis.ConditionTrue}) {
			log.Printf("Start NATS Subscription %+v", newSub)
			supervisor.StartSubscriptionReq(newSub)
		}
	}
}

func getAddFnWithEventActivationCheck(supervisor actors.SubscriptionsSupervisorInterface) func(obj interface{}) {
	return func(obj interface{}) {
		subscription, ok := obj.(*subApis.Subscription)
		if !ok {
			log.Printf("unknown object type added %+v", obj)
			return
		}
		if subscription.HasCondition(subApis.SubscriptionCondition{Type: subApis.EventsActivated, Status: subApis.ConditionTrue}) {
			log.Printf("Subscription custom resource created %v", obj)
			supervisor.StartSubscriptionReq(subscription)
		}
	}
}
