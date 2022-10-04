package beb

import (
	"context"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"

	apigatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/beb"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/eventmesh"
	"golang.org/x/xerrors"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"go.uber.org/zap"

	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/sink"
	backendutils "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/utils"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
)

// Reconciler reconciles a Subscription object.
type Reconciler struct {
	ctx context.Context
	client.Client
	logger            *logger.Logger
	recorder          record.EventRecorder
	Backend           eventmesh.Backend
	Domain            string
	cleaner           cleaner.Cleaner
	oauth2credentials *beb.OAuth2ClientCredentials
	// nameMapper is used to map the Kyma subscription name to a subscription name on BEB
	nameMapper    backendutils.NameMapper
	sinkValidator sink.Validator
}

const (
	reconcilerName = "beb-subscription-reconciler"
)

func NewReconciler(ctx context.Context, client client.Client, logger *logger.Logger, recorder record.EventRecorder,
	cfg env.Config, cleaner cleaner.Cleaner, bebBackend eventmesh.Backend, credential *beb.OAuth2ClientCredentials,
	mapper backendutils.NameMapper, validator sink.Validator) *Reconciler {
	if err := bebBackend.Initialize(cfg); err != nil {
		logger.WithContext().Errorw("Failed to start reconciler", "name", reconcilerName, "error", err)
		panic(err)
	}
	return &Reconciler{
		ctx:               ctx,
		Client:            client,
		logger:            logger,
		recorder:          recorder,
		Backend:           bebBackend,
		Domain:            cfg.Domain,
		cleaner:           cleaner,
		oauth2credentials: credential,
		nameMapper:        mapper,
		sinkValidator:     validator,
	}
}

// SetupUnmanaged creates a controller under the client control.
func (r *Reconciler) SetupUnmanaged(mgr ctrl.Manager) error {
	ctru, err := controller.NewUnmanaged(reconcilerName, mgr, controller.Options{Reconciler: r})
	if err != nil {
		return xerrors.Errorf("failed to create unmanaged controller: %v", err)
	}

	if err := ctru.Watch(&source.Kind{Type: &eventingv1alpha2.Subscription{}}, &handler.EnqueueRequestForObject{}); err != nil {
		return xerrors.Errorf("failed to watch subscriptions: %v", err)
	}

	apiRuleEventHandler := &handler.EnqueueRequestForOwner{OwnerType: &eventingv1alpha2.Subscription{}, IsController: false}
	if err := ctru.Watch(&source.Kind{Type: &apigatewayv1beta1.APIRule{}}, apiRuleEventHandler); err != nil {
		return xerrors.Errorf("failed to watch APIRule: %v", err)
	}

	go func(r *Reconciler, c controller.Controller) {
		if err := c.Start(r.ctx); err != nil {
			r.namedLogger().Fatalw("Failed to start controller", "name", reconcilerName, "error", err)
		}
	}(r, ctru)

	return nil
}

// +kubebuilder:rbac:groups=eventing.kyma-project.io,resources=subscriptions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=eventing.kyma-project.io,resources=subscriptions/status,verbs=get;update;patch
// Generate required RBAC to emit kubernetes events in the controller.
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
// +kubebuilder:rbac:groups=gateway.kyma-project.io,resources=apirules,verbs=get;list;watch;create;update;patch;delete
// Generated required RBAC to list Applications (required by event type cleaner).
// +kubebuilder:rbac:groups="applicationconnector.kyma-project.io",resources=applications,verbs=get;list;watch

func (r *Reconciler) Reconcile(_ context.Context, _ ctrl.Request) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (r *Reconciler) namedLogger() *zap.SugaredLogger {
	return r.logger.WithContext().Named(reconcilerName)
}
