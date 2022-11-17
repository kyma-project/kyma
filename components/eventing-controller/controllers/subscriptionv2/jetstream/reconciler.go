package jetstream

import (
	"context"
	"reflect"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/components/eventing-controller/controllers/events"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"
	sinkv2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/sink/v2"
	backendutilsv2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/utils/v2"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	jetstream "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/jetstreamv2"
)

const (
	reconcilerName  = "jetstream-subscription-v2-reconciler"
	requeueDuration = 10 * time.Second
)

type Reconciler struct {
	client.Client
	ctx                 context.Context
	Backend             jetstream.Backend
	recorder            record.EventRecorder
	logger              *logger.Logger
	cleaner             cleaner.Cleaner
	sinkValidator       sinkv2.Validator
	customEventsChannel chan event.GenericEvent
}

func NewReconciler(ctx context.Context, client client.Client, jsHandler jetstream.Backend, logger *logger.Logger,
	recorder record.EventRecorder, cleaner cleaner.Cleaner, defaultSinkValidator sinkv2.Validator) *Reconciler {
	reconciler := &Reconciler{
		Client:              client,
		ctx:                 ctx,
		Backend:             jsHandler,
		recorder:            recorder,
		logger:              logger,
		cleaner:             cleaner,
		sinkValidator:       defaultSinkValidator,
		customEventsChannel: make(chan event.GenericEvent),
	}
	if err := jsHandler.Initialize(reconciler.handleNatsConnClose); err != nil {
		logger.WithContext().Errorw("Failed to start reconciler", "name", reconcilerName, "error", err)
		panic(err)
	}
	return reconciler
}

// SetupUnmanaged creates a controller under the client control.
func (r *Reconciler) SetupUnmanaged(mgr ctrl.Manager) error {
	ctru, err := controller.NewUnmanaged(reconcilerName, mgr, controller.Options{Reconciler: r})
	if err != nil {
		r.namedLogger().Errorw("Failed to create unmanaged controller", "error", err)
		return err
	}

	if err := ctru.Watch(&source.Kind{Type: &eventingv1alpha2.Subscription{}},
		&handler.EnqueueRequestForObject{}); err != nil {
		r.namedLogger().Errorw("Failed to setup watch for subscriptions", "error", err)
		return err
	}

	if err := ctru.Watch(&source.Channel{Source: r.customEventsChannel},
		&handler.EnqueueRequestForObject{}); err != nil {
		r.namedLogger().Errorw("Failed to setup watch for custom channel", "error", err)
		return err
	}

	go func(r *Reconciler, c controller.Controller) {
		if err := c.Start(r.ctx); err != nil {
			r.namedLogger().Fatalw("Failed to start controller", "error", err)
		}
	}(r, ctru)

	return nil
}

