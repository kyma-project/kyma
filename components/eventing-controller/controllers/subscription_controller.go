package controllers

import (
	"context"

	"github.com/go-logr/logr"
	// TODO: use different package
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
)

// SubscriptionReconciler reconciles a Subscription object
type SubscriptionReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

const (
	// TODO:
	finalizerName = "todo"
)

// +kubebuilder:rbac:groups=eventing.kyma-project.io,resources=subscriptions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=eventing.kyma-project.io,resources=subscriptions/status,verbs=get;update;patch

func (r *SubscriptionReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("subscription", req.NamespacedName)

	// TODO: pass logger via constructor
	// logger := r.Log.WithName("controllers").WithName("beb")

	subscription := &eventingv1alpha1.Subscription{}
	ctx := context.TODO()

	// Ensure the object was not deleted in the meantime
	if err := r.Client.Get(ctx, req.NamespacedName, subscription); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Bind fields to logger
	log := r.Log.WithValues("kind", subscription.GetObjectKind().GroupVersionKind().Kind,
		"name", subscription.GetName(),
		"namespace", subscription.GetNamespace(),
		"version", subscription.GetGeneration(),
	)

	// Ensure the finalizer is set
	if err := r.syncFinalizer(subscription, ctx, log); err != nil {
		return ctrl.Result{}, err
	}

	// Sync the BEB Subscription with the Subscription CR
	if err := r.syncBEBSubscription(subscription, ctx, log); err != nil {
		return ctrl.Result{}, err
	}

	log.Info("reconciled obj", req.Namespace, req.Name)

	return ctrl.Result{}, nil
}

// syncFinalizer sets the finalizer in the Subscription
func (r *SubscriptionReconciler) syncFinalizer(subscription *eventingv1alpha1.Subscription, ctx context.Context, logger logr.Logger) error {
	// Check if finalizer is already set
	if r.isFinalizerSet(subscription) {
		return nil
	}

	// Add Finalizer
	return r.addFinalizer(subscription, ctx)
}


func (r *SubscriptionReconciler) syncBEBSubscription(subscription *eventingv1alpha1.Subscription, ctx context.Context, logger logr.Logger) error {
	logger.Info("Syncing subscription with BEB")
	// TODO: get beb credentials from secret
	// TODO: CRUD BEB subscription

	// TODO: react on finalizer
	if r.isInDeletion(subscription) {
		logger.Info("Deleting BEB subscription")
		if err := r.removeFinalizer(subscription, ctx); err != nil {
			return err
		}
		// TODO: delete BEB subscription
		return nil
	}

	logger.Info("Updating BEB subscription")

	return nil
}

func (r *SubscriptionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&eventingv1alpha1.Subscription{}).
		Complete(r)
}
func (r *SubscriptionReconciler) addFinalizer(subscription *eventingv1alpha1.Subscription, ctx context.Context) error {
	subscription.ObjectMeta.Finalizers = append(subscription.ObjectMeta.Finalizers, finalizerName)
	if err := r.Update(ctx, subscription); err != nil {
		return errors.Wrapf(err, "error while adding Finalizer with name: %s", finalizerName)
	}
	return nil
}

func (r *SubscriptionReconciler) removeFinalizer(subscription *eventingv1alpha1.Subscription, ctx context.Context) error {
	var finalizers []string

	// Build finalizer list without the one the controller owns
	for _, finalizer := range subscription.ObjectMeta.Finalizers {
		if finalizer == finalizerName {
			continue
		}
		finalizers = append(finalizers, finalizer)
	}

	subscription.ObjectMeta.Finalizers = finalizers
	if err := r.Update(ctx, subscription); err != nil {
		return errors.Wrapf(err, "error while removing Finalizer with name: %s", finalizerName)
	}
	return nil
}

// isFinalizerSet checks if a finalizer is set on the Subscription which belongs to this controller
func (r *SubscriptionReconciler) isFinalizerSet(subscription *eventingv1alpha1.Subscription) bool {
	// Check if finalizer is already set
	for _, finalizer := range subscription.ObjectMeta.Finalizers {
		if finalizer == finalizerName {
			return true
		}
	}
	return false
}

// isInDeletion checks if the Subscription shall be deleted
func (r *SubscriptionReconciler) isInDeletion(subscription *eventingv1alpha1.Subscription) bool {
	return !subscription.DeletionTimestamp.IsZero()
}
