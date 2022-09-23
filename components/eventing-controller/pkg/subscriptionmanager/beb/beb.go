package beb

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"

	apigatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
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
	"github.com/kyma-project/kyma/components/eventing-controller/controllers/subscription/beb"
	bebv2 "github.com/kyma-project/kyma/components/eventing-controller/controllers/subscriptionv2/beb"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application"
	backendbeb "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/beb"
	backendeventmesh "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/eventmesh"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/eventtype"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/sink"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/utils"
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
	bebBackend       backendbeb.Backend
	eventMeshBackend backendeventmesh.Backend
	logger           *logger.Logger
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

	oauth2credential, err := getOAuth2ClientCredentials(params)
	if err != nil {
		return errors.Wrap(err, "get oauth2client credentials failed")
	}

	// Need to read env to read BEB related secrets
	c.envCfg = env.GetConfig()
	nameMapper := utils.NewBEBSubscriptionNameMapper(strings.TrimSpace(c.envCfg.Domain), backendbeb.MaxBEBSubscriptionNameLength)
	ctrl.Log.WithName("BEB-subscription-manager").Info("using BEB name mapper",
		"domainName", c.envCfg.Domain,
		"maxNameLength", backendbeb.MaxBEBSubscriptionNameLength)

	client := c.mgr.GetClient()
	recorder := c.mgr.GetEventRecorderFor("eventing-controller-beb")
	if c.envCfg.EnableNewCRDVersion {
		eventMeshHandler := backendeventmesh.NewEventMesh(oauth2credential, nameMapper, c.logger)
		eventMeshcleaner := cleaner.NewEventMeshCleaner(c.logger)
		eventMeshReconciler := bebv2.NewReconciler(
			ctx,
			client,
			c.logger,
			recorder,
			c.envCfg,
			eventMeshcleaner,
			eventMeshHandler,
			oauth2credential,
			nameMapper,
			sink.NewValidator(ctx, client, recorder, c.logger),
		)
		c.eventMeshBackend = eventMeshReconciler.Backend
		if err := eventMeshReconciler.SetupUnmanaged(c.mgr); err != nil {
			return xerrors.Errorf("setup EventMesh subscription controller failed: %v", err)
		}
		c.namedLogger().Info("Started v1alpha2 EventMesh subscription manager")
	} else {
		bebHandler := backendbeb.NewBEB(oauth2credential, nameMapper, c.logger)
		eventMeshcleaner := eventtype.NewCleaner(c.envCfg.EventTypePrefix, applicationLister, c.logger)
		bebReconciler := beb.NewReconciler(
			ctx,
			client,
			c.logger,
			recorder,
			c.envCfg,
			eventMeshcleaner,
			bebHandler,
			oauth2credential,
			nameMapper,
			sink.NewValidator(ctx, client, recorder, c.logger),
		)
		c.bebBackend = bebReconciler.Backend
		if err := bebReconciler.SetupUnmanaged(c.mgr); err != nil {
			return xerrors.Errorf("setup BEB subscription controller failed: %v", err)
		}
		c.namedLogger().Info("Started BEB subscription manager")
	}
	return nil
}

// Stop implements the subscriptionmanager.Manager interface and stops the EventMesh subscription manager.
// If runCleanup is false, it will only mark the subscriptions as not ready. If it is true, it will
// clean up subscriptions on EventMesh.
func (c *SubscriptionManager) Stop(runCleanup bool) error {
	if c.cancel != nil {
		c.cancel()
	}
	dynamicClient := dynamic.NewForConfigOrDie(c.restCfg)
	if !runCleanup {
		return markAllSubscriptionsAsNotReady(dynamicClient, c.namedLogger())
	}

	if c.envCfg.EnableNewCRDVersion {
		return cleanupEventMesh(c.eventMeshBackend, dynamicClient, c.namedLogger())
	}

	return cleanup(c.bebBackend, dynamicClient, c.namedLogger())
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

// cleanup removes all created BEB artifacts (based on Subscription v1alpha1).
// This method will be depreciated once Subscription v1alpha2 is active.
func cleanup(backend backendbeb.Backend, dynamicClient dynamic.Interface, logger *zap.SugaredLogger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var bebBackend *backendbeb.BEB
	var ok bool
	if bebBackend, ok = backend.(*backendbeb.BEB); !ok {
		return xerrors.Errorf("no BEB backend exists: convert backend handler to BEB handler failed")
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
	subscriptionsUnstructured, err := dynamicClient.Resource(utils.SubscriptionGroupVersionResource()).Namespace(corev1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrapf(err, "list subscriptions failed")
	}
	subs, err := eventingv1alpha2.ConvertUnstructuredListToSubscriptionList(subscriptionsUnstructured)
	if err != nil {
		return errors.Wrapf(err, "convert subscriptionList from unstructured list failed")
	}

	// Clean APIRules.
	isCleanupSuccessful := true
	for _, v := range subs.Items {
		sub := v
		if apiRule := sub.Status.Backend.APIRuleName; apiRule != "" {
			if err := dynamicClient.Resource(utils.APIRuleGroupVersionResource()).Namespace(sub.Namespace).
				Delete(ctx, apiRule, metav1.DeleteOptions{}); err != nil {
				isCleanupSuccessful = false
				logger.Errorw("Failed to delete APIRule", "namespace", sub.Namespace, "name", apiRule, "error", err)
			}
		}

		// Clean statuses.
		desiredSub := sub.DuplicateWithStatusDefaults()
		if err := utils.UpdateSubscriptionV1Alpha2Status(ctx, dynamicClient, desiredSub); err != nil {
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

func getOAuth2ClientCredentials(params subscriptionmanager.Params) (*backendbeb.OAuth2ClientCredentials, error) {
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
	return &backendbeb.OAuth2ClientCredentials{
		ClientID:     string(id),
		ClientSecret: string(secret),
	}, nil
}

func (c *SubscriptionManager) namedLogger() *zap.SugaredLogger {
	return c.logger.WithContext().Named(subscriptionManagerName)
}
