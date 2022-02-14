package nats_test

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"testing"
	"time"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application/applicationtest"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application/fake"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers"
	reconcilertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
	"github.com/nats-io/nats.go"
	"github.com/onsi/gomega"
	gomegatypes "github.com/onsi/gomega/types"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	//logf "sigs.k8s.io/controller-runtime/pkg/log"
	//"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"k8s.io/client-go/kubernetes/scheme"

	natsreconciler "github.com/kyma-project/kyma/components/eventing-controller/controllers/subscription/nats"
	natsserver "github.com/nats-io/nats-server/v2/server"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

const (
	natsPort = 4221

	smallTimeout         = 10 * time.Second
	smallPollingInterval = 1 * time.Second

	timeout         = 60 * time.Second
	pollingInterval = 5 * time.Second

	namespaceName          = "test"
	subscriptionNameFormat = "nats-sub-%d"
	subscriberNameFormat   = "subscriber-%d"
)

const (
	useExistingCluster       = false
	attachControlPlaneOutput = false
)

var (
	testID      int
	natsURL     string
	cfg         *rest.Config
	k8sClient   client.Client
	testEnv     *envtest.Environment
	natsServer  *natsserver.Server
	reconciler  *natsreconciler.Reconciler
	natsBackend *handlers.Nats
	cancel      context.CancelFunc

	defaultSubsConfig = env.DefaultSubscriptionConfig{MaxInFlightMessages: 1, DispatcherRetryPeriod: time.Second, DispatcherMaxRetries: 1}
)

func TestCreateSubscription(t *testing.T) {
	ctx := context.Background()
	var id int

	cancel = setupTestEnvironment(reconcilertesting.EventTypePrefix, t)
	defer cancel()

	subscriberSvc := newSubscriberSvc(ctx, t)

	var testCase = []struct {
		name                     string
		subscriptionOpts         []reconcilertesting.SubscriptionOpt
		subscriptionExpectations []gomegatypes.GomegaMatcher
	}{
		{
			name: "empty event type",
			subscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(reconcilertesting.EventSource, ""),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURLFromSvc(subscriberSvc),
			},
			subscriptionExpectations: []gomegatypes.GomegaMatcher{
				reconcilertesting.HaveCondition(
					eventingv1alpha1.MakeCondition(
						eventingv1alpha1.ConditionSubscriptionActive,
						eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
						v1.ConditionFalse, nats.ErrBadSubject.Error(),
					),
				),
			},
		},
	}

	for _, tc := range testCase {
		//create subscription
		//todo create a function that does the following lines
		subscriptionName := fmt.Sprintf(subscriptionNameFormat, id)
		subscription := reconcilertesting.NewSubscription(subscriptionName, subscriberSvc.Namespace, tc.subscriptionOpts...)
		createSubscriptionInK8s(ctx, subscription, t)

		//test subscription against expectations
		//todo replace with a function that does both of the following lines
		subExpectations := append(tc.subscriptionExpectations, reconcilertesting.HaveSubscriptionName(subscriptionName))
		getSubscription(ctx, subscription).Should(gomega.And(subExpectations...))

		id++
	}
}

func setupTestEnvironment(eventTypePrefix string, t *testing.T) context.CancelFunc {
	natsServer, natsURL = startNATS(natsPort)
	testEnv = startTestEnv(t)
	cancel := startReconciler(eventTypePrefix, natsURL, t)
	return cancel
}

func startTestEnv(t *testing.T) *envtest.Environment {
	g := gomega.NewGomegaWithT(t)

	useExistingCluster := useExistingCluster
	testEnv := &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("../../../", "config", "crd", "bases"),
			filepath.Join("../../../", "config", "crd", "external"),
		},
		AttachControlPlaneOutput: attachControlPlaneOutput,
		UseExistingCluster:       &useExistingCluster,
	}

	var err error
	cfg, err = testEnv.Start()
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(cfg).ToNot(gomega.BeNil())

	return testEnv
}

func startNATS(port int) (*natsserver.Server, string) {
	//todo any tests/ error handling needed here?
	natsServer := reconcilertesting.RunNatsServerOnPort(port)
	clientURL := natsServer.ClientURL()
	log.Printf("NATS server started %v", clientURL)
	return natsServer, clientURL
}

