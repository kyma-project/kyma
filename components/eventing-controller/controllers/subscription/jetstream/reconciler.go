package jetstream

import (
	"context"
	"os"
	"reflect"

	equality "github.com/kyma-project/kyma/components/eventing-controller/controllers/subscription"
	"sigs.k8s.io/controller-runtime/pkg/event"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/eventtype"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/sink"
	corev1 "k8s.io/api/core/v1"

	"github.com/nats-io/nats.go"

	"github.com/kyma-project/kyma/components/eventing-controller/controllers/events"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"
	"github.com/pkg/errors"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"go.uber.org/zap"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	reconcilerName = "jetstream-subscription-reconciler"
)

var Finalizer = eventingv1alpha1.GroupVersion.Group

type Reconciler struct {
	client.Client
	ctx                 context.Context
	Backend             handlers.JetStreamBackend
	recorder            record.EventRecorder
	logger              *logger.Logger
	eventTypeCleaner    eventtype.Cleaner
	subsConfig          env.DefaultSubscriptionConfig
	sinkValidator       sink.Validator
	customEventsChannel chan event.GenericEvent
}

func NewReconciler(ctx context.Context, client client.Client, jsHandler handlers.JetStreamBackend, logger *logger.Logger,
	recorder record.EventRecorder, cleaner eventtype.Cleaner, subsCfg env.DefaultSubscriptionConfig, defaultSinkValidator sink.Validator) *Reconciler {
	reconciler := &Reconciler{
		Client:              client,
		ctx:                 ctx,
		Backend:             jsHandler,
		recorder:            recorder,
		logger:              logger,
		eventTypeCleaner:    cleaner,
		subsConfig:          subsCfg,
		sinkValidator:       defaultSinkValidator,
		customEventsChannel: make(chan event.GenericEvent),
	}
	if err := jsHandler.Initialize(reconciler.handleNatsConnClose); err != nil {
		logger.WithContext().Errorw("start reconciler failed", "name", reconcilerName, "error", err)
		panic(err)
	}
	return reconciler
}

