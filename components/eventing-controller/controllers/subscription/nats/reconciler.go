package nats

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strings"

	"github.com/nats-io/nats.go"

	"github.com/kyma-project/kyma/components/eventing-controller/controllers/events"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	k8stypes "k8s.io/apimachinery/pkg/types"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/eventtype"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"
)

type sinkValidator func(ctx context.Context, r *Reconciler, subscription *eventingv1alpha1.Subscription) error

// Reconciler reconciles a Subscription object
type Reconciler struct {
	client.Client
	sinkValidator

	ctx              context.Context
	Backend          handlers.NatsBackend
	logger           *logger.Logger
	recorder         record.EventRecorder
	subsConfig       env.DefaultSubscriptionConfig
	eventTypeCleaner eventtype.Cleaner
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
	clusterLocalURLSuffix = "svc.cluster.local"
)

func NewReconciler(ctx context.Context, client client.Client, natsHandler handlers.NatsBackend, cleaner eventtype.Cleaner,
	logger *logger.Logger, recorder record.EventRecorder, subsCfg env.DefaultSubscriptionConfig) *Reconciler {
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
		logger.WithContext().Errorw("start reconciler failed", "name", reconcilerName, "error", err)
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
		r.namedLogger().Errorw("create unmanaged controller failed", "name", reconcilerName, "error", err)
		return err
	}

	if err := ctru.Watch(&source.Kind{Type: &eventingv1alpha1.Subscription{}}, &handler.EnqueueRequestForObject{}); err != nil {
		r.namedLogger().Errorw("setup watch for subscriptions failed", "error", err)
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
		r.namedLogger().Errorw("setup watch for nats server failed", "pod", natsFirstInstanceName, "error", err)
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
	if req.Name == natsFirstInstanceName && req.Namespace == natsNamespace {
		r.namedLogger().Debugw("received watch request", "namespace", req.Namespace, "name", req.Name)
		return r.syncInvalidSubscriptions(ctx)
	}

	r.namedLogger().Debugw("received subscription reconciliation request", "namespace", req.Namespace, "name", req.Name)

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
		log.Errorw("sync initial status failed", "error", err)
		if syncErr := r.syncSubscriptionStatus(ctx, subscription, false, statusChanged, err.Error()); err != nil {
			return ctrl.Result{}, syncErr
		}
		return ctrl.Result{}, err
	}

	// Check for valid sink
	if err := r.sinkValidator(ctx, r, subscription); err != nil {
		log.Errorw("sink URL validation failed", "error", err)
		if syncErr := r.syncSubscriptionStatus(ctx, subscription, false, statusChanged, err.Error()); err != nil {
			return ctrl.Result{}, syncErr
		}
		// No point in reconciling as the sink is invalid, return latest error to requeue the reconciliation request
		return ctrl.Result{}, err
	}

	// Synchronize Kyma subscription to NATS backend
	syncErr := r.Backend.SyncSubscription(subscription)
	if syncErr != nil {
		log.Errorw("sync subscription failed", "error", syncErr)
		if err := r.syncSubscriptionStatus(ctx, subscription, false, statusChanged, syncErr.Error()); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, syncErr
	}
	log.Debug("create NATS subscriptions succeeded")

	// Update status
	if err := r.syncSubscriptionStatus(ctx, subscription, true, statusChanged, ""); err != nil {
		return checkIsConflict(err)
	}

	return ctrl.Result{}, nil
}

