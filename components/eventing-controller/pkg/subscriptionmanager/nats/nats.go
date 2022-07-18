package nats

import (
	"context"
	"fmt"

	pkgmetrics "github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/metrics"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/sink"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/eventtype"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	subscription "github.com/kyma-project/kyma/components/eventing-controller/controllers/subscription/nats"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/subscriptionmanager"
)

const (
	subscriptionManagerName = "nats-subscription-manager"
)

// AddToScheme adds the own schemes to the runtime scheme.
func AddToScheme(scheme *runtime.Scheme) error {
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return err
	}
	if err := eventingv1alpha1.AddToScheme(scheme); err != nil {
		return err
	}
	return nil
}

// SubscriptionManager implements the subscriptionmanager.Manager interface.
type SubscriptionManager struct {
	cancel           context.CancelFunc
	envCfg           env.NatsConfig
	restCfg          *rest.Config
	metricsAddr      string
	metricsCollector *pkgmetrics.Collector
	mgr              manager.Manager
	backend          handlers.NatsBackend
	logger           *logger.Logger
}

// NewSubscriptionManager creates the subscription manager for BEB and initializes it as far as it
// does not depend on non-common options.
func NewSubscriptionManager(restCfg *rest.Config, natsConfig env.NatsConfig, metricsAddr string, metricsCollector *pkgmetrics.Collector, logger *logger.Logger) *SubscriptionManager {
	return &SubscriptionManager{
		envCfg:           natsConfig,
		restCfg:          restCfg,
		metricsAddr:      metricsAddr,
		metricsCollector: metricsCollector,
		logger:           logger,
	}
}

// Init implements the subscriptionmanager.Manager interface.
func (c *SubscriptionManager) Init(mgr manager.Manager) error {
	if len(c.envCfg.URL) == 0 {
		return fmt.Errorf("env var URL must be a non-empty value")
	}
	c.mgr = mgr
	return nil
}

// Start implements the subscriptionmanager.Manager interface for the NATS subscription manager.
func (c *SubscriptionManager) Start(defaultSubsConfig env.DefaultSubscriptionConfig, _ subscriptionmanager.Params) error {
	ctx, cancel := context.WithCancel(context.Background())

	c.cancel = cancel
	client := c.mgr.GetClient()
	recorder := c.mgr.GetEventRecorderFor("eventing-controller-nats")
	dynamicClient := dynamic.NewForConfigOrDie(c.restCfg)
	applicationLister := application.NewLister(ctx, dynamicClient)
	natsHandler := handlers.NewNats(c.envCfg, defaultSubsConfig, c.metricsCollector, c.logger)
	cleaner := eventtype.NewCleaner(c.envCfg.EventTypePrefix, applicationLister, c.logger)
	natsReconciler := subscription.NewReconciler(
		ctx,
		client,
		natsHandler,
		cleaner,
		c.logger,
		recorder,
		defaultSubsConfig,
		sink.NewValidator(ctx, client, recorder, c.logger),
	)
	c.backend = natsReconciler.Backend
	if err := natsReconciler.SetupUnmanaged(c.mgr); err != nil {
		return fmt.Errorf("unable to setup the NATS subscription controller: %v", err)
	}
	return nil
}

// Stop implements the subscriptionmanager.Manager interface and stops the NATS subscription manager.
// It cleans up the subscriptions on NATS, if the runCleanup flag is true.
func (c *SubscriptionManager) Stop(runCleanup bool) error {
	c.cancel()
	if !runCleanup {
		return nil
	}
	dynamicClient := dynamic.NewForConfigOrDie(c.restCfg)
	return cleanup(c.backend, dynamicClient, c.namedLogger())
}

// clean removes all NATS artifacts.
func cleanup(backend handlers.NatsBackend, dynamicClient dynamic.Interface, logger *zap.SugaredLogger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var ok bool

	var natsBackend *handlers.Nats
	if natsBackend, ok = backend.(*handlers.Nats); !ok {
		err := errors.New("convert backend handler to NATS handler failed")
		logger.Errorw("No NATS backend exists", "error", err)
		return err
	}

	// Fetch all subscriptions.
	subscriptionsUnstructured, err := dynamicClient.Resource(handlers.SubscriptionGroupVersionResource()).Namespace(corev1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrapf(err, "list subscriptions failed")
	}

	subs, err := handlers.ToSubscriptionList(subscriptionsUnstructured)
	if err != nil {
		return errors.Wrapf(err, "convert subscriptionList from unstructured list failed")
	}

	// Clean statuses.
	isCleanupSuccessful := true
	for _, v := range subs.Items {
		sub := v
		subKey := types.NamespacedName{Namespace: sub.Namespace, Name: sub.Name}
		log := logger.With("key", subKey.String())

		desiredSub := handlers.ResetStatusToDefaults(sub)
		if err := handlers.UpdateSubscriptionStatus(ctx, dynamicClient, desiredSub); err != nil {
			isCleanupSuccessful = false
			log.Errorw("Failed to update NATS subscription status", "error", err)
		}

		// Clean subscriptions from NATS.
		if natsBackend != nil {
			if err := natsBackend.DeleteSubscription(&sub); err != nil {
				isCleanupSuccessful = false
				log.Errorw("Failed to delete NATS subscription", "error", err)
			}
		}
	}

	logger.Debugw("Finished cleanup process", "success", isCleanupSuccessful)
	return nil
}

func (c *SubscriptionManager) namedLogger() *zap.SugaredLogger {
	return c.logger.WithContext().Named(subscriptionManagerName)
}