// SetupUnmanaged creates a controller under the client control
func (r *Reconciler) SetupUnmanaged(mgr ctrl.Manager) error {
	ctru, err := controller.NewUnmanaged(reconcilerName, mgr, controller.Options{Reconciler: r})
	if err != nil {
		r.namedLogger().Errorw("create unmanaged controller failed", "name", reconcilerName, "error", err)
		return err
	}

	if err := ctru.Watch(&source.Kind{Type: &eventingv1alpha1.Subscription{}}, &handler.EnqueueRequestForObject{}); err != nil {
		r.namedLogger().Errorw("setup watch for subscriptions failed", "error", err)
		return err
	}

	if err := ctru.Watch(&source.Channel{Source: r.customEventsChannel}, &handler.EnqueueRequestForObject{}); err != nil {
		r.namedLogger().Errorw("setup watch for custom channel failed", "error", err)
		return err
	}

	go func(r *Reconciler, c controller.Controller) {
		if err := c.Start(r.ctx); err != nil {
			r.namedLogger().Errorw("start controller failed", "name", reconcilerName, "error", err)
			os.Exit(1)
		}
	}(r, ctru)

	return nil
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.namedLogger().Debugw("received subscription reconciliation request", "namespace", req.Namespace, "name", req.Name)

	actualSubscription := &eventingv1alpha1.Subscription{}
	// Ensure the object was not deleted in the meantime
	err := r.Client.Get(ctx, req.NamespacedName, actualSubscription)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Handle only the new subscription
	desiredSubscription := actualSubscription.DeepCopy()
	// Bind fields to logger
	log := utils.LoggerWithSubscription(r.namedLogger(), desiredSubscription)

	if isInDeletion(desiredSubscription) {
		// The object is being deleted
		err := r.handleSubscriptionDeletion(ctx, desiredSubscription, log)
		return ctrl.Result{}, err
	}

	// The object is not being deleted, so if it does not have our finalizer,
	// then lets add the finalizer and update the object.
	if !containsFinalizer(desiredSubscription) {
		err := r.addFinalizerToSubscription(desiredSubscription, log)
		return ctrl.Result{}, err
	}

	// update the cleanEventTypes and config values in the subscription status, if changed
	statusChanged, err := r.syncInitialStatus(desiredSubscription, log)
	if err != nil {
		log.Errorw("sync initial status failed", "error", err)
		if syncErr := r.syncSubscriptionStatus(ctx, desiredSubscription, statusChanged, err); err != nil {
			return ctrl.Result{}, syncErr
		}
		return ctrl.Result{}, err
	}

	// Check for valid sink
	if err := r.sinkValidator.Validate(desiredSubscription); err != nil {
		log.Errorw("sink URL validation failed", "error", err)
		if syncErr := r.syncSubscriptionStatus(ctx, desiredSubscription, statusChanged, err); err != nil {
			return ctrl.Result{}, syncErr
		}
		// No point in reconciling as the sink is invalid, return latest error to requeue the reconciliation request
		return ctrl.Result{}, err
	}

	// Synchronize Kyma subscription to JetStream backend
	syncErr := r.Backend.SyncSubscription(desiredSubscription)
	if syncErr != nil {
		log.Errorw("sync subscription failed", "error", syncErr)
		if err := r.syncSubscriptionStatus(ctx, desiredSubscription, statusChanged, syncErr); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, syncErr
	}

	// Update Subscription status
	if err := r.syncSubscriptionStatus(ctx, desiredSubscription, statusChanged, nil); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// handleNatsConnClose is called by NATS when the connection to the NATS server is closed. When it
// is called, the reconnect-attempts have exceeded the defined value.
// It forces reconciling the subscription to make sure the subscription is marked as not ready, until
// it is possible to connect to the NATS server again.
func (r *Reconciler) handleNatsConnClose(_ *nats.Conn) {
	r.namedLogger().Info("JetStream connection is closed and reconnect attempts are exceeded!")
	var subs eventingv1alpha1.SubscriptionList
	if err := r.Client.List(context.Background(), &subs); err != nil {
		// NATS reconnect attempts are exceeded, and we cannot reconcile subscriptions! If we ignore this,
		// there will be no future chance to retry connecting to NATS!
		panic(err)
	}
	r.enqueueReconciliationForSubscriptions(subs.Items)
}

// syncSubscriptionStatus syncs Subscription status and keeps the status up to date
func (r *Reconciler) syncSubscriptionStatus(ctx context.Context, sub *eventingv1alpha1.Subscription, updateStatus bool, error error) error {
	isNatsReady := error == nil
	readyStatusChanged := setSubReadyStatus(&sub.Status, isNatsReady)

	desiredConditions := initializeDesiredConditions()
	setConditionSubscriptionActive(desiredConditions, error)
	// check if the conditions are missing or changed
	if !equality.ConditionsEquals(sub.Status.Conditions, desiredConditions) {
		sub.Status.Conditions = desiredConditions
		updateStatus = true
	}

	// Update the status only if something needs to be updated
	if updateStatus || readyStatusChanged {
		err := r.Client.Status().Update(ctx, sub, &client.UpdateOptions{})
		if err != nil {
			events.Warn(r.recorder, sub, events.ReasonUpdateFailed, "Update Subscription status failed %s", sub.Name)
			return errors.Wrapf(err, "update subscription status failed")
		}
		events.Normal(r.recorder, sub, events.ReasonUpdate, "Update Subscription status succeeded %s", sub.Name)
	}
	return nil
}

// handleSubscriptionDeletion deletes the JetStream subscription and removes its finalizer if it is set.
func (r *Reconciler) handleSubscriptionDeletion(ctx context.Context, subscription *eventingv1alpha1.Subscription, log *zap.SugaredLogger) error {
	if utils.ContainsString(subscription.ObjectMeta.Finalizers, Finalizer) {
		if err := r.Backend.DeleteSubscription(subscription); err != nil {
			log.Errorw("delete JetStream subscription failed", "error", err)
			// if failed to delete the external dependency here, return with error
			// so that it can be retried
			return err
		}

		// remove our finalizer from the list and update it.
		subscription.ObjectMeta.Finalizers = utils.RemoveString(subscription.ObjectMeta.Finalizers, Finalizer)
		if err := r.Client.Update(ctx, subscription); err != nil {
			events.Warn(r.recorder, subscription, events.ReasonUpdateFailed, "Update Subscription failed %s", subscription.Name)
			log.Errorw("remove finalizer from subscription failed", "error", err)
			return err
		}
		log.Debug("remove finalizer from subscription succeeded")
	}
	return nil
}

// addFinalizerToSubscription appends the eventing finalizer to the subscription.
func (r *Reconciler) addFinalizerToSubscription(subscription *eventingv1alpha1.Subscription, log *zap.SugaredLogger) error {
	subscription.ObjectMeta.Finalizers = append(subscription.ObjectMeta.Finalizers, Finalizer)
	// TODO: check if this can be done at the end
	if err := r.Update(context.Background(), subscription); err != nil {
		log.Errorw("add finalizer to subscription failed", "error", err)
		return err
	}
	log.Debug("add finalizer to subscription succeeded")
	return nil
}

// syncInitialStatus keeps the latest cleanEventTypes and Config in the subscription
func (r *Reconciler) syncInitialStatus(subscription *eventingv1alpha1.Subscription, log *zap.SugaredLogger) (bool, error) {
	statusChanged := false
	cleanedSubjects, err := handlers.GetCleanSubjects(subscription, r.eventTypeCleaner)
	if err != nil {
		log.Errorw("getting clean subjects failed", "error", err)
		return false, err
	}
	if !reflect.DeepEqual(subscription.Status.CleanEventTypes, cleanedSubjects) {
		subscription.Status.CleanEventTypes = r.Backend.GetJetStreamSubjects(cleanedSubjects)
		statusChanged = true
	}
	subscriptionConfig := eventingv1alpha1.MergeSubsConfigs(subscription.Spec.Config, &r.subsConfig)
	if subscription.Status.Config == nil || !reflect.DeepEqual(subscriptionConfig, subscription.Status.Config) {
		subscription.Status.Config = subscriptionConfig
		statusChanged = true
	}
	return statusChanged, nil
}

func (r *Reconciler) enqueueReconciliationForSubscriptions(subs []eventingv1alpha1.Subscription) {
	r.namedLogger().Debug("enqueuing reconciliation request for all subscriptions")
	for i := range subs {
		r.customEventsChannel <- event.GenericEvent{Object: &subs[i]}
	}
}

// initializeDesiredConditions initializes the required conditions for the subscription status
func initializeDesiredConditions() []eventingv1alpha1.Condition {
	desiredConditions := make([]eventingv1alpha1.Condition, 0)
	condition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscriptionActive,
		eventingv1alpha1.ConditionReasonNATSSubscriptionNotActive, corev1.ConditionFalse, "")
	desiredConditions = append(desiredConditions, condition)
	return desiredConditions
}