// +kubebuilder:rbac:groups=eventing.kyma-project.io,resources=subscriptions,verbs=get;list;watch;create;update;patch;delete //nolint
// +kubebuilder:rbac:groups=eventing.kyma-project.io,resources=subscriptions/status,verbs=get;update;patch
// Generate required RBAC to emit kubernetes events in the controller.
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.namedLogger().Debugw("Received subscription v1alpha2 reconciliation request",
		"namespace", req.Namespace, "name", req.Name)

	actualSubscription := &eventingv1alpha2.Subscription{}
	// Ensure the object was not deleted in the meantime
	err := r.Client.Get(ctx, req.NamespacedName, actualSubscription)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Handle only the new subscription
	desiredSubscription := actualSubscription.DeepCopy()
	// Bind fields to logger
	log := backendutilsv2.LoggerWithSubscription(r.namedLogger(), desiredSubscription)

	if isInDeletion(desiredSubscription) {
		// The object is being deleted
		if delErr := r.handleSubscriptionDeletion(ctx, desiredSubscription, log); delErr != nil {
			log.Errorw("Failed to delete the Subscription", "error", delErr)
			if syncErr := r.syncSubscriptionStatus(ctx, desiredSubscription, false, delErr); syncErr != nil {
				return ctrl.Result{}, syncErr
			}
			return ctrl.Result{}, delErr
		}
		return ctrl.Result{}, nil
	}

	// The object is not being deleted, so if it does not have our finalizer,
	// then lets add the finalizer and update the object.
	if !containsFinalizer(desiredSubscription) {
		if finalizerErr := r.addFinalizerToSubscription(desiredSubscription, log); finalizerErr != nil {
			log.Errorw("Failed to add finalizer to Subscription", "error", finalizerErr)
			if syncErr := r.syncSubscriptionStatus(ctx, desiredSubscription, false, finalizerErr); syncErr != nil {
				return ctrl.Result{}, syncErr
			}
			return ctrl.Result{}, finalizerErr
		}
		return ctrl.Result{}, nil
	}

	// update the cleanEventTypes and config values in the subscription status, if changed
	statusChanged := r.syncInitialStatus(desiredSubscription)

	// Check for valid sink
	if validateErr := r.sinkValidator.Validate(desiredSubscription); validateErr != nil {
		if syncErr := r.syncSubscriptionStatus(ctx, desiredSubscription, statusChanged, validateErr); syncErr != nil {
			return ctrl.Result{}, syncErr
		}
		// No point in reconciling as the sink is invalid, return latest error to requeue the reconciliation request
		return ctrl.Result{}, validateErr
	}

	// Synchronize Kyma subscription to JetStream backend
	if syncSubErr := r.Backend.SyncSubscription(desiredSubscription); syncSubErr != nil {
		result := ctrl.Result{}
		if syncErr := r.syncSubscriptionStatus(ctx, desiredSubscription, statusChanged, syncSubErr); syncErr != nil {
			return result, syncErr
		}
		// Requeue the Request to reconcile it again if there are no NATS Subscriptions synced
		if errors.Is(syncSubErr, jetstream.ErrMissingSubscription) {
			result = ctrl.Result{RequeueAfter: requeueDuration}
			syncSubErr = nil
		}
		return result, syncSubErr
	}

	// Update Subscription status
	if syncErr := r.syncSubscriptionStatus(ctx, desiredSubscription, statusChanged, nil); syncErr != nil {
		return ctrl.Result{}, syncErr
	}
	return ctrl.Result{}, nil
}

// handleNatsConnClose is called by NATS when the connection to the NATS server is closed. When it
// is called, the reconnect-attempts have exceeded the defined value.
// It forces reconciling the subscription to make sure the subscription is marked as not ready, until
// it is possible to connect to the NATS server again.
func (r *Reconciler) handleNatsConnClose(_ *nats.Conn) {
	r.namedLogger().Info("JetStream connection is closed and reconnect attempts are exceeded!")
	var subs eventingv1alpha2.SubscriptionList
	if err := r.Client.List(context.Background(), &subs); err != nil {
		// NATS reconnect attempts are exceeded, and we cannot reconcile subscriptions! If we ignore this,
		// there will be no future chance to retry connecting to NATS!
		panic(err)
	}
	r.enqueueReconciliationForSubscriptions(subs.Items)
}

// enqueueReconciliationForSubscriptions adds the subscriptions to the customEventsChannel
// which is being watched by the controller.
func (r *Reconciler) enqueueReconciliationForSubscriptions(subs []eventingv1alpha2.Subscription) {
	r.namedLogger().Debug("Enqueuing reconciliation request for all subscriptions")
	for i := range subs {
		r.customEventsChannel <- event.GenericEvent{Object: &subs[i]}
	}
}

