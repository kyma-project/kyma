package beb

import (
	"context"
	"fmt"
	"time"

	hydrav1alpha1 "github.com/ory/hydra-maester/api/v1alpha1"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	apigatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/commander"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers"
	"github.com/kyma-project/kyma/components/eventing-controller/reconciler/subscription"
)

const (
	commanderName = "beb-commander"
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
	if err := hydrav1alpha1.AddToScheme(scheme); err != nil {
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
	logger       *logger.Logger
}

// NewCommander creates the Commander for BEB and initializes it as far as it
// does not depend on non-common options.
func NewCommander(restCfg *rest.Config, metricsAddr string, resyncPeriod time.Duration, logger *logger.Logger) *Commander {
	return &Commander{
		envCfg:       env.GetConfig(),
		restCfg:      restCfg,
		metricsAddr:  metricsAddr,
		resyncPeriod: resyncPeriod,
		logger:       logger,
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
func (c *Commander) Start(_ *eventingv1alpha1.SubscriptionConfig, params commander.Params) error {
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	dynamicClient := dynamic.NewForConfigOrDie(c.restCfg)
	applicationLister := application.NewLister(ctx, dynamicClient)
	oauth2credential, err := getOAuth2ClientCredentials(params)
	if err != nil {
		return errors.Wrap(err, "get oauth2client credentials failed")
	}

	// Need to read env so as to read BEB related secrets
	c.envCfg = env.GetConfig()
	reconciler := subscription.NewReconciler(
		ctx,
		c.mgr.GetClient(),
		applicationLister,
		c.mgr.GetCache(),
		c.logger,
		c.mgr.GetEventRecorderFor("eventing-controller-beb"),
		c.envCfg,
		oauth2credential,
	)

	c.backend = reconciler.Backend
	if err := reconciler.SetupUnmanaged(c.mgr); err != nil {
		return fmt.Errorf("setup BEB subscription controller failed: %v", err)
	}
	return nil
}

// Stop implements the Commander interface and stops the commander.
func (c *Commander) Stop() error {
	c.cancel()

	dynamicClient := dynamic.NewForConfigOrDie(c.restCfg)
	return cleanup(c.backend, dynamicClient, c.namedLogger())
}

func (c *Commander) namedLogger() *zap.SugaredLogger {
	return c.logger.WithContext().Named(commanderName)
}

// cleanup removes all created BEB artifacts.
func cleanup(backend handlers.MessagingBackend, dynamicClient dynamic.Interface, logger *zap.SugaredLogger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var bebBackend *handlers.Beb
	var ok bool
	if bebBackend, ok = backend.(*handlers.Beb); !ok {
		err := errors.New("convert backend handler to BEB handler failed")
		logger.Errorw("no BEB backend exists", "error", err)
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

	// Clean APIRules.
	isCleanupSuccessful := true
	for _, sub := range subs.Items {
		if apiRule := sub.Status.APIRuleName; apiRule != "" {
			if err := dynamicClient.Resource(handlers.APIRuleGroupVersionResource()).Namespace(sub.Namespace).
				Delete(ctx, apiRule, metav1.DeleteOptions{}); err != nil {
				isCleanupSuccessful = false
				logger.Errorw("delete APIRule failed", "namespace", sub.Namespace, "name", apiRule, "error", err)
			}
		}

		// Clean statuses.
		desiredSub := handlers.RemoveStatus(sub)
		if err := handlers.UpdateSubscriptionStatus(ctx, dynamicClient, desiredSub); err != nil {
			isCleanupSuccessful = false
			logger.Errorw("update BEB subscription status failed", "namespace", sub.Namespace, "name", sub.Name, "error", err)
		}

		// Clean subscriptions from BEB.
		if bebBackend != nil {
			if err := bebBackend.DeleteSubscription(&sub); err != nil {
				isCleanupSuccessful = false
				logger.Errorw("delete BEB subscription failed", "namespace", sub.Namespace, "name", sub.Name, "error", err)
			}
		}
	}

	logger.Debugw("cleanup process finished", "success", isCleanupSuccessful)
	return nil
}

func getOAuth2ClientCredentials(params commander.Params) (*handlers.OAuth2ClientCredentials, error) {
	val := params["client_id"]
	id, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("expected string value for client_id, but received %T", val)
	}
	val = params["client_secret"]
	secret, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("expected string value for client_secret, but received %T", val)
	}
	return &handlers.OAuth2ClientCredentials{
		ClientID:     id,
		ClientSecret: secret,
	}, nil
}
