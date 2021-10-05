package nats

import (
	"context"
	"fmt"
	"time"

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
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/commander"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers"
	subscription "github.com/kyma-project/kyma/components/eventing-controller/reconciler/subscription/nats"
)

const (
	commanderName = "nats-commander"
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

// Commander implements the Commander interface.
type Commander struct {
	cancel      context.CancelFunc
	envCfg      env.NatsConfig
	restCfg     *rest.Config
	metricsAddr string
	mgr         manager.Manager
	backend     handlers.MessagingBackend
	logger      *logger.Logger
}

// NewCommander creates the Commander for BEB and initializes it as far as it
// does not depend on non-common options.
func NewCommander(restCfg *rest.Config, metricsAddr string, maxReconnects int, reconnectWait time.Duration, logger *logger.Logger) *Commander {
	return &Commander{
		envCfg:      env.GetNatsConfig(maxReconnects, reconnectWait), // TODO Harmonization.
		restCfg:     restCfg,
		metricsAddr: metricsAddr,
		logger:      logger,
	}
}

// Init implements the Commander interface.
func (c *Commander) Init(mgr manager.Manager) error {
	if len(c.envCfg.URL) == 0 {
		return fmt.Errorf("env var URL must be a non-empty value")
	}
	c.mgr = mgr
	return nil
}

// Start implements the Commander interface and starts the commander.
func (c *Commander) Start(defaultSubsConfig env.DefaultSubscriptionConfig, _ commander.Params) error {
	ctx, cancel := context.WithCancel(context.Background())

	c.cancel = cancel
	dynamicClient := dynamic.NewForConfigOrDie(c.restCfg)
	applicationLister := application.NewLister(ctx, dynamicClient)
	natsReconciler := subscription.NewReconciler(
		ctx,
		c.mgr.GetClient(),
		applicationLister,
		c.mgr.GetCache(),
		c.logger,
		c.mgr.GetEventRecorderFor("eventing-controller-nats"),
		c.envCfg,
		defaultSubsConfig,
	)
	c.backend = natsReconciler.Backend
	if err := natsReconciler.SetupUnmanaged(c.mgr); err != nil {
		return fmt.Errorf("unable to setup the NATS subscription controller: %v", err)
	}
	return nil
}

// Stop implements the Commander interface and stops the commander.
func (c *Commander) Stop() error {
	c.cancel()

	dynamicClient := dynamic.NewForConfigOrDie(c.restCfg)
	return cleanup(c.backend, dynamicClient, c.namedLogger())
}

// clean removes all NATS artifacts.
func cleanup(backend handlers.MessagingBackend, dynamicClient dynamic.Interface, logger *zap.SugaredLogger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var ok bool

	var natsBackend *handlers.Nats
	if natsBackend, ok = backend.(*handlers.Nats); !ok {
		err := errors.New("convert backend handler to NATS handler failed")
		logger.Errorw("no NATS backend exists", "error", err)
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

		desiredSub := handlers.RemoveStatus(sub)
		if err := handlers.UpdateSubscriptionStatus(ctx, dynamicClient, desiredSub); err != nil {
			isCleanupSuccessful = false
			log.Errorw("update NATS subscription status failed", "error", err)
		}

		// Clean subscriptions from NATS.
		if natsBackend != nil {
			if err := natsBackend.DeleteSubscription(&sub); err != nil {
				isCleanupSuccessful = false
				log.Errorw("delete NATS subscription failed", "error", err)
			}
		}
	}

	logger.Debugw("cleanup process finished", "success", isCleanupSuccessful)
	return nil
}

func (c *Commander) namedLogger() *zap.SugaredLogger {
	return c.logger.WithContext().Named(commanderName)
}
