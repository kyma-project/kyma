package jetstream

import (
	"context"
	"fmt"
	"time"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/eventtype"
	"k8s.io/client-go/dynamic"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	"k8s.io/client-go/rest"

	"github.com/kyma-project/kyma/components/eventing-controller/controllers/subscription/jetstream"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/subscriptionmanager"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	subscriptionManagerName = "jetstream-subscription-manager"
)

// AddToScheme adds all types of clientset and eventing into the given scheme.
func AddToScheme(scheme *runtime.Scheme) error {
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return err
	}
	if err := eventingv1alpha1.AddToScheme(scheme); err != nil {
		return err
	}
	return nil
}

type SubscriptionManager struct {
	cancel      context.CancelFunc
	envCfg      env.NatsConfig
	restCfg     *rest.Config
	metricsAddr string
	mgr         manager.Manager
	backend     handlers.JetStreamBackend
	logger      *logger.Logger
}

// NewSubscriptionManager creates the subscription manager for JetStream.
func NewSubscriptionManager(restCfg *rest.Config, metricsAddr string, maxReconnects int, reconnectWait time.Duration, logger *logger.Logger) *SubscriptionManager {
	return &SubscriptionManager{
		envCfg:      env.GetNatsConfig(maxReconnects, reconnectWait),
		restCfg:     restCfg,
		metricsAddr: metricsAddr,
		logger:      logger,
	}
}

// Init initialize the JetStream subscription manager.
func (sm *SubscriptionManager) Init(mgr manager.Manager) error {
	if len(sm.envCfg.URL) == 0 {
		return fmt.Errorf("env var URL must be a non-empty value")
	}
	sm.mgr = mgr
	sm.namedLogger().Info("initialized JetStream subscription manager")
	return nil
}

func (sm *SubscriptionManager) Start(_ env.DefaultSubscriptionConfig, _ subscriptionmanager.Params) error {
	ctx, cancel := context.WithCancel(context.Background())
	sm.cancel = cancel
	jetStreamHandler := handlers.NewJetStream(sm.envCfg, sm.logger)
	dynamicClient := dynamic.NewForConfigOrDie(sm.restCfg)
	applicationLister := application.NewLister(ctx, dynamicClient)
	cleaner := eventtype.NewCleaner(sm.envCfg.EventTypePrefix, applicationLister, sm.logger)

	jetStreamReconciler := jetstream.NewReconciler(
		ctx,
		sm.mgr.GetClient(),
		jetStreamHandler,
		sm.logger,
		sm.mgr.GetEventRecorderFor("eventing-controller-jetstream"),
		cleaner,
	)
	// TODO: this could be refactored (also in other backends), so that the backend is created here and passed to
	//  the reconciler, not the other way around.
	sm.backend = jetStreamReconciler.Backend
	if err := jetStreamReconciler.SetupUnmanaged(sm.mgr); err != nil {
		return fmt.Errorf("unable to setup the NATS subscription controller: %v", err)
	}
	sm.namedLogger().Info("started JetStream subscription manager")
	return nil
}

func (sm *SubscriptionManager) Stop(_ bool) error {
	sm.cancel()
	sm.namedLogger().Info("stopped JetStream subscription manager")
	return nil
}

func (sm *SubscriptionManager) namedLogger() *zap.SugaredLogger {
	return sm.logger.WithContext().Named(subscriptionManagerName)
}
