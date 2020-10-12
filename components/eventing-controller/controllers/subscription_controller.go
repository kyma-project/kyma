package controllers

import (
	"context"
	"reflect"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"

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
	Log      logr.Logger
	recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

// TODO: emit events
// TODO: use additional printer columns: https://book.kubebuilder.io/reference/generating-crd.html#additional-printer-columns

var (
	FinalizerName = eventingv1alpha1.GroupVersion.Group
)

func NewSubscriptionReconciler(
	client client.Client,
	log logr.Logger,
	recorder record.EventRecorder,
	scheme *runtime.Scheme,
) *SubscriptionReconciler {
	return &SubscriptionReconciler{
		Client:   client,
		Log:      log,
		recorder: recorder,
		Scheme:   scheme,
	}
}

// +kubebuilder:rbac:groups=eventing.kyma-project.io,resources=subscriptions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=eventing.kyma-project.io,resources=subscriptions/status,verbs=get;update;patch

// Generate required RBAC to emit kubernetes events in the controller
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
// Source: https://book-v1.book.kubebuilder.io/beyond_basics/creating_events.html

func (r *SubscriptionReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("subscription", req.NamespacedName)

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

	// TODO:
	r.recorder.Event(subscription, corev1.EventTypeNormal, "foo", "bar")

	// Ensure the finalizer is set
	if err := r.syncFinalizer(subscription, ctx, log); err != nil {
		log.Error(err, "error while syncing finalizer")
		return ctrl.Result{}, err
	}

	// Sync the BEB Subscription with the Subscription CR
	if err := r.syncBEBSubscription(subscription, ctx, log); err != nil {
		log.Error(err, "error while syncing BEB subscription")
		return ctrl.Result{}, err
	}

	if !r.isInDeletion(subscription) {
		if err := r.syncStatus(subscription, ctx, log); err != nil {
			log.Error(err, "error while syncing status")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// syncFinalizer sets the finalizer in the Subscription
func (r *SubscriptionReconciler) syncFinalizer(subscription *eventingv1alpha1.Subscription, ctx context.Context, logger logr.Logger) error {
	// Check if finalizer is already set
	if r.isFinalizerSet(subscription) {
		return nil
	}

	// Add Finalizer if not in deletion mode
	if !r.isInDeletion(subscription) {
		return r.addFinalizer(subscription, ctx, logger)
	}

	return nil
}

func (r *SubscriptionReconciler) syncBEBSubscription(subscription *eventingv1alpha1.Subscription, ctx context.Context, logger logr.Logger) error {
	// TODO: get beb credentials from secret
	// TODO: CRUD BEB subscription

	// TODO: react on finalizer
	if r.isInDeletion(subscription) {
		logger.Info("Deleting BEB subscription")
		if err := r.removeFinalizer(subscription, ctx, logger); err != nil {
			return err
		}
		// TODO: delete BEB subscription
		return nil
	}

	logger.Info("Creating BEB subscription")

	return nil
}

// syncStatus determines the desires status and updates it
func (r *SubscriptionReconciler) syncStatus(subscription *eventingv1alpha1.Subscription, ctx context.Context, logger logr.Logger) error {
	currentStatus := subscription.Status
	currentStatus.InitializeConditions()
	expectedStatus := r.computeStatus(subscription, ctx, logger)

	if reflect.DeepEqual(currentStatus, expectedStatus) {
		return nil
	}

	subscription.Status = *expectedStatus
	if err := r.Status().Update(ctx, subscription); err != nil {
		return err
	}

	return nil
}

// computeStatus computes the status of the Subscription
func (r *SubscriptionReconciler) computeStatus(subscription *eventingv1alpha1.Subscription, ctx context.Context, logger logr.Logger) *eventingv1alpha1.SubscriptionStatus {
	status := subscription.Status.DeepCopy()
	status.InitializeConditions()
	// TODO: set status for BEB subscription

	return status
}

func conditionEquals(c1 eventingv1alpha1.Condition, c2 eventingv1alpha1.Condition) bool {
	if c1.Type == c2.Type && c1.Reason == c2.Reason && c1.Status == c2.Status {
		return true
	}
	return false
}

// TODO: create own file for status handling, this is more a init method (constructor)
func makeCondition(conditionType eventingv1alpha1.ConditionType, reason eventingv1alpha1.ConditionReason, status corev1.ConditionStatus) eventingv1alpha1.Condition {
	return eventingv1alpha1.Condition{
		Type:               conditionType,
		Status:             status,
		LastTransitionTime: metav1.Time{},
		Reason:             reason,
		// TODO:
		Message: "",
	}

}

func (r *SubscriptionReconciler) updateStatus(subscription *eventingv1alpha1.Subscription, ctx context.Context) error {
	eventType := corev1.EventTypeNormal
	reason := "todo"
	message := "todo"

	if err := r.Client.Status().Update(ctx, subscription); err != nil {
		return err
	}
	r.recorder.Event(subscription, eventType, reason, message)

	return nil
}

// TODO: do not update when nothing changed

func (r *SubscriptionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&eventingv1alpha1.Subscription{}).
		Complete(r)
}
func (r *SubscriptionReconciler) addFinalizer(subscription *eventingv1alpha1.Subscription, ctx context.Context, logger logr.Logger) error {
	subscription.ObjectMeta.Finalizers = append(subscription.ObjectMeta.Finalizers, FinalizerName)
	logger.V(1).Info("Adding finalizer")
	if err := r.Update(ctx, subscription); err != nil {
		return errors.Wrapf(err, "error while adding Finalizer with name: %s", FinalizerName)
	}
	logger.V(1).Info("Added finalizer")
	return nil
}

func (r *SubscriptionReconciler) removeFinalizer(subscription *eventingv1alpha1.Subscription, ctx context.Context, logger logr.Logger) error {
	var finalizers []string

	// Build finalizer list without the one the controller owns
	for _, finalizer := range subscription.ObjectMeta.Finalizers {
		if finalizer == FinalizerName {
			continue
		}
		finalizers = append(finalizers, finalizer)
	}

	logger.V(1).Info("Removing finalizer")
	subscription.ObjectMeta.Finalizers = finalizers
	if err := r.Update(ctx, subscription); err != nil {
		return errors.Wrapf(err, "error while removing Finalizer with name: %s", FinalizerName)
	}
	logger.V(1).Info("Removed finalizer")
	return nil
}

// isFinalizerSet checks if a finalizer is set on the Subscription which belongs to this controller
func (r *SubscriptionReconciler) isFinalizerSet(subscription *eventingv1alpha1.Subscription) bool {
	// Check if finalizer is already set
	for _, finalizer := range subscription.ObjectMeta.Finalizers {
		if finalizer == FinalizerName {
			return true
		}
	}
	return false
}

// isInDeletion checks if the Subscription shall be deleted
func (r *SubscriptionReconciler) isInDeletion(subscription *eventingv1alpha1.Subscription) bool {
	return !subscription.DeletionTimestamp.IsZero()
}
