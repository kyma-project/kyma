package beb

import (
	"context"
	"fmt"
	"time"

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

	apigatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers"
	"github.com/kyma-project/kyma/components/eventing-controller/reconciler/subscription"
)

// AddToScheme adds the own schemes to the runtime scheme.
func AddToScheme(scheme *runtime.Scheme) error {
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return err
	}
	if err := eventingv1alpha1.AddToScheme(scheme); err != nil {
		return err
	}
	if err := apigatewayv1alpha1.AddToScheme(scheme); err != nil {
		return err
	}
	return nil
}

// Commander implements the Commander interface.
type Commander struct {
	cancel       context.CancelFunc
	envCfg       env.Config
	restCfg      *rest.Config
	metricsAddr  string
	resyncPeriod time.Duration
	mgr          manager.Manager
	backend      handlers.MessagingBackend
}

// NewCommander creates the Commander for BEB and initializes it as far as it
// does not depend on non-common options.
func NewCommander(restCfg *rest.Config, metricsAddr string, resyncPeriod time.Duration) *Commander {
	return &Commander{
		envCfg:       env.GetConfig(),
		restCfg:      restCfg,
		metricsAddr:  metricsAddr,
		resyncPeriod: resyncPeriod,
	}
}

// Init implements the Commander interface.
func (c *Commander) Init(mgr manager.Manager) error {
	if len(c.envCfg.Domain) == 0 {
		return fmt.Errorf("env var DOMAIN must be a non-empty value")
	}
	c.mgr = mgr
	return nil
}

// Start implements the Commander interface and starts the manager.
func (c *Commander) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	dynamicClient := dynamic.NewForConfigOrDie(c.restCfg)
	applicationLister := application.NewLister(ctx, dynamicClient)

	// Need to read env so as to read BEB related secrets
	c.envCfg = env.GetConfig()
	reconciler := subscription.NewReconciler(
		ctx,
		c.mgr.GetClient(),
		applicationLister,
		c.mgr.GetCache(),
		ctrl.Log.WithName("reconciler").WithName("Subscription"),
		c.mgr.GetEventRecorderFor("eventing-controller-beb"),
		c.envCfg,
	)

	c.backend = reconciler.Backend
	if err := reconciler.SetupUnmanaged(c.mgr); err != nil {
		return fmt.Errorf("unable to setup the BEB Subscription Controller: %v", err)
	}
	return nil
}

// Stop implements the Commander interface and stops the commander.
func (c *Commander) Stop() error {
	c.cancel()

	dynamicClient := dynamic.NewForConfigOrDie(c.restCfg)
	return cleanup(c.backend, dynamicClient)
}

// cleanup removes all created BEB artifacts.
func cleanup(backend handlers.MessagingBackend, dynamicClient dynamic.Interface) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger := ctrl.Log.WithName("eventing-controller-beb-cleaner").WithName("Subscription")
	var bebBackend *handlers.Beb
	var ok bool
	var bebBackendErr error
	if bebBackend, ok = backend.(*handlers.Beb); !ok {
		bebBackendErr = errors.New("failed to convert backend to handlers.Beb")
		logger.Error(bebBackendErr, "no BEB backend exists")
	}

	// Fetch all subscriptions.
	//dynamicClient := dynamic.NewForConfigOrDie(c.restCfg)
	subscriptionsUnstructured, err := dynamicClient.Resource(handlers.GroupVersionResource()).Namespace(corev1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	subs, err := handlers.ToSubscriptionList(subscriptionsUnstructured)
	if err != nil {
		return err
	}

	statusDeletionResult := make(map[string]error)
	subDeletionResult := make(map[string]error)
	apiRuleDeletionResult := make(map[string]error)

	for _, sub := range subs.Items {
		// Clean APIRules.
		apiRule := sub.Status.APIRuleName
		keyAPIRule := types.NamespacedName{
			Namespace: sub.Namespace,
			Name:      apiRule,
		}
		if apiRule != "" {
			err := dynamicClient.Resource(handlers.APIRuleGroupVersionResource()).Namespace(sub.Namespace).Delete(ctx, apiRule, metav1.DeleteOptions{})
			if err != nil {
				apiRuleDeletionResult[keyAPIRule.String()] = err
			}
		}

		// Clean statuses.
		key := types.NamespacedName{
			Namespace: sub.Namespace,
			Name:      sub.Name,
		}
		desiredSub := handlers.RemoveStatus(sub)
		err := handlers.UpdateSubscription(ctx, dynamicClient, desiredSub)
		if err != nil {
			statusDeletionResult[key.String()] = err
		}

		// Clean subscriptions from BEB.
		if bebBackend != nil {
			err = bebBackend.DeleteSubscription(&sub)
			if err != nil {
				subDeletionResult[key.String()] = err
			}
		} else {
			subDeletionResult[key.String()] = bebBackendErr
		}
	}

	if len(apiRuleDeletionResult) == 0 {
		logger.Info("Deletion of APIRules succeeded")
	} else {
		logger.Info("Deletion of APIRules failed: %+v", apiRuleDeletionResult)
	}

	if len(subDeletionResult) == 0 {
		logger.Info("Deletion of Subscriptions in BEB succeeded")
	} else {
		logger.Info("Deletion of Subscriptions in BEB failed: %+v", subDeletionResult)
	}

	if len(statusDeletionResult) == 0 {
		logger.Info("Deletion of Statuses in Subscriptions succeeded")
	} else {
		logger.Info("Deletion of Statuses in Subscriptions failed: %+v", statusDeletionResult)
	}

	return nil
}
