package controllers

import (
	"log"

	"time"

	subscriptionApis "github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma.cx/v1alpha1"
	subscriptionClientSet "github.com/kyma-project/kyma/components/event-bus/generated/push/clientset/versioned"
	"github.com/kyma-project/kyma/components/event-bus/generated/push/informers/externalversions/eventing.kyma.cx/v1alpha1"
	"github.com/kyma-project/kyma/components/event-bus/internal/push/actors"
	"github.com/kyma-project/kyma/components/event-bus/internal/push/opts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

// SubscriptionsController observes Subscription CRs and ensures that for each there is matching NATS Streaming subscription
type SubscriptionsController struct {
	informer    cache.SharedIndexInformer
	supervisor  actors.SubscriptionsSupervisorInterface
	stopChannel chan struct{}
}

// StartSubscriptionsController is a factory method for the SubscriptionsController
func StartSubscriptionsController(supervisor *actors.SubscriptionsSupervisor, pushOptions *opts.Options) *SubscriptionsController {
	informer := createSubscriptionsInformer()
	return StartSubscriptionsControllerWithInformer(supervisor, informer, pushOptions)
}

// StartSubscriptionsControllerWithInformer is a factory for the SubscriptionsController which method uses the specified informer
func StartSubscriptionsControllerWithInformer(supervisor *actors.SubscriptionsSupervisor, informer cache.SharedIndexInformer, pushOptions *opts.Options) *SubscriptionsController {
	stopChan := make(chan struct{})

	controller := &SubscriptionsController{
		informer:    informer,
		supervisor:  supervisor,
		stopChannel: stopChan,
	}

	if pushOptions.CheckEventsActivation {
		controller.startInformerWithEventActivationCheck()
	} else {
		controller.startInformerWithoutEventActivationCheck()
	}

	return controller
}

func createSubscriptionsInformer() cache.SharedIndexInformer {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Panicf("Error in getting cluster config - %+v", err)
	}
	subscriptionClient, err := subscriptionClientSet.NewForConfig(config)
	if err != nil {
		log.Panicf("Error in creating client - %+v", err)
	}
	informer := v1alpha1.NewSubscriptionInformer(subscriptionClient, metav1.NamespaceAll, 1*time.Minute, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	return informer
}

// Stop halts the SubscriptionsController
func (controller *SubscriptionsController) Stop() {
	controller.stopChannel <- struct{}{}
}

func (controller *SubscriptionsController) startInformerWithoutEventActivationCheck() {
	controller.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: getAddFnWithoutEventActivationCheck(controller.supervisor),
		UpdateFunc: func(oldObj, newObj interface{}) {
			//log.Print("Updated")
		},
		DeleteFunc: controller.getDeleteFn(),
	})

	go controller.informer.Run(controller.stopChannel)
}

func (controller *SubscriptionsController) startInformerWithEventActivationCheck() {
	controller.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    getAddFnWithEventActivationCheck(controller.supervisor),
		UpdateFunc: getUpdateFnWithEventActivationCheck(controller.supervisor),
		DeleteFunc: controller.getDeleteFn(),
	})

	go controller.informer.Run(controller.stopChannel)
}

func (controller *SubscriptionsController) getDeleteFn() func(obj interface{}) {
	return func(obj interface{}) {
		subscription, ok := obj.(*subscriptionApis.Subscription)
		if ok {
			log.Printf("Deleted Subscription %+v", subscription)
			controller.supervisor.StopSubscriptionReq(subscription)
		} else {
			log.Printf("unknown object type deleted %+v", obj)
		}
	}
}
