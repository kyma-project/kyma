package nats

import (
	"context"
	"fmt"
	"time"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/commander"

	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers"
	subscription "github.com/kyma-project/kyma/components/eventing-controller/reconciler/subscription-nats"
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
}

// NewCommander creates the Commander for BEB and initializes it as far as it
// does not depend on non-common options.
func NewCommander(restCfg *rest.Config, metricsAddr string, maxReconnects int, reconnectWait time.Duration) *Commander {
	return &Commander{
		envCfg:      env.GetNatsConfig(maxReconnects, reconnectWait), // TODO Harmonization.
		restCfg:     restCfg,
		metricsAddr: metricsAddr,
	}
}

// Init implements the Commander interface.
func (c *Commander) Init(mgr manager.Manager) error {
	if len(c.envCfg.Url) == 0 {
		return fmt.Errorf("env var URL must be a non-empty value")
	}
	c.mgr = mgr
	return nil
}

// Start implements the Commander interface and starts the commander.
func (c *Commander) Start(_ commander.Params) error {
	ctx, cancel := context.WithCancel(context.Background())

	c.cancel = cancel
	dynamicClient := dynamic.NewForConfigOrDie(c.restCfg)
	applicationLister := application.NewLister(ctx, dynamicClient)
	natsReconciler := subscription.NewReconciler(
		ctx,
		c.mgr.GetClient(),
		applicationLister,
		c.mgr.GetCache(),
		ctrl.Log.WithName("reconciler").WithName("Subscription"),
		c.mgr.GetEventRecorderFor("eventing-controller-nats"),
		c.envCfg,
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
	return cleanup(c.backend, dynamicClient)
}

// clean removes all NATS artifacts.
func cleanup(backend handlers.MessagingBackend, dynamicClient dynamic.Interface) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger := ctrl.Log.WithName("eventing-controller-nats-cleaner").WithName("Subscription")
	var natsBackend *handlers.Nats
	var ok bool
	isCleanupSuccessful := true
	if natsBackend, ok = backend.(*handlers.Nats); !ok {
		isCleanupSuccessful = false
		natsBackendErr := errors.New("failed to convert backend to handlers.Nats")
		logger.Error(natsBackendErr, "no NATS backend exists")
	}
	// Fetch all subscriptions.
	subscriptionsUnstructured, err := dynamicClient.Resource(handlers.SubscriptionGroupVersionResource()).Namespace(corev1.NamespaceAll).List(ctx, metav1.ListOptions{})

	if err != nil {
		return errors.Wrapf(err, "failed to list subscriptions")
	}
	subs, err := handlers.ToSubscriptionList(subscriptionsUnstructured)
	if err != nil {
		return errors.Wrapf(err, "failed to convert to subscriptionList from unstructured list")
	}

	for _, sub := range subs.Items {
		// Clean statuses.
		subKey := types.NamespacedName{
			Namespace: sub.Namespace,
			Name:      sub.Name,
		}
		desiredSub := handlers.RemoveStatus(sub)
		err := handlers.UpdateSubscriptionStatus(ctx, dynamicClient, desiredSub)
		if err != nil {
			isCleanupSuccessful = false
			logger.Error(err, fmt.Sprintf("failed to update status of Subscription: %s", subKey.String()))
		}

		// Clean subscriptions from NATS.
		if natsBackend != nil {
			err = natsBackend.DeleteSubscription(&sub)
			if err != nil {
				isCleanupSuccessful = false
				logger.Error(err, fmt.Sprintf("failed to update status of Subscription: %s", subKey.String()))
			}
		}
	}

	if isCleanupSuccessful {
		logger.Info("Cleanup process succeeded!")
	}
	return nil
}
