package jetstream

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"testing"
	"time"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	reconcilertestingv2 "github.com/kyma-project/kyma/components/eventing-controller/controllers/subscriptionv2/reconcilertesting"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	cleanerv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/jetstreamv2"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/metrics"
	backendnats "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/nats"
	sinkv2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/sink/v2"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	v1 "github.com/kyma-project/kyma/components/eventing-controller/testing"
	v2 "github.com/kyma-project/kyma/components/eventing-controller/testing/v2"
	"github.com/onsi/gomega"
	gomegatypes "github.com/onsi/gomega/types"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

const (
	useExistingCluster       = false
	attachControlPlaneOutput = false
	emptyEventSource         = ""
	syncPeriod               = time.Second * 2
)

type jetStreamTestEnsemble struct {
	Reconciler       *Reconciler
	jetStreamBackend *jetstreamv2.JetStream
	JSStreamName     string
	*reconcilertestingv2.TestEnsemble
}

func setupTestEnsemble(t *testing.T) *jetStreamTestEnsemble {
	ctx := context.Background()
	g := gomega.NewGomegaWithT(t)
	useExistingCluster := useExistingCluster
	natsPort, err := v2.GetFreePort()
	require.NoError(t, err)
	natsServer := v1.StartDefaultJetStreamServer(natsPort)
	log.Printf("NATS server with JetStream started %v", natsServer.ClientURL())

	ens := &reconcilertestingv2.TestEnsemble{
		Ctx: ctx,
		G:   g,
		T:   t,
		DefaultSubscriptionConfig: env.DefaultSubscriptionConfig{
			MaxInFlightMessages: 1,
		},
		NatsPort:   natsPort,
		NatsServer: natsServer,
		TestEnv: &envtest.Environment{
			CRDDirectoryPaths: []string{
				filepath.Join("../../../", "config", "crd", "bases", "eventing.kyma-project.io_eventingbackends.yaml"),
				filepath.Join("../../../", "config", "crd", "basesv1alpha2"),
				filepath.Join("../../../", "config", "crd", "external"),
			},
			AttachControlPlaneOutput: attachControlPlaneOutput,
			UseExistingCluster:       &useExistingCluster,
			WebhookInstallOptions: envtest.WebhookInstallOptions{
				Paths: []string{filepath.Join("../../../", "config", "webhook")},
			},
		},
	}

	jsTestEnsemble := &jetStreamTestEnsemble{
		TestEnsemble: ens,
		JSStreamName: fmt.Sprintf("%s%d", v2.JSStreamName, natsPort),
	}

	reconcilertestingv2.StartTestEnv(ens)
	startReconciler(jsTestEnsemble)
	reconcilertestingv2.StartSubscriberSvc(ens)

	return jsTestEnsemble
}

func startReconciler(ens *jetStreamTestEnsemble) *jetStreamTestEnsemble {
	ctx, cancel := context.WithCancel(context.Background())
	ens.Cancel = cancel

	err := eventingv1alpha2.AddToScheme(scheme.Scheme)
	require.NoError(ens.T, err)

	var metricsPort int
	metricsPort, err = v2.GetFreePort()
	require.NoError(ens.T, err)

	syncPeriod := syncPeriod
	webhookInstallOptions := &ens.TestEnv.WebhookInstallOptions
	k8sManager, err := ctrl.NewManager(ens.Cfg, ctrl.Options{
		Scheme:             scheme.Scheme,
		SyncPeriod:         &syncPeriod,
		Host:               webhookInstallOptions.LocalServingHost,
		Port:               webhookInstallOptions.LocalServingPort,
		CertDir:            webhookInstallOptions.LocalServingCertDir,
		MetricsBindAddress: fmt.Sprintf("localhost:%v", metricsPort),
	})
	require.NoError(ens.T, err)

	envConf := backendnats.Config{
		URL:                     ens.NatsServer.ClientURL(),
		MaxReconnects:           reconcilertestingv2.MaxReconnects,
		ReconnectWait:           time.Second,
		EventTypePrefix:         v2.EventTypePrefix,
		JSStreamDiscardPolicy:   jetstreamv2.DiscardPolicyNew,
		JSStreamName:            ens.JSStreamName,
		JSSubjectPrefix:         ens.JSStreamName,
		JSStreamStorageType:     jetstreamv2.StorageTypeMemory,
		JSStreamMaxBytes:        "-1",
		JSStreamMaxMessages:     -1,
		JSStreamRetentionPolicy: "interest",
		EnableNewCRDVersion:     true,
	}

	// init the metrics collector
	metricsCollector := metrics.NewCollector()

	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	require.NoError(ens.T, err)

	defaultSubConfig := env.DefaultSubscriptionConfig{}
	cleaner := cleanerv1alpha2.NewJetStreamCleaner(defaultLogger)
	jetStreamHandler := jetstreamv2.NewJetStream(envConf, metricsCollector, cleaner, defaultSubConfig, defaultLogger)

	k8sClient := k8sManager.GetClient()
	recorder := k8sManager.GetEventRecorderFor("eventing-controller-jetstream")

	ens.Reconciler = NewReconciler(ctx,
		k8sClient,
		jetStreamHandler,
		defaultLogger,
		recorder,
		cleaner,
		sinkv2.NewValidator(ctx, k8sClient, recorder),
	)

	err = ens.Reconciler.SetupUnmanaged(k8sManager)
	require.NoError(ens.T, err)

	jsBackend, ok := ens.Reconciler.Backend.(*jetstreamv2.JetStream)
	if !ok {
		panic("cannot convert the Backend interface to Jetstreamv2")
	}
	ens.jetStreamBackend = jsBackend

	go func() {
		err = k8sManager.Start(ctx)
		require.NoError(ens.T, err)
	}()

	ens.K8sClient = k8sManager.GetClient()
	require.NotNil(ens.T, ens.K8sClient)

	err = reconcilertestingv2.StartAndWaitForWebhookServer(k8sManager, webhookInstallOptions)
	require.NoError(ens.T, err)

	return ens
}

