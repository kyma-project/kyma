package subscription_nats

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/go-logr/logr"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/eventtype"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Reconciler reconciles a Subscription object
type Reconciler struct {
	ctx context.Context
	client.Client
	cache.Cache
	Backend          handlers.MessagingBackend
	Log              logr.Logger
	recorder         record.EventRecorder
	eventTypeCleaner eventtype.Cleaner
}

var (
	Finalizer = eventingv1alpha1.GroupVersion.Group
)

const (
	NATSProtocol = "NATS"
	// NATSFirstInstanceName the name of first instance of NATS cluster
	NATSFirstInstanceName = "eventing-nats-1"
	// NATSNamespace namespace of NATS cluster
	NATSNamespace = "kyma-system"
)

func NewReconciler(ctx context.Context, client client.Client, applicationLister *application.Lister, cache cache.Cache,
	log logr.Logger, recorder record.EventRecorder, cfg env.NatsConfig) *Reconciler {
	natsHandler := handlers.NewNats(cfg, log)
	err := natsHandler.Initialize(env.Config{})
	if err != nil {
		log.Error(err, "reconciler can't start")
		panic(err)
	}
	return &Reconciler{
		ctx:              ctx,
		Client:           client,
		Cache:            cache,
		Backend:          natsHandler,
		Log:              log,
		recorder:         recorder,
		eventTypeCleaner: eventtype.NewCleaner(cfg.EventTypePrefix, applicationLister, log),
	}
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&eventingv1alpha1.Subscription{}).
		Complete(r)
}

