package nats

import (
	"context"
	"reflect"

	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/controllers/events"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/eventtype"
	nats2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/nats"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/nats/core"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/sink"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"
)

// Reconciler reconciles a Subscription object
type Reconciler struct {
	client.Client
	ctx              context.Context
	Backend          core.NatsBackend
	logger           *logger.Logger
	recorder         record.EventRecorder
	subsConfig       env.DefaultSubscriptionConfig
	eventTypeCleaner eventtype.Cleaner
	sinkValidator    sink.Validator
	// This channel is used to enqueue a reconciliation request for a subscription
	customEventsChannel chan event.GenericEvent
}

var (
	Finalizer = eventingv1alpha1.GroupVersion.Group
)

const (
	natsFirstInstanceName = "eventing-nats-1" // natsFirstInstanceName the name of first instance of NATS cluster
	natsNamespace         = "kyma-system"     // natsNamespace of NATS cluster
	reconcilerName        = "nats-subscription-reconciler"
)

func NewReconciler(ctx context.Context, client client.Client, natsHandler core.NatsBackend, cleaner eventtype.Cleaner,
	logger *logger.Logger, recorder record.EventRecorder, subsCfg env.DefaultSubscriptionConfig, defaultSinkValidator sink.Validator) *Reconciler {
	reconciler := &Reconciler{
		ctx:                 ctx,
		Backend:             natsHandler,
		Client:              client,
		logger:              logger,
		recorder:            recorder,
		subsConfig:          subsCfg,
		eventTypeCleaner:    cleaner,
		sinkValidator:       defaultSinkValidator,
		customEventsChannel: make(chan event.GenericEvent),
	}
	if err := natsHandler.Initialize(reconciler.handleNatsConnClose); err != nil {
		logger.WithContext().Errorw("Failed to start reconciler", "name", reconcilerName, "error", err)
		panic(err)
	}

	return reconciler
}

// handleNatsConnClose is called by NATS when the connection to the NATS server is closed. When it
// is called, the reconnect-attempts have exceeded the defined value.
// It forces reconciling the subscription to make sure the subscription is marked as not ready, until
// it is possible to connect to the NATS server again.
// See https://github.com/kyma-project/kyma/issues/12930
func (r *Reconciler) handleNatsConnClose(_ *nats.Conn) {
	r.namedLogger().Info("NATS connection is closed and reconnect attempts are exceeded")
	subs, err := r.getAllKymaSubscriptions(context.Background())
	if err != nil {
		// NATS reconnect attempts are exceeded, and we cannot reconcile subscriptions! If we ignore this,
		// there will be no future chance to retry connecting to NATS!
		panic(err)
	}
	r.enqueueReconciliationForSubscriptions(subs)
}

func (r *Reconciler) getAllKymaSubscriptions(ctx context.Context) ([]eventingv1alpha1.Subscription, error) {
	var subs eventingv1alpha1.SubscriptionList
	if err := r.Client.List(ctx, &subs); err != nil {
		return nil, err
	}
	return subs.Items, nil
}

