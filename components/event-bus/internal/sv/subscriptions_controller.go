package ea

import (
	"log"

	subApis "github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma.cx/v1alpha1"
	eaclientset "github.com/kyma-project/kyma/components/event-bus/generated/ea/clientset/versioned"
	subscriptionClientSet "github.com/kyma-project/kyma/components/event-bus/generated/push/clientset/versioned"
	subv1alpha1 "github.com/kyma-project/kyma/components/event-bus/generated/push/informers/externalversions/eventing.kyma.cx/v1alpha1"
	"github.com/kyma-project/kyma/components/event-bus/internal/sv/opts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

// Controller observes EventActivations CRs and updates the related Subscriptions-CRs status
type SubscriptionsController struct {
	informer    cache.SharedIndexInformer
	stopChannel chan struct{}
	running     bool
}

// StartController creates and starts an EventActivationsController
func StartSubscriptionsController() *SubscriptionsController {
	return StartSubscriptionsControllerWithInformer(createSubscriptionsInformer())
}

// StartControllerWithInformer is a factory for the Controller using the specified informer
func StartSubscriptionsControllerWithInformer(informer cache.SharedIndexInformer) *SubscriptionsController {
	controller := &SubscriptionsController{
		informer:    informer,
		stopChannel: make(chan struct{}),
	}
	go controller.informer.Run(controller.stopChannel)
	controller.running = true
	return controller
}

// IsRunning...
func (controller *SubscriptionsController) IsRunning() bool {
	return controller.running
}

// Stop the Controller
func (controller *SubscriptionsController) Stop() {
	controller.stopChannel <- struct{}{}
	controller.running = false
}

func createSubscriptionsInformer() cache.SharedIndexInformer {
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

	informer := subv1alpha1.NewSubscriptionInformer(subClient, metav1.NamespaceAll, opts.GetOptions().ResyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			log.Printf("Subscription custom resource created:\n    %v\n", obj)
			subObj, ok := obj.(*subApis.Subscription)
			if !ok {
				log.Printf("Error: Not a Subscription object:\n    %v\n", obj)
				return
			}
			if checkEventActivationForSubscription(eaClient, subObj) {
				activateSubscriptions(subClient, subObj.GetNamespace(), []*subApis.Subscription{subObj})
			} else {
				deactivateSubscriptions(subClient, subObj.GetNamespace(), []*subApis.Subscription{subObj})
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if oldObj != newObj {
				log.Printf("Subscription custom resource updated, old:\n    %v\n    new: %v\n", oldObj, newObj)
			}
		},
		DeleteFunc: func(obj interface{}) {
			log.Printf("Subscription custom resource deleted:\n    %v\n", obj)
		},
	})
	return informer
}
