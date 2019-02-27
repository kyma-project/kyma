package subscription

import (
	"context"
	"log"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma-project.io/v1alpha1"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/opts"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	knative "github.com/kyma-project/kyma/components/event-bus/internal/knative/util"
)

// Add creates a new Subscription Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, opts *opts.Options) error {
	return add(mgr, newReconciler(mgr, opts))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, opts *opts.Options) reconcile.Reconciler {
	return &ReconcileSubscription{Client: mgr.GetClient(), scheme: mgr.GetScheme(), opts: opts}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("subscription-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Subscription
	err = c.Watch(&source.Kind{Type: &eventingv1alpha1.Subscription{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileSubscription{}

// ReconcileSubscription reconciles a Subscription object
type ReconcileSubscription struct {
	client.Client
	scheme *runtime.Scheme
	opts   *opts.Options
}

// Reconcile reads that state of the cluster for a Subscription object and makes changes based on the state read
// and what is in the Subscription.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a Deployment as an example
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=eventing.kyma-project.io,resources=subscriptions,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileSubscription) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the Subscription instance
	instance := &eventingv1alpha1.Subscription{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Printf("Kyma Subscription is deleted: %v", err)
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Printf("Could not fetch Kyma Subscription: %v", err)
		return reconcile.Result{}, err
	}

	subscription := instance.DeepCopy()

	requeue, err := r.reconcile(subscription)
	if err != nil {
		log.Printf("Error reconciling Subscription: %v", err)
	} else {
		log.Printf("Subscription reconciled")
	}

	return reconcile.Result{
		Requeue: requeue,
	}, err
}

func (r *ReconcileSubscription) reconcile(subscription *eventingv1alpha1.Subscription) (bool, error) {
	// init the knative lib
	knativeLib, err := knative.GetKnativeLib()
	if err != nil {
		log.Fatalf("Failed to get knative library: %v", err)
	}

	knativeSubsName := subscription.Name
	knativeSubsNamespace := subscription.Namespace
	knativeSubsURI := subscription.Endpoint
	knativeChannelName := knative.GetChannelName(&subscription.SourceID, &subscription.EventType, &subscription.EventTypeVersion)
	knativeChannelProvisioner := "natss"
	timeout := r.opts.ChannelTimeout

	// Finalizer for deleting Knative Subscriptions
	finalizerName := "subscription.finalizers.kyma-project.io"
	if subscription.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object.
		if !containsString(subscription.ObjectMeta.Finalizers, finalizerName) {
			subscription.ObjectMeta.Finalizers = append(subscription.ObjectMeta.Finalizers, finalizerName)
			if err := r.Update(context.Background(), subscription); err != nil {
				return true, nil
			}
		}
	} else {
		// The object is being deleted
		if containsString(subscription.ObjectMeta.Finalizers, finalizerName) {
			// our finalizer is present, so lets handle our external dependency
			if err := r.deleteExternalDependency(subscription, knativeLib, knativeChannelName); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return false, err
			}

			// remove our finalizer from the list and update it.
			subscription.ObjectMeta.Finalizers = removeString(subscription.ObjectMeta.Finalizers, finalizerName)
			if err := r.Update(context.Background(), subscription); err != nil {
				return true, nil
			}
		}

		// Our finalizer has finished, so the reconciler can do nothing.
		return false, nil
	}

	// Check if Kyma Subscription has events-activated condition.
	if subscription.HasCondition(eventingv1alpha1.SubscriptionCondition{Type: eventingv1alpha1.EventsActivated, Status: eventingv1alpha1.ConditionTrue}) {
		// Check if Knative Channel already exists, create if not.
		_, err := knativeLib.GetChannel(knativeChannelName, knativeSubsNamespace)
		if err != nil && !errors.IsNotFound(err) {
			return false, err
		} else if errors.IsNotFound(err) {
			knativeChannel, err := knativeLib.CreateChannel(knativeChannelProvisioner, knativeChannelName, knativeSubsNamespace, timeout)
			if err != nil {
				return false, err
			}
			log.Printf("Knative Channel is created: %v", knativeChannel)
		}

		// Check if Knative Subsription already exists, if not create one.
		sub, err := knativeLib.GetSubscription(knativeSubsName, knativeSubsNamespace)
		if err != nil && !errors.IsNotFound(err) {
			return false, err
		} else if errors.IsNotFound(err) {
			err = knativeLib.CreateSubscription(knativeSubsName, knativeSubsNamespace, knativeChannelName, &knativeSubsURI)
			if err != nil {
				return false, err
			}
			log.Printf("Knative Subscription is created: %s", knativeSubsName)
		} else {
			// In case there is a change in Channel name or URI, delete and re-create Knative Subscription because update does not work.
			if knativeChannelName != sub.Spec.Channel.Name || knativeSubsURI != *sub.Spec.Subscriber.DNSName {
				err = knativeLib.DeleteSubscription(knativeSubsName, knativeSubsNamespace)
				if err != nil {
					return false, err
				}
				log.Printf("Knative Subscription is deleted: %s", knativeSubsName)
				err = knativeLib.CreateSubscription(knativeSubsName, knativeSubsNamespace, knativeChannelName, &knativeSubsURI)
				if err != nil {
					return false, err
				}
				log.Printf("Knative Subscription is re-created: %s", knativeSubsName)
			}
		}
	} else {
		// In case Kyma Subscription does not have events-activated condition, delete Knative Subscription if exists.
		knativeSubs, err := knativeLib.GetSubscription(knativeSubsName, knativeSubsNamespace)
		if err != nil && !errors.IsNotFound(err) {
			return false, err
		} else if err == nil && knativeSubs != nil {
			err = knativeLib.DeleteSubscription(knativeSubsName, knativeSubsNamespace)
			if err != nil {
				return false, err
			}
			log.Printf("Knative Subscription is deleted: %s", knativeSubsName)
		}

		// Check if Channel has any other Subscription, if not, delete it.
		knativeChannel, err := knativeLib.GetChannel(knativeChannelName, knativeSubsNamespace)
		if err != nil && !errors.IsNotFound(err) {
			return false, err
		} else if err == nil && knativeChannel != nil {
			if knativeChannel.Spec.Subscribable == nil || len(knativeChannel.Spec.Subscribable.Subscribers) == 0 ||
				(len(knativeChannel.Spec.Subscribable.Subscribers) == 1 && knativeChannel.Spec.Subscribable.Subscribers[0].SubscriberURI == subscription.Endpoint) {
				err = knativeLib.DeleteChannel(knativeChannelName, knativeSubsNamespace)
				if err != nil {
					return false, err
				}
				log.Printf("Knative Channel is deleted: %v", knativeChannel)
			}
		}
	}

	return false, nil
}

func (r *ReconcileSubscription) deleteExternalDependency(subscription *eventingv1alpha1.Subscription, knativeLib *knative.KnativeLib, channelName string) error {
	log.Printf("Deleting the external dependencies")

	// In case Knative Subscription exists, delete it.
	knativeSubs, err := knativeLib.GetSubscription(subscription.Name, subscription.Namespace)
	if err != nil && !errors.IsNotFound(err) {
		return err
	} else if err == nil && knativeSubs != nil {
		err = knativeLib.DeleteSubscription(knativeSubs.Name, knativeSubs.Namespace)
		if err != nil {
			return err
		}
		log.Printf("Knative Subscription is deleted: %s", knativeSubs.Name)
	}

	// Check if Channel has any other Subscription, if not, delete it.
	knativeChannel, err := knativeLib.GetChannel(channelName, subscription.Namespace)
	if err != nil && !errors.IsNotFound(err) {
		return err
	} else if err == nil && knativeChannel != nil {
		if knativeChannel.Spec.Subscribable == nil || (len(knativeChannel.Spec.Subscribable.Subscribers) == 1 &&
			knativeChannel.Spec.Subscribable.Subscribers[0].SubscriberURI == subscription.Endpoint) {
			err = knativeLib.DeleteChannel(channelName, subscription.Namespace)
			if err != nil {
				return err
			}
			log.Printf("Knative Channel is deleted: %v", knativeChannel)
		}
	}
	return nil
}

//
// Helper functions to check and remove string from a slice of strings.
//
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}
