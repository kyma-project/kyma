package jetstream

import (
	"context"
	"os"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/eventtype"

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

type Reconciler struct {
	client.Client
	ctx              context.Context
	Backend          handlers.JetStreamBackend
	recorder         record.EventRecorder
	logger           *logger.Logger
	eventTypeCleaner eventtype.Cleaner
}

func NewReconciler(ctx context.Context, client client.Client, jsHandler handlers.JetStreamBackend, logger *logger.Logger, recorder record.EventRecorder, cleaner eventtype.Cleaner) *Reconciler {
	reconciler := &Reconciler{
		Client:           client,
		ctx:              ctx,
		Backend:          jsHandler,
		recorder:         recorder,
		logger:           logger,
		eventTypeCleaner: cleaner,
	}
	if err := jsHandler.Initialize(reconciler.handleNatsConnClose); err != nil {
		logger.WithContext().Errorw("start reconciler failed", "name", reconcilerName, "error", err)
		panic(err)
	}
	return reconciler
}

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

	desiredSubscription := actualSubscription.DeepCopy()
	log := utils.LoggerWithSubscription(r.namedLogger(), desiredSubscription)

	// TODO: Do this as part of sync initial status
	cleanedSubjects, _ := handlers.GetCleanSubjects(desiredSubscription, r.eventTypeCleaner)
	desiredSubscription.Status.CleanEventTypes = r.Backend.GetJetStreamSubjects(cleanedSubjects)
	desiredSubscription.Status.Config = &eventingv1alpha1.SubscriptionConfig{
		MaxInFlightMessages: 10,
	}
	_ = r.Backend.SyncSubscription(desiredSubscription)

	// Mark subscription as not ready, since we have not implemented the reconciliation logic.
	desiredSubscription.Status = eventingv1alpha1.SubscriptionStatus{}
	if err := r.syncSubscriptionStatus(ctx, desiredSubscription); err != nil {
		return ctrl.Result{}, err
	}

	log.Error("cannot reconcile JetStream subscription (not implemented)")
	return ctrl.Result{}, nil
}

// handleNatsConnClose is called by NATS when the connection to the NATS server is closed. When it
// is called, the reconnect-attempts have exceeded the defined value.
// It forces reconciling the subscription to make sure the subscription is marked as not ready, until
// it is possible to connect to the NATS server again.
func (r *Reconciler) handleNatsConnClose(_ *nats.Conn) {
	// TODO: implement me!
	// TODO: Enable TestSubscription_JetStreamServerRestart once implemented
}

func (r *Reconciler) syncSubscriptionStatus(ctx context.Context, sub *eventingv1alpha1.Subscription) error {
	if err := r.Client.Status().Update(ctx, sub); err != nil {
		events.Warn(r.recorder, sub, events.ReasonUpdateFailed, "Update Subscription status failed %s", sub.Name)
		return errors.Wrapf(err, "update subscription status failed")
	}
	events.Normal(r.recorder, sub, events.ReasonUpdate, "Update Subscription status succeeded %s", sub.Name)
	return nil
}

func (r *Reconciler) namedLogger() *zap.SugaredLogger {
	return r.logger.WithContext().Named(reconcilerName)
}
