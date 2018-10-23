package controllers

import (
	"log"

	"github.com/kyma-project/kyma/components/event-bus/internal/common"
	"github.com/kyma-project/kyma/components/event-bus/internal/push/actors"
)

func getAddFnWithoutEventActivationCheck(supervisor actors.SubscriptionsSupervisorInterface) func(obj interface{}) {
	return func(obj interface{}) {
		if subscription, ok := checkSubscription(obj); ok {
			log.Printf("Added Subscription %+v", subscription)
			supervisor.StartSubscriptionReq(subscription, common.DefaultRequestProvider)
		}
	}
}

func getUpdateFnWithoutEventActivationCheck(supervisor actors.SubscriptionsSupervisorInterface) func(oldObj, newObj interface{}) {
	return func(oldObj, newObj interface{}) {
		if oldObj == newObj {
			return
		}
		if oldSub, newSub, ok := checkSubscriptions(oldObj, newObj); ok {
			log.Printf("Stop old NATS Subscription %+v", oldSub)
			supervisor.StopSubscriptionReq(oldSub)

			log.Printf("Start new NATS Subscription %+v", newSub)
			supervisor.StartSubscriptionReq(newSub, common.DefaultRequestProvider)
		}
	}
}
