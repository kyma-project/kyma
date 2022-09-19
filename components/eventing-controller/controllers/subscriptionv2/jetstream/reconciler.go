package jetstream

import (
	"context"

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
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/eventtype"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/natsv2/jetstream"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/sink"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
)

const (
	reconcilerName = "jetstream-subscription-reconciler"
)

type Reconciler struct {
	client.Client
	ctx                 context.Context
	Backend             jetstream.Backend
	recorder            record.EventRecorder
	logger              *logger.Logger
	eventTypeCleaner    eventtype.Cleaner
	subsConfig          env.DefaultSubscriptionConfig
	sinkValidator       sink.Validator
	customEventsChannel chan event.GenericEvent
}

func NewReconciler(ctx context.Context, client client.Client, jsHandler jetstream.Backend, logger *logger.Logger,
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
	//TODO: add it back
	//if err := jsHandler.Initialize(reconciler.handleNatsConnClose); err != nil {
	//	logger.WithContext().Errorw("Failed to start reconciler", "name", reconcilerName, "error", err)
	//	panic(err)
	//}
	return reconciler
}

// SetupUnmanaged creates a controller under the client control.
func (r *Reconciler) SetupUnmanaged(mgr ctrl.Manager) error {
	ctru, err := controller.NewUnmanaged(reconcilerName, mgr, controller.Options{Reconciler: r})
	if err != nil {
		r.namedLogger().Errorw("Failed to create unmanaged controller", "error", err)
		return err
	}

	if err := ctru.Watch(&source.Kind{Type: &eventingv1alpha2.Subscription{}}, &handler.EnqueueRequestForObject{}); err != nil {
		r.namedLogger().Errorw("Failed to setup watch for subscriptions", "error", err)
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
// Generated required RBAC to list Applications (required by event type cleaner).
// +kubebuilder:rbac:groups="applicationconnector.kyma-project.io",resources=applications,verbs=get;list;watch

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.namedLogger().Debugw("Received subscription reconciliation request", "namespace", req.Namespace, "name", req.Name)

	return ctrl.Result{}, nil
}

func (r *Reconciler) namedLogger() *zap.SugaredLogger {
	return r.logger.WithContext().Named(reconcilerName)
}
