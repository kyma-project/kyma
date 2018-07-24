package controllers

import (
	"log"

	subscriptionApis "github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma.cx/v1alpha1"
	"github.com/kyma-project/kyma/components/event-bus/internal/push/actors"
)

func getAddFnWithoutEventActivationCheck(supervisor actors.SubscriptionsSupervisorInterface) func(obj interface{}) {
	return func(obj interface{}) {
		subscription, ok := obj.(*subscriptionApis.Subscription)
		if ok {
			log.Printf("Added Subscription %+v", subscription)
			supervisor.StartSubscriptionReq(subscription)
		} else {
			log.Printf("unknown object type added %+v", obj)
		}
	}
}
