package controllers

import (
	"log"

	subApis "github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma-project.io/v1alpha1"
)

func checkSubscriptions(oldObj, newObj interface{}) (*subApis.Subscription, *subApis.Subscription, bool) {
	oldSub, oldSubOk := oldObj.(*subApis.Subscription)
	newSub, newSubOK := newObj.(*subApis.Subscription)
	if !oldSubOk || !newSubOK {
		log.Printf("checkSubscriptions() failed: unknown object type either updated %+v or original +%v", newObj, oldObj)
		return nil, nil, false
	}
	return oldSub, newSub, true
}

func checkSubscription(obj interface{}) (*subApis.Subscription, bool) {
	sub, ok := obj.(*subApis.Subscription)
	if !ok {
		log.Printf("checkSubscription() failed: unknown object type %+v", obj)
		return nil, false
	}
	return sub, true
}