// syncSubscriptionStatus syncs Subscription status and keeps the status up to date.
func (r *Reconciler) syncSubscriptionStatus(ctx context.Context,
	sub *eventingv1alpha2.Subscription, updateStatus bool, err error) error {
	isNatsReady := err == nil
	readyStatusChanged := setSubReadyStatus(&sub.Status, isNatsReady)

	desiredConditions := initializeDesiredConditions()
	setConditionSubscriptionActive(desiredConditions, err)
	// check if the conditions are missing or changed
	if !eventingv1alpha2.ConditionsEquals(sub.Status.Conditions, desiredConditions) {
		sub.Status.Conditions = desiredConditions
		updateStatus = true
	}

	// Update the status only if something needs to be updated
	if updateStatus || readyStatusChanged {
		if updateErr := r.Client.Status().Update(ctx, sub, &client.UpdateOptions{}); updateErr != nil {
			events.Warn(r.recorder, sub, events.ReasonUpdateFailed, "Update Subscription status failed %s", sub.Name)
			return utils.MakeError(errFailedToUpdateStatus, updateErr)
		}
		events.Normal(r.recorder, sub, events.ReasonUpdate, "Update Subscription status succeeded %s", sub.Name)
	}
	return nil
}

// handleSubscriptionDeletion deletes the JetStream subscription and removes its finalizer if it is set.
func (r *Reconciler) handleSubscriptionDeletion(ctx context.Context,
	subscription *eventingv1alpha2.Subscription, log *zap.SugaredLogger) error {
	if utils.ContainsString(subscription.ObjectMeta.Finalizers, eventingv1alpha2.Finalizer) {
		if err := r.Backend.DeleteSubscription(subscription); err != nil {
			// if failed to delete the external dependency here, return with error
			// so that it can be retried
			return utils.MakeError(errFailedToDeleteSub, err)
		}

		// remove our finalizer from the list and update it.
		subscription.ObjectMeta.Finalizers = utils.RemoveString(subscription.ObjectMeta.Finalizers,
			eventingv1alpha2.Finalizer)
		if err := r.Client.Update(ctx, subscription); err != nil {
			events.Warn(r.recorder, subscription, events.ReasonUpdateFailed, "Update Subscription failed %s", subscription.Name)
			return utils.MakeError(errFailedToRemoveFinalizer, err)
		}
		log.Debug("Removed finalizer from subscription")
	}
	return nil
}

// addFinalizerToSubscription appends the eventing finalizer to the subscription.
func (r *Reconciler) addFinalizerToSubscription(sub *eventingv1alpha2.Subscription, log *zap.SugaredLogger) error {
	sub.ObjectMeta.Finalizers = append(sub.ObjectMeta.Finalizers, eventingv1alpha2.Finalizer)
	// to avoid a dangling subscription, we update the subscription as soon as the finalizer is added to it
	if err := r.Client.Update(context.Background(), sub); err != nil {
		return utils.MakeError(errFailedToAddFinalizer, err)
	}
	log.Debug("Added finalizer to subscription")
	return nil
}

// syncInitialStatus keeps the latest cleaned EventTypes and jetStreamTypes in the subscription.
func (r *Reconciler) syncInitialStatus(subscription *eventingv1alpha2.Subscription) bool {
	statusChanged := false
	cleanedTypes := jetstream.GetCleanEventTypes(subscription, r.cleaner)
	if !reflect.DeepEqual(subscription.Status.Types, cleanedTypes) {
		subscription.Status.Types = cleanedTypes
		statusChanged = true
	}
	jsSubjects := r.Backend.GetJetStreamSubjects(subscription.Spec.Source,
		jetstream.GetCleanEventTypesFromEventTypes(cleanedTypes),
		subscription.Spec.TypeMatching)
	jsTypes := jetstream.GetBackendJetStreamTypes(subscription, jsSubjects)
	if !reflect.DeepEqual(subscription.Status.Backend.Types, jsTypes) {
		subscription.Status.Backend.Types = jsTypes
		statusChanged = true
	}
	return statusChanged
}

func (r *Reconciler) namedLogger() *zap.SugaredLogger {
	return r.logger.WithContext().Named(reconcilerName)
}