func startReconciler(eventTypePrefix string, natsURL string, t *testing.T) context.CancelFunc {
	g := gomega.NewGomegaWithT(t)
	ctx, cancel := context.WithCancel(context.Background())
	//logf.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(GinkgoWriter)))

	err := eventingv1alpha1.AddToScheme(scheme.Scheme)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	syncPeriod := time.Second * 2
	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme.Scheme,
		SyncPeriod:         &syncPeriod,
		MetricsBindAddress: "localhost:7070",
	})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	envConf := env.NatsConfig{
		URL:             natsURL,
		MaxReconnects:   10,
		ReconnectWait:   time.Second,
		EventTypePrefix: eventTypePrefix,
	}

	// prepare application-lister
	app := applicationtest.NewApplication(reconcilertesting.ApplicationNameNotClean, nil)
	applicationLister := fake.NewApplicationListerOrDie(context.Background(), app)

	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	g.Expect(err).To(gomega.BeNil())

	reconciler = natsreconciler.NewReconciler(
		ctx,
		k8sManager.GetClient(),
		applicationLister,
		defaultLogger,
		k8sManager.GetEventRecorderFor("eventing-controller-nats"),
		envConf,
		defaultSubsConfig,
	)

	err = reconciler.SetupUnmanaged(k8sManager)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	natsBackend = reconciler.Backend.(*handlers.Nats)

	go func() { //todo is this func needed any longer
		err = k8sManager.Start(ctx)
		g.Expect(err).ToNot(gomega.HaveOccurred())
	}()

	k8sClient = k8sManager.GetClient()
	g.Expect(k8sClient).ToNot(gomega.BeNil())

	return cancel
}

func newSubscriberSvc(ctx context.Context, t *testing.T) *v1.Service {
	subscriberSvc := reconcilertesting.NewSubscriberSvc("test-subscriber", "test")
	createSubscriberSvcInK8s(ctx, subscriberSvc, t)
	return subscriberSvc
}

func createSubscriptionInK8s(ctx context.Context, subscription *eventingv1alpha1.Subscription, t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	if subscription.Namespace != "default " {
		// create testing namespace
		namespace := fixtureNamespace(subscription.Namespace)
		if namespace.Name != "default" {
			err := k8sClient.Create(ctx, namespace)
			if !k8serrors.IsAlreadyExists(err) {
				fmt.Println(err)
				g.Expect(err).ShouldNot(gomega.HaveOccurred())
			}
		}
	}

	// create subscription
	err := k8sClient.Create(ctx, subscription)
	g.Expect(err).Should(gomega.BeNil())
}

func createSubscriberSvcInK8s(ctx context.Context, svc *v1.Service, t *testing.T) *v1.Service {
	g := gomega.NewGomegaWithT(t)

	//if the namespace is not "default" create it on the cluster
	if svc.Namespace != "default " {
		namespace := fixtureNamespace(svc.Namespace)
		if namespace.Name != "default" {
			err := k8sClient.Create(ctx, namespace)
			if !k8serrors.IsAlreadyExists(err) {
				fmt.Println(err)
				g.Expect(err).ShouldNot(gomega.HaveOccurred())
			}
		}
	}

	err := k8sClient.Create(ctx, svc)
	g.Expect(err).Should(gomega.BeNil())
	return svc
}

func fixtureNamespace(name string) *v1.Namespace {
	namespace := v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "natsNamespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	return &namespace
}

// getSubscription fetches a subscription using the lookupKey and allows making assertions on it
func getSubscription(ctx context.Context, subscription *eventingv1alpha1.Subscription, intervals ...interface{}) gomega.AsyncAssertion {
	if len(intervals) == 0 {
		intervals = []interface{}{smallTimeout, smallPollingInterval}
	}
	return gomega.Eventually(func() *eventingv1alpha1.Subscription {
		lookupKey := types.NamespacedName{
			Namespace: subscription.Namespace,
			Name:      subscription.Name,
		}
		if err := k8sClient.Get(ctx, lookupKey, subscription); err != nil {
			return &eventingv1alpha1.Subscription{}
		}
		return subscription
	}, intervals...)
}