// cleanupResources stop the testEnv and removes the stream from NATS test server.
func cleanupResources(ens *jetStreamTestEnsemble) {
	reconcilertestingv2.StopTestEnv(ens.TestEnsemble)

	jsCtx := ens.Reconciler.Backend.GetJetStreamContext()
	require.NoError(ens.T, jsCtx.DeleteStream(ens.JSStreamName))

	v1.ShutDownNATSServer(ens.NatsServer)
}

func testSubscriptionOnNATS(ens *jetStreamTestEnsemble, subscription *eventingv1alpha2.Subscription,
	subject string, expectations ...gomegatypes.GomegaMatcher) {
	description := "Failed to match nats subscriptions"
	getSubscriptionFromJetStream(ens,
		subscription,
		ens.jetStreamBackend.GetJetStreamSubject(
			subscription.Spec.Source,
			subject,
			subscription.Spec.TypeMatching),
	).Should(gomega.And(expectations...), description)
}

// testSubscriptionDeletion deletes the subscription and ensures it is not found anymore on the apiserver.
func testSubscriptionDeletion(ens *jetStreamTestEnsemble, subscription *eventingv1alpha2.Subscription) {
	g := ens.G
	g.Eventually(func() error {
		return ens.K8sClient.Delete(ens.Ctx, subscription)
	}, reconcilertestingv2.SmallTimeout, reconcilertestingv2.SmallPollingInterval).ShouldNot(gomega.HaveOccurred())
	reconcilertestingv2.IsSubscriptionDeletedOnK8s(ens.TestEnsemble, subscription).
		Should(v2.HaveNotFoundSubscription(), "Failed to delete subscription")
}

// ensureNATSSubscriptionIsDeleted ensures that the NATS subscription is not found anymore.
// This ensures the controller did delete it correctly then the Subscription was deleted.
func ensureNATSSubscriptionIsDeleted(ens *jetStreamTestEnsemble,
	subscription *eventingv1alpha2.Subscription, subject string) {
	getSubscriptionFromJetStream(ens, subscription, subject).
		ShouldNot(reconcilertestingv2.BeExistingSubscription(), "Failed to delete NATS subscription")
}

// getSubscriptionFromJetStream returns a NATS subscription for a given subscription and subject.
// NOTE: We need to give the controller enough time to react on the changes.
// Otherwise, the returned NATS subscription could have the wrong state.
// For this reason Eventually is used here.
func getSubscriptionFromJetStream(ens *jetStreamTestEnsemble, subscription *eventingv1alpha2.Subscription,
	subject string) gomega.AsyncAssertion {
	g := ens.G

	return g.Eventually(func() jetstreamv2.Subscriber {
		subscriptions := ens.jetStreamBackend.GetNATSSubscriptions()
		subscriptionSubject := jetstreamv2.NewSubscriptionSubjectIdentifier(subscription, subject)
		for key, sub := range subscriptions {
			if key.ConsumerName() == subscriptionSubject.ConsumerName() {
				return sub
			}
		}
		return nil
	}, reconcilertestingv2.SmallTimeout, reconcilertestingv2.SmallPollingInterval)
}
