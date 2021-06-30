package beb

import (
	"context"
	"fmt"
	"time"

	hydrav1alpha1 "github.com/ory/hydra-maester/api/v1alpha1"

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
func (c *Commander) Start(params commander.Params) error {
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	dynamicClient := dynamic.NewForConfigOrDie(c.restCfg)
	applicationLister := application.NewLister(ctx, dynamicClient)
	oauth2credential, err := getOAuth2ClientCredentials(params)
	if err != nil {
		return errors.Wrap(err, "cannot get oauth2client credentials")
	}

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
		oauth2credential,
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
	isCleanupSuccessful := true
	if bebBackend, ok = backend.(*handlers.Beb); !ok {
		isCleanupSuccessful = false
		bebBackendErr := errors.New("failed to convert backend to handlers.Beb")
		logger.Error(bebBackendErr, "no BEB backend exists")
		return bebBackendErr
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
		// Clean APIRules.
		apiRule := sub.Status.APIRuleName
		keyAPIRule := types.NamespacedName{
			Namespace: sub.Namespace,
			Name:      apiRule,
		}
		if apiRule != "" {
			err := dynamicClient.Resource(handlers.APIRuleGroupVersionResource()).Namespace(sub.Namespace).Delete(ctx, apiRule, metav1.DeleteOptions{})
			if err != nil {
				isCleanupSuccessful = false
				logger.Error(err, fmt.Sprintf("failed to delete APIRule: %s", keyAPIRule.String()))
			}
		}

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

		// Clean subscriptions from BEB.
		if bebBackend != nil {
			err = bebBackend.DeleteSubscription(&sub)
			if err != nil {
				isCleanupSuccessful = false
				logger.Error(err, fmt.Sprintf("failed to delete Subscription: %s in BEB", subKey.String()))
			}
		}
	}

	if isCleanupSuccessful {
		logger.Info("Cleanup process succeeded!")
	}

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
