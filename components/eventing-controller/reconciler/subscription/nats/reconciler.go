package nats

import (
	"context"
	"net/url"
	"os"
	"reflect"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/eventtype"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"
)

// Reconciler reconciles a Subscription object
type Reconciler struct {
	ctx context.Context
	client.Client
	cache.Cache
	Backend          handlers.MessagingBackend
	logger           *logger.Logger
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

	reconcilerName = "nats-subscription-reconciler"
)

func NewReconciler(ctx context.Context, client client.Client, applicationLister *application.Lister, cache cache.Cache, logger *logger.Logger, recorder record.EventRecorder, cfg env.NatsConfig, subsCfg env.DefaultSubscriptionConfig) *Reconciler {
	natsHandler := handlers.NewNats(cfg, subsCfg, logger)
	if err := natsHandler.Initialize(env.Config{}); err != nil {
		logger.WithContext().Errorw("start reconciler failed", "name", reconcilerName, "error", err)
		panic(err)
	}

	return &Reconciler{
		ctx:              ctx,
		Client:           client,
		Cache:            cache,
		Backend:          natsHandler,
		logger:           logger,
		recorder:         recorder,
		eventTypeCleaner: eventtype.NewCleaner(cfg.EventTypePrefix, applicationLister, logger),
	}
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&eventingv1alpha1.Subscription{}).
		Complete(r)
}