// SetupUnmanaged creates a controller under the client control
func (r *Reconciler) SetupUnmanaged(mgr ctrl.Manager) error {
	ctru, err := controller.NewUnmanaged(reconcilerName, mgr, controller.Options{Reconciler: r})
	if err != nil {
		r.namedLogger().Errorw("Failed to create unmanaged controller", "error", err)
		return err
	}

	if err := ctru.Watch(&source.Kind{Type: &eventingv1alpha1.Subscription{}}, &handler.EnqueueRequestForObject{}); err != nil {
		r.namedLogger().Errorw("Failed to setup watch for subscriptions", "error", err)
		return err
	}

	p := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			if e.Object.GetName() == natsFirstInstanceName && e.Object.GetNamespace() == natsNamespace {
				return true
			}
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			if e.Object.GetName() == natsFirstInstanceName && e.Object.GetNamespace() == natsNamespace {
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
		r.namedLogger().Errorw("Failed to setup watch for NATS server", "pod", natsFirstInstanceName, "error", err)
		return err
	}

	if err := ctru.Watch(&source.Channel{Source: r.customEventsChannel}, &handler.EnqueueRequestForObject{}); err != nil {
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

// +kubebuilder:rbac:groups=eventing.kyma-project.io,resources=subscriptions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=eventing.kyma-project.io,resources=subscriptions/status,verbs=get;update;patch
// Generate required RBAC to emit kubernetes events in the controller.
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
// Generate required RBAC to watch the NATS pods.
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;delete
// Generated required RBAC to list Applications (required by event type cleaner).
// +kubebuilder:rbac:groups="applicationconnector.kyma-project.io",resources=applications,verbs=get;list;watch

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if req.Name == natsFirstInstanceName && req.Namespace == natsNamespace {
		r.namedLogger().Debugw("Received watch request", "namespace", req.Namespace, "name", req.Name)
		return r.syncInvalidSubscriptions(ctx)
	}

	r.namedLogger().Debugw("Received subscription reconciliation request", "namespace", req.Namespace, "name", req.Name)

	cachedSubscription := &eventingv1alpha1.Subscription{}

	// Ensure the object was not deleted in the meantime
	err := r.Client.Get(ctx, req.NamespacedName, cachedSubscription)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Handle only the new subscription
	subscription := cachedSubscription.DeepCopy()

	// Bind fields to logger
	log := utils.LoggerWithSubscription(r.namedLogger(), subscription)

	if !subscription.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is being deleted
		err := r.handleSubscriptionDeletion(ctx, subscription, log)
		return ctrl.Result{}, err
	}

	// The object is not being deleted, so if it does not have our finalizer,
	// then lets add the finalizer and update the object. This is equivalent to
	// registering our finalizer.
	if !utils.ContainsString(subscription.ObjectMeta.Finalizers, Finalizer) {
		err := r.addFinalizerToSubscription(subscription, log)
		return ctrl.Result{}, err
	}

	// update the cleanEventTypes and config values in the subscription, if changed
	statusChanged, err := r.syncInitialStatus(subscription, log)
	if err != nil {
		log.Errorw("Failed to sync initial status", "error", err)
		if syncErr := r.syncSubscriptionStatus(ctx, subscription, false, statusChanged, err.Error()); syncErr != nil {
			return ctrl.Result{}, syncErr
		}
		return ctrl.Result{}, err
	}

	// Check for valid sink
	if err := r.sinkValidator.Validate(subscription); err != nil {
		log.Errorw("Failed to validate sink URL", "error", err)
		if syncErr := r.syncSubscriptionStatus(ctx, subscription, false, statusChanged, err.Error()); syncErr != nil {
			return ctrl.Result{}, syncErr
		}
		// No point in reconciling as the sink is invalid, return latest error to requeue the reconciliation request
		return ctrl.Result{}, err
	}

	// Synchronize Kyma subscription to NATS backend
	syncErr := r.Backend.SyncSubscription(subscription)
	if syncErr != nil {
		log.Errorw("Failed to sync subscription to NATS backend", "error", syncErr)
		if err := r.syncSubscriptionStatus(ctx, subscription, false, statusChanged, syncErr.Error()); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, syncErr
	}
	log.Debug("Creation of NATS subscriptions succeeded")

	// Update status
	if err := r.syncSubscriptionStatus(ctx, subscription, true, statusChanged, ""); err != nil {
		return checkIsConflict(err)
	}

	return ctrl.Result{}, nil
}

// syncInitialStatus keeps the latest cleanEventTypes and Config in the subscription
func (r *Reconciler) syncInitialStatus(subscription *eventingv1alpha1.Subscription, log *zap.SugaredLogger) (bool, error) {
	statusChanged := false
	cleanEventTypes, err := nats2.GetCleanSubjects(subscription, r.eventTypeCleaner)
	if err != nil {
		log.Errorw("Failed to get clean subject", "error", err)
		subscription.Status.InitializeCleanEventTypes()
		return true, err
	}
	if !reflect.DeepEqual(subscription.Status.CleanEventTypes, cleanEventTypes) {
		subscription.Status.CleanEventTypes = cleanEventTypes
		statusChanged = true
	}
	subscriptionConfig := eventingv1alpha1.MergeSubsConfigs(subscription.Spec.Config, &r.subsConfig)
	if subscription.Status.Config == nil || !reflect.DeepEqual(subscriptionConfig, subscription.Status.Config) {
		subscription.Status.Config = subscriptionConfig
		statusChanged = true
	}
	if subscription.Status.CleanEventTypes == nil {
		subscription.Status.InitializeCleanEventTypes()
		statusChanged = true
	}
	return statusChanged, nil
}

// handleSubscriptionDeletion deletes the NATS subscription and removes its finalizer if it is set.
func (r *Reconciler) handleSubscriptionDeletion(ctx context.Context, subscription *eventingv1alpha1.Subscription, log *zap.SugaredLogger) error {
	if utils.ContainsString(subscription.ObjectMeta.Finalizers, Finalizer) {
		if err := r.Backend.DeleteSubscription(subscription); err != nil {
			log.Errorw("Failed to delete NATS subscription", "error", err)
			// if failed to delete the external dependency here, return with error
			// so that it can be retried
			return err
		}

		// remove our finalizer from the list and update it.
		subscription.ObjectMeta.Finalizers = utils.RemoveString(subscription.ObjectMeta.Finalizers, Finalizer)
		if err := r.Client.Update(ctx, subscription); err != nil {
			events.Warn(r.recorder, subscription, events.ReasonUpdateFailed, "Update Subscription failed %s", subscription.Name)
			log.Errorw("Failed to remove finalizer from subscription", "error", err)
			return err
		}
		log.Debug("Removed finalizer from subscription")
	}
	return nil
}

