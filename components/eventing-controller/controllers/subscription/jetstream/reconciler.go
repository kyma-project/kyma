package jetstream

import (
	"context"
	"reflect"
	"time"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/metrics"

	k8stypes "k8s.io/apimachinery/pkg/types"

	"github.com/kyma-project/kyma/components/eventing-controller/controllers/events"

	"github.com/nats-io/nats.go"

	pkgerrors "github.com/kyma-project/kyma/components/eventing-controller/pkg/errors"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/object"

	"github.com/pkg/errors"

	"go.uber.org/zap"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/sink"
	backendutils "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/utils"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"

	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	jetstream "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/jetstream"
)

const (
	reconcilerName  = "jetstream-subscription-reconciler"
	requeueDuration = 10 * time.Second
	backendType     = "NATS_Jetstream"
)

type Reconciler struct {
	client.Client
	ctx                 context.Context
	Backend             jetstream.Backend
	recorder            record.EventRecorder
	logger              *logger.Logger
	cleaner             cleaner.Cleaner
	sinkValidator       sink.Validator
	customEventsChannel chan event.GenericEvent
	collector           *metrics.Collector
}

func NewReconciler(ctx context.Context, client client.Client, jsBackend jetstream.Backend,
	logger *logger.Logger, recorder record.EventRecorder, cleaner cleaner.Cleaner,
	defaultSinkValidator sink.Validator, collector *metrics.Collector) *Reconciler {
	reconciler := &Reconciler{
		Client:              client,
		ctx:                 ctx,
		Backend:             jsBackend,
		recorder:            recorder,
		logger:              logger,
		cleaner:             cleaner,
		sinkValidator:       defaultSinkValidator,
		customEventsChannel: make(chan event.GenericEvent),
		collector:           collector,
	}
	if err := jsBackend.Initialize(reconciler.handleNatsConnClose); err != nil {
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

//nolint:lll
// +kubebuilder:rbac:groups=eventing.kyma-project.io,resources=subscriptions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=eventing.kyma-project.io,resources=subscriptions/status,verbs=get;update;patch
// Generate required RBAC to emit kubernetes events in the controller.
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.namedLogger().Debugw("Received subscription v1alpha2 reconciliation request",
		"namespace", req.Namespace, "name", req.Name)

	// fetch current subscription object and ensure the object was not deleted in the meantime
	currentSubscription := &eventingv1alpha2.Subscription{}
	err := r.Client.Get(ctx, req.NamespacedName, currentSubscription)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// copy the subscription object, so we don't modify the source object
	desiredSubscription := currentSubscription.DeepCopy()

	// Bind fields to logger
	log := backendutils.LoggerWithSubscription(r.namedLogger(), desiredSubscription)

	if isInDeletion(desiredSubscription) {
		// The object is being deleted
		return r.handleSubscriptionDeletion(ctx, desiredSubscription, log)
	}

	defer func() {
		// Update metrics
		for _, cc := range currentSubscription.Status.Backend.Types {
			found := false
			for _, dc := range desiredSubscription.Status.Backend.Types {
				if cc.ConsumerName == dc.ConsumerName {
					found = true
				}
			}
			if !found {
				r.collector.RemoveSubscriptionStatus(
					currentSubscription.Name,
					currentSubscription.Namespace,
					backendType,
					cc.ConsumerName,
					r.Backend.GetConfig().JSStreamName)
			}
		}
		for _, dc := range desiredSubscription.Status.Backend.Types {
			r.collector.RecordSubscriptionStatus(desiredSubscription.Status.Ready,
				desiredSubscription.Name,
				desiredSubscription.Namespace,
				backendType,
				dc.ConsumerName,
				r.Backend.GetConfig().JSStreamName,
			)
		}

	}()

	// The object is not being deleted, so if it does not have our finalizer,
	// then lets add the finalizer and update the object.
	if !containsFinalizer(desiredSubscription) {
		return r.addFinalizer(ctx, desiredSubscription)
	}

	// update the cleanEventTypes and config values in the subscription status, if changed
	if err = r.syncEventTypes(desiredSubscription); err != nil {
		if syncErr := r.syncSubscriptionStatus(ctx, desiredSubscription, err, log); syncErr != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}

	// Check for valid sink
	if err := r.sinkValidator.Validate(desiredSubscription); err != nil {
		// No point in reconciling as the sink is invalid,
		// return latest error to requeue the reconciliation request
		if syncErr := r.syncSubscriptionStatus(ctx, desiredSubscription, err, log); syncErr != nil {
			return ctrl.Result{}, err
		}
		// No point in reconciling as the sink is invalid, return latest error to requeue the reconciliation request
		return ctrl.Result{}, err
	}

	// Synchronize Kyma subscription to JetStream backend
	if syncSubErr := r.Backend.SyncSubscription(desiredSubscription); syncSubErr != nil {
		result := ctrl.Result{}
		if syncErr := r.syncSubscriptionStatus(ctx, desiredSubscription, syncSubErr, log); syncErr != nil {
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
	return ctrl.Result{}, r.syncSubscriptionStatus(ctx, desiredSubscription, nil, log)
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

// handleSubscriptionDeletion deletes the JetStream subscription and removes its finalizer if it is set.
func (r *Reconciler) handleSubscriptionDeletion(ctx context.Context,
	subscription *eventingv1alpha2.Subscription, log *zap.SugaredLogger) (ctrl.Result, error) {
	// delete the JetStream subscription/consumer
	if utils.ContainsString(subscription.ObjectMeta.Finalizers, eventingv1alpha2.Finalizer) {
		if err := r.Backend.DeleteSubscription(subscription); err != nil {
			deleteSubErr := pkgerrors.MakeError(errFailedToDeleteSub, err)
			// if failed to delete the external dependency here, return with error
			// so that it can be retried
			if syncErr := r.syncSubscriptionStatus(ctx, subscription, deleteSubErr, log); syncErr != nil {
				return ctrl.Result{}, syncErr
			}
			return ctrl.Result{}, deleteSubErr
		}

		types := subscription.Status.Backend.Types
		// remove the eventing finalizer from the list and update the subscription.
		subscription.ObjectMeta.Finalizers = utils.RemoveString(subscription.ObjectMeta.Finalizers,
			eventingv1alpha2.Finalizer)

		// update the subscription's finalizers in k8s
		if err := r.Update(ctx, subscription); err != nil {
			return ctrl.Result{}, pkgerrors.MakeError(errFailedToUpdateFinalizers, err)
		}
		for _, t := range types {
			r.collector.RemoveSubscriptionStatus(subscription.Name, subscription.Namespace, backendType, t.ConsumerName, r.Backend.GetConfig().JSStreamName)
		}

		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, nil
}

// syncSubscriptionStatus syncs Subscription status and updates the k8s subscription.
func (r *Reconciler) syncSubscriptionStatus(ctx context.Context,
	desiredSubscription *eventingv1alpha2.Subscription, err error, log *zap.SugaredLogger) error {
	// set ready state
	desiredSubscription.Status.Ready = err == nil

	// compile the desired conditions
	desiredSubscription.Status.Conditions = eventingv1alpha2.GetSubscriptionActiveCondition(desiredSubscription, err)

	// Update the subscription
	return r.updateSubscriptionStatus(ctx, desiredSubscription, log)
}

// updateSubscriptionStatus updates the subscription's status changes to k8s.
func (r *Reconciler) updateSubscriptionStatus(ctx context.Context,
	sub *eventingv1alpha2.Subscription, logger *zap.SugaredLogger) error {
	namespacedName := &k8stypes.NamespacedName{
		Name:      sub.Name,
		Namespace: sub.Namespace,
	}

	// fetch the latest subscription object, to avoid k8s conflict errors
	actualSubscription := &eventingv1alpha2.Subscription{}
	if err := r.Client.Get(ctx, *namespacedName, actualSubscription); err != nil {
		return err
	}

	// copy new changes to the latest object
	desiredSubscription := actualSubscription.DeepCopy()
	desiredSubscription.Status = sub.Status

	// sync subscription status with k8s
	if err := r.updateStatus(ctx, actualSubscription, desiredSubscription, logger); err != nil {
		return pkgerrors.MakeError(errFailedToUpdateStatus, err)
	}

	return nil
}

// updateStatus updates the status to k8s if modified.
func (r *Reconciler) updateStatus(ctx context.Context, oldSubscription,
	newSubscription *eventingv1alpha2.Subscription, logger *zap.SugaredLogger) error {
	// compare the status taking into consideration lastTransitionTime in conditions
	if object.IsSubscriptionStatusEqual(oldSubscription.Status, newSubscription.Status) {
		return nil
	}

	// update the status for subscription in k8s
	if err := r.Status().Update(ctx, newSubscription); err != nil {
		events.Warn(r.recorder, newSubscription, events.ReasonUpdateFailed,
			"Update Subscription status failed %s", newSubscription.Name)
		return pkgerrors.MakeError(errFailedToUpdateStatus, err)
	}
	events.Normal(r.recorder, newSubscription, events.ReasonUpdate,
		"Update Subscription status succeeded %s", newSubscription.Name)

	logger.Debugw("Updated subscription status",
		"oldStatus", oldSubscription.Status, "newStatus", newSubscription.Status)

	return nil
}

// addFinalizer appends the eventing finalizer to the subscription and updates it in k8s.
func (r *Reconciler) addFinalizer(ctx context.Context, sub *eventingv1alpha2.Subscription) (ctrl.Result, error) {
	sub.ObjectMeta.Finalizers = append(sub.ObjectMeta.Finalizers, eventingv1alpha2.Finalizer)

	// update the subscription's finalizers in k8s
	if err := r.Update(ctx, sub); err != nil {
		return ctrl.Result{}, pkgerrors.MakeError(errFailedToUpdateFinalizers, err)
	}

	return ctrl.Result{}, nil
}

// syncEventTypes sets the latest cleaned types and jetStreamTypes to the subscription status.
func (r *Reconciler) syncEventTypes(desiredSubscription *eventingv1alpha2.Subscription) error {
	// clean types
	cleanedTypes := jetstream.GetCleanEventTypes(desiredSubscription, r.cleaner)
	if !reflect.DeepEqual(desiredSubscription.Status.Types, cleanedTypes) {
		desiredSubscription.Status.Types = cleanedTypes
	}

	// jetStreamTypes
	jsSubjects := r.Backend.GetJetStreamSubjects(desiredSubscription.Spec.Source,
		jetstream.GetCleanEventTypesFromEventTypes(cleanedTypes),
		desiredSubscription.Spec.TypeMatching)
	jsTypes, err := jetstream.GetBackendJetStreamTypes(desiredSubscription, jsSubjects)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(desiredSubscription.Status.Backend.Types, jsTypes) {
		desiredSubscription.Status.Backend.Types = jsTypes
	}
	return nil
}

func (r *Reconciler) namedLogger() *zap.SugaredLogger {
	return r.logger.WithContext().Named(reconcilerName)
}