// syncInitialStatus keeps the latest cleanEventTypes and Config in the subscription
func (r *Reconciler) syncInitialStatus(subscription *eventingv1alpha1.Subscription, log *zap.SugaredLogger) (bool, error) {
	statusChanged := false
	cleanEventTypes, err := handlers.GetCleanSubjects(subscription, r.eventTypeCleaner)
	if err != nil {
		log.Errorw("get clean subject failed", "error", err)
		return false, err
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
	return statusChanged, nil
}

// handleSubscriptionDeletion deletes the NATS subscription and removes its finalizer if it is set.
func (r *Reconciler) handleSubscriptionDeletion(ctx context.Context, subscription *eventingv1alpha1.Subscription, log *zap.SugaredLogger) error {
	if utils.ContainsString(subscription.ObjectMeta.Finalizers, Finalizer) {
		if err := r.Backend.DeleteSubscription(subscription); err != nil {
			log.Errorw("delete NATS subscription failed", "error", err)
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
	if err := r.Update(context.Background(), subscription); err != nil {
		log.Errorw("add finalizer to subscription failed", "error", err)
		return err
	}
	log.Debug("add finalizer to subscription succeeded")
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
	natsHandler, _ := r.Backend.(*handlers.Nats)
	invalidSubs := natsHandler.GetInvalidSubscriptions()
	for _, v := range *invalidSubs {
		r.namedLogger().Debugw("found invalid subscription", "namespace", v.Namespace, "name", v.Name)
		sub := &eventingv1alpha1.Subscription{}
		if err := r.Client.Get(ctx, v, sub); err != nil {
			r.namedLogger().Errorw("get invalid subscription failed", "namespace", v.Namespace, "name", v.Name, "error", err)
			continue
		}
		// mark the subscription to be not ready, it will throw a new reconcile call
		if err := r.syncSubscriptionStatus(ctx, sub, false, false, "invalid subscription"); err != nil {
			r.namedLogger().Errorw("sync status for invalid subscription failed", "namespace", v.Namespace, "name", v.Name, "error", err)
			return checkIsConflict(err)
		}
	}
	return ctrl.Result{}, nil
}

func (r *Reconciler) getClusterLocalService(ctx context.Context, svcNs, svcName string) (*corev1.Service, error) {
	svcLookupKey := k8stypes.NamespacedName{Name: svcName, Namespace: svcNs}
	svc := &corev1.Service{}
	if err := r.Client.Get(ctx, svcLookupKey, svc); err != nil {
		return nil, err
	}
	return svc, nil
}

func (r *Reconciler) enqueueReconciliationForSubscriptions(subs []eventingv1alpha1.Subscription) {
	r.namedLogger().Debug("enqueuing reconciliation request for all subscriptions")
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

// defaultSinkValidator validates the "sink" defined in Kyma subscriptions
func defaultSinkValidator(ctx context.Context, r *Reconciler, subscription *eventingv1alpha1.Subscription) error {
	if !isValidScheme(subscription.Spec.Sink) {
		events.Warn(r.recorder, subscription, events.ReasonValidationFailed, "Sink URL scheme should be HTTP or HTTPS: %s", subscription.Spec.Sink)
		return fmt.Errorf("sink URL scheme should be 'http' or 'https'")
	}

	sURL, err := url.ParseRequestURI(subscription.Spec.Sink)
	if err != nil {
		events.Warn(r.recorder, subscription, events.ReasonValidationFailed, "Not able to parse Sink URL with error: %s", err.Error())
		return fmt.Errorf("not able to parse sink url with error: %s", err.Error())
	}

	// Validate sink URL is a cluster local URL
	trimmedHost := strings.Split(sURL.Host, ":")[0]
	if !strings.HasSuffix(trimmedHost, clusterLocalURLSuffix) {
		events.Warn(r.recorder, subscription, events.ReasonValidationFailed, "Sink does not contain suffix: %s", clusterLocalURLSuffix)
		return fmt.Errorf("sink does not contain suffix: %s in the URL", clusterLocalURLSuffix)
	}

	// we expected a sink in the format "service.namespace.svc.cluster.local"
	subDomains := strings.Split(trimmedHost, ".")
	if len(subDomains) != 5 {
		events.Warn(r.recorder, subscription, events.ReasonValidationFailed, "Sink should contain 5 sub-domains: %s", trimmedHost)
		return fmt.Errorf("sink should contain 5 sub-domains: %s", trimmedHost)
	}

	// Assumption: Subscription CR and Subscriber should be deployed in the same namespace
	svcNs := subDomains[1]
	if subscription.Namespace != svcNs {
		events.Warn(r.recorder, subscription, events.ReasonValidationFailed, "natsNamespace of subscription: %s and the subscriber: %s are different", subscription.Namespace, svcNs)
		return fmt.Errorf("namespace of subscription: %s and the namespace of subscriber: %s are different", subscription.Namespace, svcNs)
	}

	// Validate svc is a cluster-local one
	svcName := subDomains[0]
	if _, err := r.getClusterLocalService(ctx, svcNs, svcName); err != nil {
		if k8serrors.IsNotFound(err) {
			events.Warn(r.recorder, subscription, events.ReasonValidationFailed, "Sink does not correspond to a valid cluster local svc")
			return fmt.Errorf("sink is not valid cluster local svc, failed with error: %w", err)
		}

		events.Warn(r.recorder, subscription, events.ReasonValidationFailed, "Fetch cluster-local svc failed namespace %s name %s", svcNs, svcName)
		return fmt.Errorf("fetch cluster-local svc failed namespace:%s name:%s with error: %w", svcNs, svcName, err)
	}

	return nil
}

// isValidScheme returns true if the sink scheme is http or https, otherwise returns false.
func isValidScheme(sink string) bool {
	return strings.HasPrefix(sink, "http://") || strings.HasPrefix(sink, "https://")
}

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
