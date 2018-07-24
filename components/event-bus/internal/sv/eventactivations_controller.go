package ea

import (
	"log"

	eaclientset "github.com/kyma-project/kyma/components/event-bus/generated/ea/clientset/versioned"
	eav1alpha1 "github.com/kyma-project/kyma/components/event-bus/generated/ea/informers/externalversions/remoteenvironment.kyma.cx/v1alpha1"
	subscriptionClientSet "github.com/kyma-project/kyma/components/event-bus/generated/push/clientset/versioned"
	eaApis "github.com/kyma-project/kyma/components/event-bus/internal/ea/apis/remoteenvironment.kyma.cx/v1alpha1"
	"github.com/kyma-project/kyma/components/event-bus/internal/sv/opts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

// EventActivationsController observes EventActivations CRs and updates the related Subscriptions-CRs status
type EventActivationsController struct {
	informer    cache.SharedIndexInformer
	stopChannel chan struct{}
	running     bool
}

// StartEventActivationsController creates and starts an EventActivationsController
func StartEventActivationsController() *EventActivationsController {
	return StartEventActivationsControllerWithInformer(createEventActivationsInformer())
}

// StartEventActivationsControllerWithInformer is a factory for the Controller using the specified informer
func StartEventActivationsControllerWithInformer(informer cache.SharedIndexInformer) *EventActivationsController {
	controller := &EventActivationsController{
		informer:    informer,
		stopChannel: make(chan struct{}),
	}
	go controller.informer.Run(controller.stopChannel)
	controller.running = true
	return controller
}

// IsRunning...
func (controller *EventActivationsController) IsRunning() bool {
	return controller.running
}

// Stop the Controller
func (controller *EventActivationsController) Stop() {
	controller.stopChannel <- struct{}{}
	controller.running = false
}

func createEventActivationsInformer() cache.SharedIndexInformer {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Panicf("Error in getting cluster config - %+v", err)
	}
	eaClient, err := eaclientset.NewForConfig(config)
	if err != nil {
		log.Panicf("Error in creating event activation client - %+v", err)
	}
	subClient, err := subscriptionClientSet.NewForConfig(config)
	if err != nil {
		log.Panicf("Error in creating subscription client - %+v", err)
	}

	informer := eav1alpha1.NewEventActivationInformer(eaClient, metav1.NamespaceAll, opts.GetOptions().ResyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			log.Printf("Event Activation custom resource created:\n    %v\n", obj)
			eaObj, ok := obj.(*eaApis.EventActivation)
			if !ok {
				log.Printf("Error: Not an Event Activation object: %v\n", obj)
				return
			}
			if subs, err := getSubscriptionsForEventActivation(subClient, eaObj); err == nil {
				activateSubscriptions(subClient, eaObj.GetNamespace(), subs)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if oldObj != newObj {
				log.Printf("Event Activation custom resource updated, old:\n    %v\n    new: %v\n", oldObj, newObj)
			}
		},
		DeleteFunc: func(obj interface{}) {
			log.Printf("Event Activation custom resource deleted:\n    %v\n", obj)
			eaObj := obj.(*eaApis.EventActivation)
			if subs, err := getSubscriptionsForEventActivation(subClient, eaObj); err == nil {
				deactivateSubscriptions(subClient, eaObj.GetNamespace(), subs)
			}
		},
	})
	return informer
}