// setConditionSubscriptionActive updates the ConditionSubscriptionActive condition if the error is nil
func setConditionSubscriptionActive(desiredConditions []eventingv1alpha1.Condition, error error) {
	for key, c := range desiredConditions {
		if c.Type == eventingv1alpha1.ConditionSubscriptionActive {
			if error == nil {
				desiredConditions[key].Status = corev1.ConditionTrue
				desiredConditions[key].Reason = eventingv1alpha1.ConditionReasonNATSSubscriptionActive
			} else {
				desiredConditions[key].Message = error.Error()
			}
		}
	}
}

// setSubReadyStatus returns true if the subscription ready status has changed
func setSubReadyStatus(desiredSubscriptionStatus *eventingv1alpha1.SubscriptionStatus, isReady bool) bool {
	if desiredSubscriptionStatus.Ready != isReady {
		desiredSubscriptionStatus.Ready = isReady
		return true
	}
	return false
}

// isInDeletion checks if the subscription needs to be deleted
func isInDeletion(subscription *eventingv1alpha1.Subscription) bool {
	return !subscription.ObjectMeta.DeletionTimestamp.IsZero()
}

// containsFinalizer checks if the subscription contains our Finalizer
func containsFinalizer(subscription *eventingv1alpha1.Subscription) bool {
	return utils.ContainsString(subscription.ObjectMeta.Finalizers, Finalizer)
}

func (r *Reconciler) namedLogger() *zap.SugaredLogger {
	return r.logger.WithContext().Named(reconcilerName)
}
