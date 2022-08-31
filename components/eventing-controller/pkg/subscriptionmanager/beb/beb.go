package beb

import (
	"context"
	"fmt"
	"strings"
	"time"

	beb2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/beb"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/sink"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/utils"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/eventtype"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	apigatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/controllers/subscription/beb"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/subscriptionmanager"
)

const (
	subscriptionManagerName = "beb-subscription-manager"
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

// SubscriptionManager implements the subscriptionmanager.Manager interface.
type SubscriptionManager struct {
	cancel       context.CancelFunc
	envCfg       env.Config
	restCfg      *rest.Config
	metricsAddr  string
	resyncPeriod time.Duration
	mgr          manager.Manager
	backend      beb2.Backend
	logger       *logger.Logger
}

// NewSubscriptionManager creates the SubscriptionManager for BEB and initializes it as far as it
// does not depend on non-common options.
func NewSubscriptionManager(restCfg *rest.Config, metricsAddr string, resyncPeriod time.Duration, logger *logger.Logger) *SubscriptionManager {
	return &SubscriptionManager{
		envCfg:       env.GetConfig(),
		restCfg:      restCfg,
		metricsAddr:  metricsAddr,
		resyncPeriod: resyncPeriod,
		logger:       logger,
	}
}

// Init implements the subscriptionmanager.Manager interface.
func (c *SubscriptionManager) Init(mgr manager.Manager) error {
	if len(c.envCfg.Domain) == 0 {
		return fmt.Errorf("env var DOMAIN must be a non-empty value")
	}
	c.mgr = mgr
	return nil
}

// Start implements the subscriptionmanager.Manager interface and starts the manager.
func (c *SubscriptionManager) Start(_ env.DefaultSubscriptionConfig, params subscriptionmanager.Params) error {
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	dynamicClient := dynamic.NewForConfigOrDie(c.restCfg)
	applicationLister := application.NewLister(ctx, dynamicClient)
	cleaner := eventtype.NewCleaner(c.envCfg.EventTypePrefix, applicationLister, c.logger)

	oauth2credential, err := getOAuth2ClientCredentials(params)
	if err != nil {
		return errors.Wrap(err, "get oauth2client credentials failed")
	}

	// Need to read env to read BEB related secrets
	c.envCfg = env.GetConfig()
	nameMapper := utils.NewBEBSubscriptionNameMapper(strings.TrimSpace(c.envCfg.Domain), beb2.MaxBEBSubscriptionNameLength)
	ctrl.Log.WithName("BEB-subscription-manager").Info("using BEB name mapper",
		"domainName", c.envCfg.Domain,
		"maxNameLength", beb2.MaxBEBSubscriptionNameLength)
	bebHandler := beb2.NewBEB(oauth2credential, nameMapper, c.logger)

	client := c.mgr.GetClient()
	recorder := c.mgr.GetEventRecorderFor("eventing-controller-beb")
	bebReconciler := beb.NewReconciler(
		ctx,
		client,
		c.logger,
		recorder,
		c.envCfg,
		cleaner,
		bebHandler,
		oauth2credential,
		nameMapper,
		sink.NewValidator(ctx, client, recorder, c.logger),
	)
	c.backend = bebReconciler.Backend
	if err := bebReconciler.SetupUnmanaged(c.mgr); err != nil {
		return fmt.Errorf("setup BEB subscription controller failed: %v", err)
	}
	return nil
}

// Stop implements the subscriptionmanager.Manager interface and stops the BEB subscription manager.
// If runCleanup is false, it will only mark the subscriptions as not ready. If it is true, it will
// clean up subscriptions on BEB.
func (c *SubscriptionManager) Stop(runCleanup bool) error {
	if c.cancel != nil {
		c.cancel()
	}
	dynamicClient := dynamic.NewForConfigOrDie(c.restCfg)
	if !runCleanup {
		return markAllSubscriptionsAsNotReady(dynamicClient, c.namedLogger())
	}
	return cleanup(c.backend, dynamicClient, c.namedLogger())
}

func markAllSubscriptionsAsNotReady(dynamicClient dynamic.Interface, logger *zap.SugaredLogger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Fetch all subscriptions.
	subscriptionsUnstructured, err := dynamicClient.Resource(utils.SubscriptionGroupVersionResource()).Namespace(corev1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrapf(err, "list subscriptions failed")
	}
	subs, err := utils.ToSubscriptionList(subscriptionsUnstructured)
	if err != nil {
		return errors.Wrapf(err, "convert subscriptionList from unstructured list failed")
	}
	// Mark all as not ready
	for _, sub := range subs.Items {
		if !sub.Status.Ready {
			continue
		}
		desiredSub := utils.SetStatusAsNotReady(sub)
		if err = utils.UpdateSubscriptionStatus(ctx, dynamicClient, desiredSub); err != nil {
			logger.Errorw("Failed to update BEB subscription status", "namespace", sub.Namespace, "name", sub.Name, "error", err)
		}
	}
	return err
}

// cleanup removes all created BEB artifacts.
func cleanup(backend beb2.Backend, dynamicClient dynamic.Interface, logger *zap.SugaredLogger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var bebBackend *beb2.BEB
	var ok bool
	if bebBackend, ok = backend.(*beb2.BEB); !ok {
		err := errors.New("convert backend handler to BEB handler failed")
		logger.Errorw("No BEB backend exists", "error", err)
		return err
	}

	// Fetch all subscriptions.
	subscriptionsUnstructured, err := dynamicClient.Resource(utils.SubscriptionGroupVersionResource()).Namespace(corev1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrapf(err, "list subscriptions failed")
	}
	subs, err := utils.ToSubscriptionList(subscriptionsUnstructured)
	if err != nil {
		return errors.Wrapf(err, "convert subscriptionList from unstructured list failed")
	}

	// Clean APIRules.
	isCleanupSuccessful := true
	for _, v := range subs.Items {
		sub := v
		if apiRule := sub.Status.APIRuleName; apiRule != "" {
			if err := dynamicClient.Resource(utils.APIRuleGroupVersionResource()).Namespace(sub.Namespace).
				Delete(ctx, apiRule, metav1.DeleteOptions{}); err != nil {
				isCleanupSuccessful = false
				logger.Errorw("Failed to delete APIRule", "namespace", sub.Namespace, "name", apiRule, "error", err)
			}
		}

		// Clean statuses.
		desiredSub := utils.ResetStatusToDefaults(sub)
		if err := utils.UpdateSubscriptionStatus(ctx, dynamicClient, desiredSub); err != nil {
			isCleanupSuccessful = false
			logger.Errorw("Failed to update BEB subscription status", "namespace", sub.Namespace, "name", sub.Name, "error", err)
		}

		// Clean subscriptions from BEB.
		if bebBackend != nil {
			if err := bebBackend.DeleteSubscription(&sub); err != nil {
				isCleanupSuccessful = false
				logger.Errorw("Failed to delete BEB subscription", "namespace", sub.Namespace, "name", sub.Name, "error", err)
			}
		}
	}

	logger.Debugw("Finished cleanup process", "success", isCleanupSuccessful)
	return nil
}

func getOAuth2ClientCredentials(params subscriptionmanager.Params) (*beb2.OAuth2ClientCredentials, error) {
	val := params["client_id"]
	id, ok := val.([]byte)
	if !ok {
		return nil, fmt.Errorf("expected []byte value for client_id, but received %T", val)
	}
	val = params["client_secret"]
	secret, ok := val.([]byte)
	if !ok {
		return nil, fmt.Errorf("expected []byte value for client_secret, but received %T", val)
	}
	return &beb2.OAuth2ClientCredentials{
		ClientID:     string(id),
		ClientSecret: string(secret),
	}, nil
}

func (c *SubscriptionManager) namedLogger() *zap.SugaredLogger {
	return c.logger.WithContext().Named(subscriptionManagerName)
}