//  SetupUnmanaged creates a controller under the client control
func (r *Reconciler) SetupUnmanaged(mgr ctrl.Manager) error {
	ctru, err := controller.NewUnmanaged("nats-subscription-controller", mgr, controller.Options{
		Reconciler: r,
	})
	if err != nil {
		r.Log.Error(err, "failed to create a unmanaged NATS controller")
		return err
	}

	if err := ctru.Watch(&source.Kind{Type: &eventingv1alpha1.Subscription{}}, &handler.EnqueueRequestForObject{}); err != nil {
		r.Log.Error(err, "unable to watch subscriptions")
		return err
	}

	p := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			if e.Object.GetName() == NATSFirstInstanceName && e.Object.GetNamespace() == NATSNamespace {
				return true
			}
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			if e.Object.GetName() == NATSFirstInstanceName && e.Object.GetNamespace() == NATSNamespace {
				return true
			}
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return false
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}
	if err := ctru.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForObject{}, p); err != nil {
		r.Log.Error(err, "unable to watch eventing-nats-1 pod")
		return err
	}

	go func(r *Reconciler, c controller.Controller) {
		if err := c.Start(r.ctx); err != nil {
			r.Log.Error(err, "failed to start the nats-subscription-controller")
			os.Exit(1)
		}
	}(r, ctru)

	return nil
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	if req.Name == NATSFirstInstanceName && req.Namespace == NATSNamespace {
		r.Log.Info("received watch request", "namespace", req.Namespace, "name", req.Name)
		return r.syncInvalidSubscriptions(ctx)
	}

	r.Log.Info("received subscription reconciliation request", "namespace", req.Namespace, "name", req.Name)

	cachedSubscription := &eventingv1alpha1.Subscription{}

	// Ensure the object was not deleted in the meantime
	err := r.Client.Get(ctx, req.NamespacedName, cachedSubscription)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Handle only the new subscription
	subscription := cachedSubscription.DeepCopy()
	//Bind fields to logger
	log := r.Log.WithValues("kind", subscription.GetObjectKind().GroupVersionKind().Kind,
		"name", subscription.GetName(),
		"namespace", subscription.GetNamespace(),
		"version", subscription.GetGeneration(),
	)

	if !subscription.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is being deleted
		if utils.ContainsString(subscription.ObjectMeta.Finalizers, Finalizer) {
			if err := r.Backend.DeleteSubscription(subscription); err != nil {
				r.Log.Error(err, "failed to delete subscription")
				// if failed to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			log.Info("Removing finalizer from subscription object")
			subscription.ObjectMeta.Finalizers = utils.RemoveString(subscription.ObjectMeta.Finalizers,
				Finalizer)
			if err := r.Client.Update(ctx, subscription); err != nil {
				log.Error(err, "failed to remove finalizer from subscription object")
				if k8serrors.IsConflict(err) {
					return ctrl.Result{Requeue: true}, nil
				}
			}
			return ctrl.Result{}, nil
		}
	}

	// The object is not being deleted, so if it does not have our finalizer,
	// then lets add the finalizer and update the object. This is equivalent
	// registering our finalizer.
	if !utils.ContainsString(subscription.ObjectMeta.Finalizers, Finalizer) {
		log.Info("Adding finalizer to subscription object")
		subscription.ObjectMeta.Finalizers = append(subscription.ObjectMeta.Finalizers, Finalizer)
		if err := r.Update(context.Background(), subscription); err != nil {
			if k8serrors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			}
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Check for valid sink
	if err := r.assertSinkValidity(subscription.Spec.Sink); err != nil {
		r.Log.Error(err, "failed to parse sink URL")
		if err := r.syncSubscriptionStatus(ctx, subscription, false, err.Error()); err != nil {
			if k8serrors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			}
		}
		// No point in reconciling as the sink is invalid
		return ctrl.Result{}, nil
	}

	// Clean up the old subscriptions
	err = r.Backend.DeleteSubscription(subscription)
	if err != nil {
		log.Error(err, "failed to delete subscriptions")
		if err := r.syncSubscriptionStatus(ctx, subscription, false, err.Error()); err != nil {
			if k8serrors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			}
		}
		return ctrl.Result{}, err
	}

	_, err = r.Backend.SyncSubscription(subscription, r.eventTypeCleaner)
	if err != nil {
		r.Log.Error(err, "failed to sync subscription")
		if err := r.syncSubscriptionStatus(ctx, subscription, false, err.Error()); err != nil {
			if k8serrors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			}
		}
		return ctrl.Result{}, err
	}

	log.Info("successfully created Nats subscriptions")

	// Update status
	if err := r.syncSubscriptionStatus(ctx, subscription, true, ""); err != nil {
		if k8serrors.IsConflict(err) {
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// syncSubscriptionStatus syncs Subscription status
func (r *Reconciler) syncSubscriptionStatus(ctx context.Context, sub *eventingv1alpha1.Subscription, isNatsSubReady bool, message string) error {
	desiredConditions := make([]eventingv1alpha1.Condition, 0)
	conditionContained := false
	conditionsUpdated := false
	condition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscriptionActive,
		eventingv1alpha1.ConditionReasonNATSSubscriptionActive, corev1.ConditionFalse, message)
	if isNatsSubReady {
		condition = eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscriptionActive,
			eventingv1alpha1.ConditionReasonNATSSubscriptionActive, corev1.ConditionTrue, message)
	}
	for _, c := range sub.Status.Conditions {
		var chosenCondition eventingv1alpha1.Condition
		if c.Type == condition.Type {
			if !conditionContained {
				if c.Status == condition.Status && c.Reason == condition.Reason && c.Message == condition.Message {
					// take the already present condition
					chosenCondition = c
				} else {
					// take the new given condition
					chosenCondition = condition
					conditionsUpdated = true
				}
				desiredConditions = append(desiredConditions, chosenCondition)
				conditionContained = true
			}
			// ignore all other conditions having the same type
			continue
		} else {
			// take the already present condition
			chosenCondition = c
		}
		desiredConditions = append(desiredConditions, chosenCondition)
	}
	if !conditionContained {
		desiredConditions = append(desiredConditions, condition)
		conditionsUpdated = true
	}

	if conditionsUpdated {
		sub.Status.Conditions = desiredConditions
		sub.Status.Ready = isNatsSubReady
		err := r.Client.Status().Update(ctx, sub, &client.UpdateOptions{})
		if err != nil {
			return errors.Wrapf(err, "failed to update subscription status")
		}
		r.Log.Info("successfully updated subscription status")
	}
	return nil
}

func (r *Reconciler) assertSinkValidity(sink string) error {
	_, err := url.ParseRequestURI(sink)
	return err
}

func (r *Reconciler) assertProtocolValidity(protocol string) error {
	if protocol != NATSProtocol {
		return fmt.Errorf("invalid protocol: %s", protocol)
	}
	return nil
}

func (r *Reconciler) syncInvalidSubscriptions(ctx context.Context) (ctrl.Result, error) {
	handler, _ := r.Backend.(*handlers.Nats)
	namespacedName := handler.GetInvalidSubscriptions()
	for _, v := range *namespacedName {
		r.Log.Info("found invalid subscription", "namespace", v.Namespace, "name", v.Name)
		sub := &eventingv1alpha1.Subscription{}
		err := r.Client.Get(ctx, v, sub)
		if err != nil {
			r.Log.Error(err, "failed to get invalid subscription", "namespace", v.Namespace, "name", v.Name)
			continue
		}
		// mark the subscription to be not ready, it will throw a new reconcile call
		if err := r.syncSubscriptionStatus(ctx, sub, false, "invalid subscription"); err != nil {
			r.Log.Error(err, "failed to save status for invalid subscription", "namespace", v.Namespace, "name", v.Name)
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}