// SetupUnmanaged creates a controller under the client control
func (r *Reconciler) SetupUnmanaged(mgr ctrl.Manager) error {
	ctru, err := controller.NewUnmanaged(reconcilerName, mgr, controller.Options{Reconciler: r})
	if err != nil {
		r.namedLogger().Errorw("create unmanaged controller failed", "name", reconcilerName, "error", err)
		return err
	}

	if err := ctru.Watch(&source.Kind{Type: &eventingv1alpha1.Subscription{}}, &handler.EnqueueRequestForObject{}); err != nil {
		r.namedLogger().Errorw("watch subscriptions failed", "error", err)
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
		r.namedLogger().Errorw("watch nats server failed", "pod", NATSFirstInstanceName, "error", err)
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
	if req.Name == NATSFirstInstanceName && req.Namespace == NATSNamespace {
		r.namedLogger().Debugw("received watch request", "namespace", req.Namespace, "name", req.Name)
		return r.syncInvalidSubscriptions(ctx)
	}

	r.namedLogger().Debugw("received subscription reconciliation request", "namespace", req.Namespace, "name", req.Name)

	actualSubscription := &eventingv1alpha1.Subscription{}
	result := ctrl.Result{}

	// Ensure the object was not deleted in the meantime
	err := r.Client.Get(ctx, req.NamespacedName, actualSubscription)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Handle only the new subscription
	desiredSubscription := actualSubscription.DeepCopy()

	// Bind fields to logger
	log := utils.LoggerWithSubscription(r.namedLogger(), desiredSubscription)

	if !desiredSubscription.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is being deleted
		if utils.ContainsString(desiredSubscription.ObjectMeta.Finalizers, Finalizer) {
			if err := r.Backend.DeleteSubscription(desiredSubscription); err != nil {
				log.Errorw("delete subscription failed", "error", err)
				// if failed to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			desiredSubscription.ObjectMeta.Finalizers = utils.RemoveString(desiredSubscription.ObjectMeta.Finalizers, Finalizer)
			if err := r.Client.Update(ctx, desiredSubscription); err != nil {
				log.Errorw("remove finalizer from subscription failed", "error", err)
				return ctrl.Result{}, err
			}
			log.Debug("remove finalizer from subscription succeeded")
			return ctrl.Result{}, nil
		}
	}
	// Check for valid sink
	if err := r.assertSinkValidity(actualSubscription.Spec.Sink); err != nil {
		log.Errorw("parse sink URL failed", "error", err)
		if err := r.syncSubscriptionStatus(ctx, actualSubscription, false, err.Error()); err != nil {
			return ctrl.Result{}, err
		}
		// No point in reconciling as the sink is invalid
		return ctrl.Result{}, nil
	}

	// Clean up the old subscriptions
	if err := r.Backend.DeleteSubscription(desiredSubscription); err != nil {
		log.Errorw("delete subscription failed", "error", err)
		if err := r.syncSubscriptionStatus(ctx, actualSubscription, false, err.Error()); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}

	// The object is not being deleted, so if it does not have our finalizer,
	// then lets add the finalizer and update the object. This is equivalent
	// registering our finalizer.
	if !utils.ContainsString(desiredSubscription.ObjectMeta.Finalizers, Finalizer) {
		desiredSubscription.ObjectMeta.Finalizers = append(desiredSubscription.ObjectMeta.Finalizers, Finalizer)
		if err := r.Update(context.Background(), desiredSubscription); err != nil {
			log.Errorw("add finalizer to subscription failed", "error", err)
			return ctrl.Result{}, err
		}
		log.Debug("add finalizer to subscription succeeded")
		result.Requeue = true
	}

	if result.Requeue {
		return result, nil
	}

	_, err = r.Backend.SyncSubscription(desiredSubscription, r.eventTypeCleaner)
	if err != nil {
		log.Errorw("sync subscription failed", "error", err)
		if err := r.syncSubscriptionStatus(ctx, actualSubscription, false, err.Error()); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}
	log.Debug("create NATS subscriptions succeeded")

	actualSubscription.Status.Config = desiredSubscription.Status.Config
	// Update status
	if err := r.syncSubscriptionStatus(ctx, actualSubscription, true, ""); err != nil {
		return ctrl.Result{}, err
	}

	return result, nil
}

// syncSubscriptionStatus syncs Subscription status
// subsConfig is the subscription configuration that was applied to the subscription. It is set only if the
// isNatsSubReady is true.
func (r *Reconciler) syncSubscriptionStatus(ctx context.Context, sub *eventingv1alpha1.Subscription, isNatsSubReady bool, message string) error {
	desiredSubscription := sub.DeepCopy()
	desiredConditions := make([]eventingv1alpha1.Condition, 0)
	conditionAdded := false
	condition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscriptionActive,
		eventingv1alpha1.ConditionReasonNATSSubscriptionActive, corev1.ConditionFalse, message)
	if isNatsSubReady {
		condition = eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscriptionActive,
			eventingv1alpha1.ConditionReasonNATSSubscriptionActive, corev1.ConditionTrue, "")
	}
	for _, c := range sub.Status.Conditions {
		var chosenCondition eventingv1alpha1.Condition
		if c.Type == condition.Type {
			// take given condition
			chosenCondition = condition
			conditionAdded = true
		} else {
			// take already present condition
			chosenCondition = c
		}
		desiredConditions = append(desiredConditions, chosenCondition)
	}
	if !conditionAdded {
		desiredConditions = append(desiredConditions, condition)
	}
	desiredSubscription.Status.Conditions = desiredConditions
	desiredSubscription.Status.Ready = isNatsSubReady

	if !reflect.DeepEqual(sub.Status, desiredSubscription.Status) {
		err := r.Client.Status().Update(ctx, desiredSubscription, &client.UpdateOptions{})
		if err != nil {
			return errors.Wrapf(err, "update subscription status failed")
		}
	}
	return nil
}

func (r *Reconciler) assertSinkValidity(sink string) error {
	_, err := url.ParseRequestURI(sink)
	return err
}

func (r *Reconciler) syncInvalidSubscriptions(ctx context.Context) (ctrl.Result, error) {
	natsHandler, _ := r.Backend.(*handlers.Nats)
	namespacedName := natsHandler.GetInvalidSubscriptions()
	for _, v := range *namespacedName {
		r.namedLogger().Debugw("found invalid subscription", "namespace", v.Namespace, "name", v.Name)
		sub := &eventingv1alpha1.Subscription{}
		if err := r.Client.Get(ctx, v, sub); err != nil {
			r.namedLogger().Errorw("get invalid subscription failed", "namespace", v.Namespace, "name", v.Name, "error", err)
			continue
		}
		// mark the subscription to be not ready, it will throw a new reconcile call
		if err := r.syncSubscriptionStatus(ctx, sub, false, "invalid subscription"); err != nil {
			r.namedLogger().Errorw("sync status for invalid subscription failed", "namespace", v.Namespace, "name", v.Name, "error", err)
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func (r *Reconciler) namedLogger() *zap.SugaredLogger {
	return r.logger.WithContext().Named(reconcilerName)
}
