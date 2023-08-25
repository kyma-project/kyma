package eventmesh

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/metrics"

	apigatewayv1beta1 "github.com/kyma-project/api-gateway/api/v1beta1"

	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/controllers/subscription/eventmesh"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	backendeventmesh "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/eventmesh"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/eventtype"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/sink"
	backendutils "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/utils"
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
	if err := apigatewayv1beta1.AddToScheme(scheme); err != nil {
		return err
	}
	return nil
}

// AddV1Alpha2ToScheme adds v1alpha2 scheme into the given scheme.
func AddV1Alpha2ToScheme(scheme *runtime.Scheme) error {
	if err := eventingv1alpha2.AddToScheme(scheme); err != nil {
		return err
	}
	return nil
}

// SubscriptionManager implements the subscriptionmanager.Manager interface.
type SubscriptionManager struct {
	cancel           context.CancelFunc
	envCfg           env.Config
	restCfg          *rest.Config
	metricsAddr      string
	resyncPeriod     time.Duration
	mgr              manager.Manager
	eventMeshBackend backendeventmesh.Backend
	logger           *logger.Logger
	collector        *metrics.Collector
}

// NewSubscriptionManager creates the SubscriptionManager for BEB and initializes it as far as it
// does not depend on non-common options.
func NewSubscriptionManager(restCfg *rest.Config, metricsAddr string, resyncPeriod time.Duration, logger *logger.Logger,
	collector *metrics.Collector) *SubscriptionManager {
	return &SubscriptionManager{
		envCfg:       env.GetConfig(),
		restCfg:      restCfg,
		metricsAddr:  metricsAddr,
		resyncPeriod: resyncPeriod,
		logger:       logger,
		collector:    collector,
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
	c.collector.ResetSubscriptionStatus()
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel

	oauth2credential, err := getOAuth2ClientCredentials(params)
	if err != nil {
		return errors.Wrap(err, "get oauth2client credentials failed")
	}

	// Need to read env to read BEB related secrets
	c.envCfg = env.GetConfig()
	nameMapper := backendutils.NewBEBSubscriptionNameMapper(strings.TrimSpace(c.envCfg.Domain),
		backendeventmesh.MaxSubscriptionNameLength)
	ctrl.Log.WithName("BEB-subscription-manager").Info("using BEB name mapper",
		"domainName", c.envCfg.Domain,
		"maxNameLength", backendeventmesh.MaxSubscriptionNameLength)

	client := c.mgr.GetClient()
	recorder := c.mgr.GetEventRecorderFor("eventing-controller-beb")

	// Initialize v1alpha1 event type cleaner for conversion webhook
	simpleCleaner := eventtype.NewSimpleCleaner(c.envCfg.EventTypePrefix, c.logger)
	eventingv1alpha1.InitializeEventTypeCleaner(simpleCleaner)

	// Initialize v1alpha2 handler for EventMesh
	eventMeshHandler := backendeventmesh.NewEventMesh(oauth2credential, nameMapper, c.logger)
	eventMeshcleaner := cleaner.NewEventMeshCleaner(c.logger)
	eventMeshReconciler := eventmesh.NewReconciler(
		ctx,
		client,
		c.logger,
		recorder,
		c.envCfg,
		eventMeshcleaner,
		eventMeshHandler,
		oauth2credential,
		nameMapper,
		sink.NewValidator(ctx, client, recorder),
		c.collector,
	)
	c.eventMeshBackend = eventMeshReconciler.Backend
	if err := eventMeshReconciler.SetupUnmanaged(c.mgr); err != nil {
		return xerrors.Errorf("setup EventMesh subscription controller failed: %v", err)
	}
	c.namedLogger().Info("Started v1alpha2 EventMesh subscription manager")

	return nil
}

// Stop implements the subscriptionmanager.Manager interface and stops the EventMesh subscription manager.
// If runCleanup is false, it will only mark the subscriptions as not ready. If it is true, it will
// clean up subscriptions on EventMesh.
func (c *SubscriptionManager) Stop(runCleanup bool) error {
	if c.cancel != nil {
		c.cancel()
	}

	return c.stopEventMeshBackend(runCleanup)
}

// stopEventMeshBackend stops and cleans all EventMesh backend (based on Subscription v1alpha2).
func (c *SubscriptionManager) stopEventMeshBackend(runCleanup bool) error {
	dynamicClient := dynamic.NewForConfigOrDie(c.restCfg)
	if !runCleanup {
		return markAllV1Alpha2SubscriptionsAsNotReady(dynamicClient, c.namedLogger())
	}

	return cleanupEventMesh(c.eventMeshBackend, dynamicClient, c.namedLogger())
}

func markAllV1Alpha2SubscriptionsAsNotReady(dynamicClient dynamic.Interface, logger *zap.SugaredLogger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Fetch all subscriptions.
	subscriptionsUnstructured, err := dynamicClient.Resource(eventingv1alpha2.SubscriptionGroupVersionResource()).Namespace(corev1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrapf(err, "list subscriptions failed")
	}
	subs, err := eventingv1alpha2.ConvertUnstructListToSubList(subscriptionsUnstructured)
	if err != nil {
		return errors.Wrapf(err, "convert subscriptionList from unstructured list failed")
	}
	// Mark all as not ready
	for _, sub := range subs.Items {
		if !sub.Status.Ready {
			continue
		}

		desiredSub := sub.DuplicateWithStatusDefaults()
		desiredSub.Status.Ready = false
		desiredSub.Status.Backend.CopyHashes(sub.Status.Backend)
		if err = backendutils.UpdateSubscriptionStatus(ctx, dynamicClient, desiredSub); err != nil {
			logger.Errorw("Failed to update subscription status", "namespace", sub.Namespace, "name", sub.Name, "error", err)
		}
	}
	return err
}

// cleanupEventMesh removes all created EventMesh artifacts (based on Subscription v1alpha2).
func cleanupEventMesh(backend backendeventmesh.Backend, dynamicClient dynamic.Interface, logger *zap.SugaredLogger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var eventMeshBackend *backendeventmesh.EventMesh
	var ok bool
	if eventMeshBackend, ok = backend.(*backendeventmesh.EventMesh); !ok {
		return xerrors.Errorf("no EventMesh backend exists: convert backend handler to EventMesh handler failed")
	}

	// Fetch all subscriptions.
	subscriptionsUnstructured, err := dynamicClient.Resource(eventingv1alpha2.SubscriptionGroupVersionResource()).Namespace(corev1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrapf(err, "list subscriptions failed")
	}
	subs, err := eventingv1alpha2.ConvertUnstructListToSubList(subscriptionsUnstructured)
	if err != nil {
		return errors.Wrapf(err, "convert subscriptionList from unstructured list failed")
	}

	// Clean APIRules.
	isCleanupSuccessful := true
	for _, v := range subs.Items {
		sub := v
		if apiRule := sub.Status.Backend.APIRuleName; apiRule != "" {
			if err := dynamicClient.Resource(backendutils.APIRuleGroupVersionResource()).Namespace(sub.Namespace).
				Delete(ctx, apiRule, metav1.DeleteOptions{}); err != nil {
				isCleanupSuccessful = false
				logger.Errorw("Failed to delete APIRule", "namespace", sub.Namespace, "name", apiRule, "error", err)
			}
		}

		// Clean statuses.
		desiredSub := sub.DuplicateWithStatusDefaults()
		if err := backendutils.UpdateSubscriptionStatus(ctx, dynamicClient, desiredSub); err != nil {
			isCleanupSuccessful = false
			logger.Errorw("Failed to update EventMesh subscription status", "namespace", sub.Namespace, "name", sub.Name, "error", err)
		}

		// Clean subscriptions from EventMesh.
		if eventMeshBackend != nil {
			if err := eventMeshBackend.DeleteSubscription(&sub); err != nil {
				isCleanupSuccessful = false
				logger.Errorw("Failed to delete EventMesh subscription", "namespace", sub.Namespace, "name", sub.Name, "error", err)
			}
		}
	}

	logger.Debugw("Finished cleanup process", "success", isCleanupSuccessful)
	return nil
}

func getOAuth2ClientCredentials(params subscriptionmanager.Params) (*backendeventmesh.OAuth2ClientCredentials, error) {
	val := params[subscriptionmanager.ParamNameClientID]
	id, ok := val.([]byte)
	if !ok {
		return nil, fmt.Errorf("expected []byte value for %s", subscriptionmanager.ParamNameClientID)
	}

	val = params[subscriptionmanager.ParamNameClientSecret]
	secret, ok := val.([]byte)
	if !ok {
		return nil, fmt.Errorf("expected []byte value for %s", subscriptionmanager.ParamNameClientSecret)
	}

	val = params[subscriptionmanager.ParamNameTokenURL]
	tokenURL, ok := val.([]byte)
	if !ok {
		return nil, fmt.Errorf("expected []byte value for %s", subscriptionmanager.ParamNameTokenURL)
	}

	val = params[subscriptionmanager.ParamNameCertsURL]
	certsURL, ok := val.([]byte)
	if !ok {
		return nil, fmt.Errorf("expected []byte value for %s", subscriptionmanager.ParamNameCertsURL)
	}

	return &backendeventmesh.OAuth2ClientCredentials{
		ClientID:     string(id),
		ClientSecret: string(secret),
		TokenURL:     string(tokenURL),
		CertsURL:     string(certsURL),
	}, nil
}

func (c *SubscriptionManager) namedLogger() *zap.SugaredLogger {
	return c.logger.WithContext().Named(subscriptionManagerName)
}
