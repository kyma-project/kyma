package jetstream

import (
	"context"

	sinkv2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/sink/v2"
	backendutilsv2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/utils/v2"

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
	jetstreamv2 "github.com/kyma-project/kyma/components/eventing-controller/controllers/subscriptionv2/jetstream"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/eventtype"
	backendjetstreamv2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/jetstreamv2"
	backendmetrics "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/metrics"
	backendjetstream "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/nats/jetstream"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/sink"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/utils"
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
	envCfg           env.NatsConfig
	restCfg          *rest.Config
	metricsAddr      string
	metricsCollector *backendmetrics.Collector
	mgr              manager.Manager
	backend          backendjetstream.Backend
	backendv2        backendjetstreamv2.Backend
	logger           *logger.Logger
}

// NewSubscriptionManager creates the subscription manager for JetStream.
func NewSubscriptionManager(restCfg *rest.Config, natsConfig env.NatsConfig, metricsAddr string, metricsCollector *backendmetrics.Collector, logger *logger.Logger) *SubscriptionManager {
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
	dynamicClient := dynamic.NewForConfigOrDie(sm.restCfg)
	applicationLister := application.NewLister(ctx, dynamicClient)

	if sm.envCfg.EnableNewCRDVersion {
		jsCleaner := cleaner.NewJetStreamCleaner(sm.logger)
		jetStreamHandler := backendjetstreamv2.NewJetStream(sm.envCfg,
			sm.metricsCollector, jsCleaner, defaultSubsConfig, sm.logger)
		jetStreamReconciler := jetstreamv2.NewReconciler(
			ctx,
			client,
			jetStreamHandler,
			sm.logger,
			recorder,
			jsCleaner,
			sinkv2.NewValidator(ctx, client, recorder, sm.logger),
		)
		sm.backendv2 = jetStreamReconciler.Backend
		if err := jetStreamReconciler.SetupUnmanaged(sm.mgr); err != nil {
			return xerrors.Errorf("unable to setup the NATS subscription controller: %v", err)
		}
		sm.namedLogger().Info("Started v1alpha2 JetStream subscription manager")
	} else {
		jsCleaner := eventtype.NewCleaner(sm.envCfg.EventTypePrefix, applicationLister, sm.logger)
		jetStreamHandler := backendjetstream.NewJetStream(sm.envCfg, sm.metricsCollector, sm.logger)
		jetStreamReconciler := jetstream.NewReconciler(
			ctx,
			client,
			jetStreamHandler,
			sm.logger,
			recorder,
			jsCleaner,
			defaultSubsConfig,
			sink.NewValidator(ctx, client, recorder, sm.logger),
		)
		sm.backend = jetStreamReconciler.Backend
		if err := jetStreamReconciler.SetupUnmanaged(sm.mgr); err != nil {
			return xerrors.Errorf("unable to setup the NATS subscription controller: %v", err)
		}
		sm.namedLogger().Info("Started JetStream subscription manager")
	}
	return nil
}

func (sm *SubscriptionManager) Stop(runCleanup bool) error {
	sm.cancel()
	if !runCleanup {
		return nil
	}
	dynamicClient := dynamic.NewForConfigOrDie(sm.restCfg)
	if sm.envCfg.EnableNewCRDVersion {
		return cleanupv2(sm.backendv2, dynamicClient, sm.namedLogger())
	}
	return cleanup(sm.backend, dynamicClient, sm.namedLogger())
}

// clean removes all JetStream artifacts.
func cleanup(backend backendjetstream.Backend, dynamicClient dynamic.Interface, logger *zap.SugaredLogger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var ok bool
	var jsBackend *backendjetstream.JetStream
	if jsBackend, ok = backend.(*backendjetstream.JetStream); !ok {
		err := errors.New("converting backend handler to JetStream handler failed")
		return err
	}

	// fetch all subscriptions.
	subscriptionsUnstructured, err := dynamicClient.Resource(utils.SubscriptionGroupVersionResource()).Namespace(corev1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrapf(err, "list subscriptions failed")
	}

	subs, err := utils.ToSubscriptionList(subscriptionsUnstructured)
	if err != nil {
		return errors.Wrapf(err, "convert subscriptionList from unstructured list failed")
	}

	// clean all status.
	isCleanupSuccessful := true
	for _, v := range subs.Items {
		sub := v
		subKey := types.NamespacedName{Namespace: sub.Namespace, Name: sub.Name}
		log := logger.With("key", subKey.String())

		desiredSub := utils.ResetStatusToDefaults(sub)
		if err := utils.UpdateSubscriptionStatus(ctx, dynamicClient, desiredSub); err != nil {
			isCleanupSuccessful = false
			log.Errorw("Failed to update JetStream subscription status", "error", err)
		}

		// clean subscriptions from JetStream.
		if jsBackend != nil {
			if err := jsBackend.DeleteSubscription(&sub); err != nil {
				isCleanupSuccessful = false
				log.Errorw("Failed to delete JetStream subscription", "error", err)
			}
		}
	}

	logger.Debugw("Finished cleanup process", "success", isCleanupSuccessful)
	return nil
}

// clean removes all JetStream artifacts.
func cleanupv2(backend backendjetstreamv2.Backend, dynamicClient dynamic.Interface, logger *zap.SugaredLogger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var ok bool
	var jsBackend *backendjetstreamv2.JetStream
	if jsBackend, ok = backend.(*backendjetstreamv2.JetStream); !ok {
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
		if updateErr := backendutilsv2.UpdateSubscriptionStatus(ctx, dynamicClient, desiredSub); updateErr != nil {
			isCleanupSuccessful = false
			log.Errorw("Failed to update JetStream v2 subscription status", "error", err)
		}

		// clean subscriptions from JetStream.
		if jsBackend != nil {
			if delErr := jsBackend.DeleteSubscription(&sub); delErr != nil {
				isCleanupSuccessful = false
				log.Errorw("Failed to delete JetStream v2 subscription", "error", err)
			}
		}
	}

	logger.Debugw("Finished cleanup process", "success", isCleanupSuccessful)
	return nil
}

func (sm *SubscriptionManager) namedLogger() *zap.SugaredLogger {
	return sm.logger.WithContext().Named(subscriptionManagerName)
}
