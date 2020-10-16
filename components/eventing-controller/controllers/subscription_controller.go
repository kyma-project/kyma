package controllers

import (
	"context"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
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
	Log       logr.Logger
	recorder  record.EventRecorder
	Scheme    *runtime.Scheme
	bebClient *handlers.Beb
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
	bebClient := &handlers.Beb{
		Log: log,
	}
	return &SubscriptionReconciler{
		Client:    client,
		Log:       log,
		recorder:  recorder,
		Scheme:    scheme,
		bebClient: bebClient,
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

	cachedSubscription := &eventingv1alpha1.Subscription{}

	ctx := context.TODO()
	result := ctrl.Result{}

	// Ensure the object was not deleted in the meantime
	if err := r.Get(ctx, req.NamespacedName, cachedSubscription); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// Handle only the new subscription
	subscription := cachedSubscription.DeepCopy()

	// Bind fields to logger
	log := r.Log.WithValues("kind", subscription.GetObjectKind().GroupVersionKind().Kind,
		"name", subscription.GetName(),
		"namespace", subscription.GetNamespace(),
		"version", subscription.GetGeneration(),
	)

	// Ensure the finalizer is set
	if err := r.syncFinalizer(subscription, &result, ctx, log); err != nil {
		log.Error(err, "error while syncing finalizer")
		return result, err
	}
	if result.Requeue {
		return result, nil
	}

	if !r.isInDeletion(subscription) {
		if err := r.syncInitialStatus(subscription, &result, ctx, log); err != nil {
			log.Error(err, "error while syncing status")
			return result, err
		}
	}
	if result.Requeue {
		return result, nil
	}

	// TODO sync the Subscription CR with he corresponding APIRule.
	// Expose the webhook

	// Sync the BEB Subscription with the Subscription CR
	if err := r.syncBEBSubscription(subscription, &result, ctx, log); err != nil {
		log.Error(err, "error while syncing BEB subscription")
		return result, err
	}

	return result, nil
}

// syncFinalizer sets the finalizer in the Subscription
func (r *SubscriptionReconciler) syncFinalizer(subscription *eventingv1alpha1.Subscription, result *ctrl.Result, ctx context.Context, logger logr.Logger) error {
	// Check if finalizer is already set
	if r.isFinalizerSet(subscription) {
		return nil
	}

	// Add Finalizer if not in deletion mode
	if !r.isInDeletion(subscription) {
		if err := r.addFinalizer(subscription, ctx, logger); err != nil {
			return err
		}
		result.Requeue = true
	}

	return nil
}

func (r *SubscriptionReconciler) syncBEBSubscription(subscription *eventingv1alpha1.Subscription, result *ctrl.Result, ctx context.Context, logger logr.Logger) error {
	logger.Info("Syncing subscription with BEB")
	// TODO: get beb credentials from secret

	r.bebClient.Initialize()

	// TODO: react on finalizer
	if r.isInDeletion(subscription) {
		logger.Info("Deleting BEB subscription")
		if err := r.bebClient.DeleteBebSubscription(subscription); err != nil {
			return err
		}
		if err := r.removeFinalizer(subscription, result, ctx, logger); err != nil {
			return err
		}
		return nil
	}

	ev2Hash := subscription.Status.Ev2hash
	emsHash := subscription.Status.Emshash
	if newEv2Hash, newEmsHash, err := r.bebClient.SyncBebSubscription(subscription, ev2Hash, emsHash); err != nil {
		logger.Error(err, "Update BEB subscription failed")
		return err
	} else {
		if ev2Hash != newEv2Hash {
			subscription.Status.Ev2hash = newEv2Hash
		}
		if emsHash != newEmsHash {
			subscription.Status.Emshash = newEmsHash
		}
	}
	// OK
	if !subscription.Status.IsConditionSubscribed() {
		condition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscribed, "Successfully synchronized with BEB subscription", corev1.ConditionTrue)
		if err := r.updateStatus(subscription, condition, ctx, logger); err != nil {
			return err
		}
		result.Requeue = false // the last call in reconcile(), don't force a new reconcile() yet.
	}

	return nil
}

// syncInitialStatus determines the desires initial status and updates it accordingly (if conditions changed)
func (r *SubscriptionReconciler) syncInitialStatus(subscription *eventingv1alpha1.Subscription, result *ctrl.Result, ctx context.Context, logger logr.Logger) error {
	currentStatus := subscription.Status

	expectedStatus := eventingv1alpha1.SubscriptionStatus{}
	expectedStatus.InitializeConditions()

	// case: conditions are already initialized
	if len(currentStatus.Conditions) >= len(expectedStatus.Conditions) {
		return nil
	}

	subscription.Status = expectedStatus
	if err := r.Status().Update(ctx, subscription); err != nil {
		return err
	}
	result.Requeue = true

	return nil
}

func (r *SubscriptionReconciler) updateStatus(subscription *eventingv1alpha1.Subscription, condition eventingv1alpha1.Condition, ctx context.Context, logger logr.Logger) error {

	// compile list of desired conditions
	desiredConditions := make([]eventingv1alpha1.Condition, 0)
	for _, c := range subscription.Status.Conditions {
		if c.Type == condition.Type {
			// take given condition
			desiredConditions = append(desiredConditions, condition)
		} else {
			// take already present condition
			desiredConditions = append(desiredConditions, c)
		}
	}

	// prevent unnecessary updates
	if isEqualConditions(subscription.Status.Conditions, desiredConditions) {
		return nil
	}

	// update the status
	subscription.Status.Conditions = desiredConditions
	if err := r.Status().Update(ctx, subscription); err != nil {
		return err
	}
	r.emitConditionEvent(subscription, condition)

	return nil
}

// emitConditionEvent emits a kubernetes event and sets the event type based on the Condition status
func (r *SubscriptionReconciler) emitConditionEvent(subscription *eventingv1alpha1.Subscription, condition eventingv1alpha1.Condition) {
	eventType := corev1.EventTypeNormal
	if condition.Status == corev1.ConditionFalse {
		eventType = corev1.EventTypeWarning
	}
	// TODO:
	reason := "todo"
	message := "todo"
	r.recorder.Event(subscription, eventType, reason, message)
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

func (r *SubscriptionReconciler) removeFinalizer(subscription *eventingv1alpha1.Subscription, result *ctrl.Result, ctx context.Context, logger logr.Logger) error {
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
	result.Requeue = true
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
