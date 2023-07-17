package jetstream

import (
	"context"
	"fmt"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/sink"
	backendutils "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/utils"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"

	"golang.org/x/xerrors"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"

	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/controllers/subscription/jetstream"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/eventtype"
	backendjetstream "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/jetstream"
	backendmetrics "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/metrics"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/subscriptionmanager"
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

// AddV1Alpha2ToScheme adds v1alpha2 scheme into the given scheme.
func AddV1Alpha2ToScheme(scheme *runtime.Scheme) error {
	if err := eventingv1alpha2.AddToScheme(scheme); err != nil {
		return err
	}
	return nil
}

type SubscriptionManager struct {
	cancel           context.CancelFunc
	envCfg           env.NATSConfig
	restCfg          *rest.Config
	metricsAddr      string
	metricsCollector *backendmetrics.Collector
	mgr              manager.Manager
	backendv2        backendjetstream.Backend
	logger           *logger.Logger
}

// NewSubscriptionManager creates the subscription manager for JetStream.
func NewSubscriptionManager(restCfg *rest.Config, natsConfig env.NATSConfig, metricsAddr string,
	metricsCollector *backendmetrics.Collector, logger *logger.Logger) *SubscriptionManager {
	return &SubscriptionManager{
		envCfg:           natsConfig,
		restCfg:          restCfg,
		metricsAddr:      metricsAddr,
		metricsCollector: metricsCollector,
		logger:           logger,
	}
}

// Init initialize the JetStream subscription manager.
func (sm *SubscriptionManager) Init(mgr manager.Manager) error {
	if len(sm.envCfg.URL) == 0 {
		return xerrors.Errorf("env var URL must be a non-empty value")
	}
	sm.mgr = mgr
	sm.namedLogger().Info("initialized JetStream subscription manager")
	return nil
}

func (sm *SubscriptionManager) Start(defaultSubsConfig env.DefaultSubscriptionConfig, _ subscriptionmanager.Params) error {
	ctx, cancel := context.WithCancel(context.Background())
	sm.cancel = cancel

	client := sm.mgr.GetClient()
	recorder := sm.mgr.GetEventRecorderFor("eventing-controller-jetstream")

	// Initialize v1alpha1 event type cleaner for conversion webhook
	simpleCleaner := eventtype.NewSimpleCleaner(sm.envCfg.EventTypePrefix, sm.logger)
	eventingv1alpha1.InitializeEventTypeCleaner(simpleCleaner)

	// Initialize v1alpha2 event type cleaner
	jsCleaner := cleaner.NewJetStreamCleaner(sm.logger)
	jetStreamHandler := backendjetstream.NewJetStream(sm.envCfg,
		sm.metricsCollector, jsCleaner, defaultSubsConfig, sm.logger)
	jetStreamReconciler := jetstream.NewReconciler(
		ctx,
		client,
		jetStreamHandler,
		sm.logger,
		recorder,
		jsCleaner,
		sink.NewValidator(ctx, client, recorder),
		sm.metricsCollector,
	)
	sm.backendv2 = jetStreamReconciler.Backend

	// delete dangling invalid consumers here
	var subs eventingv1alpha2.SubscriptionList
	if err := client.List(context.Background(), &subs); err != nil {
		return fmt.Errorf("failed to get all subscription resources: %w", err)
	}
	if err := jetStreamHandler.DeleteInvalidConsumers(subs.Items); err != nil {
		return err
	}

	// start the subscription controller
	if err := jetStreamReconciler.SetupUnmanaged(sm.mgr); err != nil {
		return xerrors.Errorf("unable to setup the NATS subscription controller: %v", err)
	}
	sm.namedLogger().Info("Started v1alpha2 JetStream subscription manager")

	return nil
}

func (sm *SubscriptionManager) Stop(runCleanup bool) error {
	sm.cancel()
	if !runCleanup {
		return nil
	}
	dynamicClient := dynamic.NewForConfigOrDie(sm.restCfg)

	return cleanupv2(sm.backendv2, dynamicClient, sm.namedLogger())
}

// clean removes all JetStream artifacts.
func cleanupv2(backend backendjetstream.Backend, dynamicClient dynamic.Interface, logger *zap.SugaredLogger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var ok bool
	var jsBackend *backendjetstream.JetStream
	if jsBackend, ok = backend.(*backendjetstream.JetStream); !ok {
		err := errors.New("converting backend to JetStream v2 backend failed")
		return err
	}

	// fetch all subscriptions.
	subscriptionsUnstructured, err := dynamicClient.Resource(
		eventingv1alpha2.SubscriptionGroupVersionResource()).Namespace(corev1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrapf(err, "list subscriptions failed")
	}

	subs, err := eventingv1alpha2.ConvertUnstructListToSubList(subscriptionsUnstructured)
	if err != nil {
		return errors.Wrapf(err, "convert subscriptionList from unstructured list failed")
	}

	// clean all status.
	isCleanupSuccessful := true
	for _, v := range subs.Items {
		sub := v
		subKey := types.NamespacedName{Namespace: sub.Namespace, Name: sub.Name}
		log := logger.With("key", subKey.String())

		desiredSub := sub.DuplicateWithStatusDefaults()
		if updateErr := backendutils.UpdateSubscriptionStatus(ctx, dynamicClient, desiredSub); updateErr != nil {
			isCleanupSuccessful = false
			log.Errorw("Failed to update JetStream v2 subscription status", "error", err)
		}

		// clean subscriptions from JetStream.
		if jsBackend != nil {
			if delErr := jsBackend.DeleteSubscription(&sub); delErr != nil {
				isCleanupSuccessful = false
				log.Errorw("Failed to delete JetStream v2 subscription", "error", delErr)
			}
		}
	}

	logger.Debugw("Finished cleanup process", "success", isCleanupSuccessful)
	return nil
}

func (sm *SubscriptionManager) namedLogger() *zap.SugaredLogger {
	return sm.logger.WithContext().Named(subscriptionManagerName)
}