// addFinalizerToSubscription appends the eventing finalizer to the subscription.
func (r *Reconciler) addFinalizerToSubscription(subscription *eventingv1alpha1.Subscription, log *zap.SugaredLogger) error {
	subscription.ObjectMeta.Finalizers = append(subscription.ObjectMeta.Finalizers, Finalizer)
	if err := r.Update(context.Background(), subscription); err != nil {
		log.Errorw("Failed to add finalizer to subscription", "error", err)
		return err
	}
	log.Debug("Added finalizer to subscription")
	return nil
}

// syncSubscriptionStatus syncs Subscription status
func (r *Reconciler) syncSubscriptionStatus(ctx context.Context, sub *eventingv1alpha1.Subscription, isNatsSubReady bool, forceUpdateStatus bool, message string) error {
	desiredConditions := make([]eventingv1alpha1.Condition, 0)
	conditionContained := false
	conditionsUpdated := false
	conditionStatus := corev1.ConditionFalse
	conditionReason := eventingv1alpha1.ConditionReasonNATSSubscriptionNotActive
	if isNatsSubReady {
		conditionStatus = corev1.ConditionTrue
		conditionReason = eventingv1alpha1.ConditionReasonNATSSubscriptionActive
	}
	condition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscriptionActive,
		conditionReason, conditionStatus, message)
	for _, c := range sub.Status.Conditions {
		var chosenCondition eventingv1alpha1.Condition
		if c.Type == condition.Type {
			if !conditionContained {
				if equalsConditionsIgnoringTime(c, condition) {
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
	}
	if conditionsUpdated || forceUpdateStatus {
		err := r.Client.Status().Update(ctx, sub, &client.UpdateOptions{})
		if err != nil {
			events.Warn(r.recorder, sub, events.ReasonUpdateFailed, "Update Subscription status failed %s", sub.Name)
			return errors.Wrapf(err, "update subscription status failed")
		}
		events.Normal(r.recorder, sub, events.ReasonUpdate, "Update Subscription status succeeded %s", sub.Name)
	}
	return nil
}

func (r *Reconciler) syncInvalidSubscriptions(ctx context.Context) (ctrl.Result, error) {
	natsHandler, _ := r.Backend.(*core.Nats)
	invalidSubs := natsHandler.GetInvalidSubscriptions()
	for _, v := range *invalidSubs {
		r.namedLogger().Debugw("Found invalid subscription", "namespace", v.Namespace, "name", v.Name)
		sub := &eventingv1alpha1.Subscription{}
		if err := r.Client.Get(ctx, v, sub); err != nil {
			r.namedLogger().Errorw("Failed to get invalid subscription", "namespace", v.Namespace, "name", v.Name, "error", err)
			continue
		}
		// mark the subscription to be not ready, it will throw a new reconcile call
		if err := r.syncSubscriptionStatus(ctx, sub, false, false, "invalid subscription"); err != nil {
			r.namedLogger().Errorw("Failed to sync status for invalid subscription", "namespace", v.Namespace, "name", v.Name, "error", err)
			return checkIsConflict(err)
		}
	}
	return ctrl.Result{}, nil
}

func (r *Reconciler) enqueueReconciliationForSubscriptions(subs []eventingv1alpha1.Subscription) {
	r.namedLogger().Debug("Enqueuing reconciliation request for all subscriptions")
	for i := range subs {
		r.customEventsChannel <- event.GenericEvent{Object: &subs[i]}
	}
}

func (r *Reconciler) namedLogger() *zap.SugaredLogger {
	return r.logger.WithContext().Named(reconcilerName)
}

//
// utilities functions
//

// equalsConditionsIgnoringTime checks if two conditions are equal, ignoring lastTransitionTime
func equalsConditionsIgnoringTime(a, b eventingv1alpha1.Condition) bool {
	return a.Type == b.Type && a.Status == b.Status && a.Reason == b.Reason && a.Message == b.Message
}

func checkIsConflict(err error) (ctrl.Result, error) {
	if k8serrors.IsConflict(err) {
		// Requeue the Request to try to reconcile it again
		return ctrl.Result{Requeue: true}, nil
	}
	return ctrl.Result{}, err
}
