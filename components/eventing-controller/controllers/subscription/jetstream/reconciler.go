package jetstream

import (
	"context"
	"os"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"

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
	ctx      context.Context
	Backend  handlers.MessagingBackend
	recorder record.EventRecorder
	logger   *logger.Logger
}

func NewReconciler(ctx context.Context, client client.Client, logger *logger.Logger, recorder record.EventRecorder, cfg env.NatsConfig) *Reconciler {
	reconciler := &Reconciler{
		Client:   client,
		ctx:      ctx,
		recorder: recorder,
		logger:   logger,
	}
	jsHandler := handlers.NewJetStream(cfg, logger)
	if err := jsHandler.Initialize(env.Config{}); err != nil {
		logger.WithContext().Errorw("start reconciler failed", "name", reconcilerName, "error", err)
		panic(err)
	}
	reconciler.Backend = jsHandler

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

func (r *Reconciler) Reconcile(_ context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.namedLogger().Errorf("cannot reconcile subscription %s (not implemented)", req)
	return ctrl.Result{}, nil
}

func (r *Reconciler) namedLogger() *zap.SugaredLogger {
	return r.logger.WithContext().Named(reconcilerName)
}
