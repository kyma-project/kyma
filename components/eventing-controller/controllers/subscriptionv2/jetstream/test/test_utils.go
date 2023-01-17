package test

import (
	"context"
	"fmt"
	"github.com/kyma-project/kyma/components/eventing-controller/controllers/subscriptionv2/jetstream"
	"github.com/pkg/errors"
	"log"
	"path/filepath"
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

var (
	k8sCancelFn    context.CancelFunc
	jsTestEnsemble *jetStreamTestEnsemble
)

type jetStreamTestEnsemble struct {
	Reconciler       *jetstream.Reconciler
	jetStreamBackend *jetstreamv2.JetStream
	JSStreamName     string
	*reconcilertestingv2.TestEnsemble
}

func setupSuite() error {
	ctx := context.Background()
	useExistingCluster := useExistingCluster

	natsPort, err := v2.GetFreePort()
	if err != nil {
		return err
	}
	natsServer := v1.StartDefaultJetStreamServer(natsPort)
	log.Printf("NATS server with JetStream started %v", natsServer.ClientURL())

	ens := &reconcilertestingv2.TestEnsemble{
		Ctx: ctx,
		DefaultSubscriptionConfig: env.DefaultSubscriptionConfig{
			MaxInFlightMessages: 1,
		},
		NatsPort:   natsPort,
		NatsServer: natsServer,
		TestEnv: &envtest.Environment{
			CRDDirectoryPaths: []string{
				filepath.Join("../../../../", "config", "crd", "bases", "eventing.kyma-project.io_eventingbackends.yaml"),
				filepath.Join("../../../../", "config", "crd", "basesv1alpha2"),
				filepath.Join("../../../../", "config", "crd", "external"),
			},
			AttachControlPlaneOutput: attachControlPlaneOutput,
			UseExistingCluster:       &useExistingCluster,
			WebhookInstallOptions: envtest.WebhookInstallOptions{
				Paths: []string{filepath.Join("../../../../", "config", "webhook")},
			},
		},
	}

	jsTestEnsemble = &jetStreamTestEnsemble{
		TestEnsemble: ens,
		JSStreamName: fmt.Sprintf("%s%d", v2.JSStreamName, natsPort),
	}

	if err := reconcilertestingv2.StartTestEnv(ens); err != nil {
		return err
	}

	if err := startReconciler(); err != nil {
		return err
	}
	return reconcilertestingv2.StartSubscriberSvc(ens)
}

func startReconciler() error {
	ctx, cancel := context.WithCancel(context.Background())
	jsTestEnsemble.Cancel = cancel

	if err := eventingv1alpha2.AddToScheme(scheme.Scheme); err != nil {
		return err
	}

	var metricsPort int
	metricsPort, err := v2.GetFreePort()
	if err != nil {
		return err
	}

	syncPeriod := syncPeriod
	webhookInstallOptions := &jsTestEnsemble.TestEnv.WebhookInstallOptions
	k8sManager, err := ctrl.NewManager(jsTestEnsemble.Cfg, ctrl.Options{
		Scheme:             scheme.Scheme,
		SyncPeriod:         &syncPeriod,
		Host:               webhookInstallOptions.LocalServingHost,
		Port:               webhookInstallOptions.LocalServingPort,
		CertDir:            webhookInstallOptions.LocalServingCertDir,
		MetricsBindAddress: fmt.Sprintf("localhost:%v", metricsPort),
	})
	if err != nil {
		return err
	}

	envConf := backendnats.Config{
		URL:                     jsTestEnsemble.NatsServer.ClientURL(),
		MaxReconnects:           reconcilertestingv2.MaxReconnects,
		ReconnectWait:           time.Second,
		EventTypePrefix:         v2.EventTypePrefix,
		JSStreamDiscardPolicy:   jetstreamv2.DiscardPolicyNew,
		JSStreamName:            jsTestEnsemble.JSStreamName,
		JSSubjectPrefix:         jsTestEnsemble.JSStreamName,
		JSStreamStorageType:     jetstreamv2.StorageTypeMemory,
		JSStreamMaxBytes:        "-1",
		JSStreamMaxMessages:     -1,
		JSStreamRetentionPolicy: "interest",
		EnableNewCRDVersion:     true,
	}

	// init the metrics collector
	metricsCollector := metrics.NewCollector()

	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	if err != nil {
		return err
	}

	defaultSubConfig := env.DefaultSubscriptionConfig{}
	cleaner := cleanerv1alpha2.NewJetStreamCleaner(defaultLogger)
	jetStreamHandler := jetstreamv2.NewJetStream(envConf, metricsCollector, cleaner, defaultSubConfig, defaultLogger)

	k8sClient := k8sManager.GetClient()
	recorder := k8sManager.GetEventRecorderFor("eventing-controller-jetstream")

	jsTestEnsemble.Reconciler = jetstream.NewReconciler(ctx,
		k8sClient,
		jetStreamHandler,
		defaultLogger,
		recorder,
		cleaner,
		sinkv2.NewValidator(ctx, k8sClient, recorder),
	)

	if err := jsTestEnsemble.Reconciler.SetupUnmanaged(k8sManager); err != nil {
		return err
	}

	jsBackend, ok := jsTestEnsemble.Reconciler.Backend.(*jetstreamv2.JetStream)
	if !ok {
		return errors.New("cannot convert the Backend interface to Jetstreamv2")
	}
	jsTestEnsemble.jetStreamBackend = jsBackend

	go func() {
		if err := k8sManager.Start(ctx); err != nil {
			panic(err)
		}
	}()

	jsTestEnsemble.K8sClient = k8sManager.GetClient()
	if jsTestEnsemble.K8sClient == nil {
		return errors.New("K8sClient cannot be nil")
	}

	if err := reconcilertestingv2.StartAndWaitForWebhookServer(k8sManager, webhookInstallOptions); err != nil {
		return err
	}

	return nil
}

func tearDownSuite() error {
	if k8sCancelFn != nil {
		k8sCancelFn()
	}
	if err := cleanupResources(); err != nil {
		return err
	}
	return nil
}

// cleanupResources stop the testEnv and removes the stream from NATS test server.
func cleanupResources() error {
	reconcilertestingv2.StopTestEnv(jsTestEnsemble.TestEnsemble)

	jsCtx := jsTestEnsemble.Reconciler.Backend.GetJetStreamContext()
	if err := jsCtx.DeleteStream(jsTestEnsemble.JSStreamName); err != nil {
		return err
	}

	v1.ShutDownNATSServer(jsTestEnsemble.NatsServer)
	return nil
}

func testSubscriptionOnNATS(g *gomega.GomegaWithT, subscription *eventingv1alpha2.Subscription,
	subject string, expectations ...gomegatypes.GomegaMatcher) {
	description := "Failed to match nats subscriptions"
	getSubscriptionFromJetStream(g, subscription,
		jsTestEnsemble.jetStreamBackend.GetJetStreamSubject(
			subscription.Spec.Source,
			subject,
			subscription.Spec.TypeMatching),
	).Should(gomega.And(expectations...), description)
}

// testSubscriptionDeletion deletes the subscription and ensures it is not found anymore on the apiserver.
func testSubscriptionDeletion(g *gomega.GomegaWithT, subscription *eventingv1alpha2.Subscription) {
	g.Eventually(func() error {
		return jsTestEnsemble.K8sClient.Delete(jsTestEnsemble.Ctx, subscription)
	}, reconcilertestingv2.SmallTimeout, reconcilertestingv2.SmallPollingInterval).ShouldNot(gomega.HaveOccurred())
	reconcilertestingv2.IsSubscriptionDeletedOnK8s(g, jsTestEnsemble.TestEnsemble, subscription).
		Should(v2.HaveNotFoundSubscription(), "Failed to delete subscription")
}

// ensureNATSSubscriptionIsDeleted ensures that the NATS subscription is not found anymore.
// This ensures the controller did delete it correctly then the Subscription was deleted.
func ensureNATSSubscriptionIsDeleted(g *gomega.GomegaWithT, subscription *eventingv1alpha2.Subscription, subject string) {
	getSubscriptionFromJetStream(g, subscription, subject).
		ShouldNot(reconcilertestingv2.BeExistingSubscription(), "Failed to delete NATS subscription")
}

// getSubscriptionFromJetStream returns a NATS subscription for a given subscription and subject.
// NOTE: We need to give the controller enough time to react on the changes.
// Otherwise, the returned NATS subscription could have the wrong state.
// For this reason Eventually is used here.
func getSubscriptionFromJetStream(g *gomega.GomegaWithT, subscription *eventingv1alpha2.Subscription,
	subject string) gomega.AsyncAssertion {

	return g.Eventually(func() jetstreamv2.Subscriber {
		subscriptions := jsTestEnsemble.jetStreamBackend.GetNATSSubscriptions()
		subscriptionSubject := jetstreamv2.NewSubscriptionSubjectIdentifier(subscription, subject)
		for key, sub := range subscriptions {
			if key.ConsumerName() == subscriptionSubject.ConsumerName() {
				return sub
			}
		}
		return nil
	}, reconcilertestingv2.SmallTimeout, reconcilertestingv2.SmallPollingInterval)
}
