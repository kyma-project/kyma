package nats_test

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"testing"
	"time"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/controllers/events"
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

type testEnsemble struct {
	testID                    int
	cfg                       *rest.Config
	k8sClient                 client.Client
	testEnv                   *envtest.Environment
	natsServer                *natsserver.Server
	reconciler                *natsreconciler.Reconciler
	natsBackend               *handlers.Nats
	defaultSubscriptionConfig env.DefaultSubscriptionConfig
	subscriberSvc             *v1.Service
	cancel                    context.CancelFunc
}

//var (
//	testID      int
//	natsURL     string
//	cfg         *rest.Config
//	k8sClient   client.Client
//	testEnv     *envtest.Environment
//	natsServer  *natsserver.Server
//	reconciler  *natsreconciler.Reconciler
//	natsBackend *handlers.Nats
//	cancel      context.CancelFunc
//
//	defaultSubsConfig = env.DefaultSubscriptionConfig{MaxInFlightMessages: 1, DispatcherRetryPeriod: time.Second, DispatcherMaxRetries: 1}
//)

func TestCreateSubscription(t *testing.T) {
	ctx := context.Background()
	g := gomega.NewGomegaWithT(t)

	ens := setupTestEnsemble(ctx, reconcilertesting.EventTypePrefix, g)
	defer ens.cancel()

	var testCase = []struct {
		name                   string
		subscriptionOpts       []reconcilertesting.SubscriptionOpt
		wantedK8sSubscription  []gomegatypes.GomegaMatcher
		wantedEvents           []v1.Event
		wantedNATSSubscription []gomegatypes.GomegaMatcher
		shouldTestDeletion     bool // a bool in golang is false by default
	}{
		{
			name: "empty event type",
			subscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(reconcilertesting.EventSource, ""),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURLFromSvc(ens.subscriberSvc),
			},
			wantedK8sSubscription: []gomegatypes.GomegaMatcher{
				reconcilertesting.HaveCondition(
					eventingv1alpha1.MakeCondition(
						eventingv1alpha1.ConditionSubscriptionActive,
						eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
						v1.ConditionFalse, nats.ErrBadSubject.Error(),
					),
				),
			},
			wantedEvents:           nil,
			wantedNATSSubscription: nil,
		},

		{
			name: "invalid sink; misses 'http' and 'https'",
			subscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(reconcilertesting.EventSource, reconcilertesting.OrderCreatedEventTypeNotClean),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURL("invalid"),
			},
			wantedK8sSubscription: []gomegatypes.GomegaMatcher{
				reconcilertesting.HaveCondition(conditionInvalidSink("sink URL scheme should be 'http' or 'https'")),
			},
			wantedEvents: []v1.Event{
				eventInvalidSink("Sink URL scheme should be HTTP or HTTPS: invalid"),
			},
			wantedNATSSubscription: nil,
		},

		{
			name: "empty protocol, protocol setting, dialect",
			subscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter("", reconcilertesting.OrderCreatedEventTypeNotClean),
				reconcilertesting.WithSinkURLFromSvc(ens.subscriberSvc),
			},
			wantedK8sSubscription: []gomegatypes.GomegaMatcher{
				reconcilertesting.HaveCondition(conditionValidSubscription("")),
			},
			wantedEvents: nil,
			wantedNATSSubscription: []gomegatypes.GomegaMatcher{
				reconcilertesting.BeNotNil(),
				reconcilertesting.BeValid(),
				reconcilertesting.HaveSubject(reconcilertesting.OrderCreatedEventType),
			},
			shouldTestDeletion: true,
		},
	}

	for _, tc := range testCase {
		ens.testID++ //todo
		//todo log name of test

		//create subscription
		//todo create a function that does the following lines
		subscriptionName := fmt.Sprintf(subscriptionNameFormat, ens.testID)
		subscription := reconcilertesting.NewSubscription(subscriptionName, ens.subscriberSvc.Namespace, tc.subscriptionOpts...)
		createSubscriptionInK8s(ctx, ens, subscription, t)

		//test subscription against expectations on k8s cluster
		//todo replace with a function that does both of the following lines
		subExpectations := append(tc.wantedK8sSubscription, reconcilertesting.HaveSubscriptionName(subscriptionName))
		getSubscriptionOnK8S(ctx, ens, subscription, t).Should(gomega.And(subExpectations...))

		//events
		//todo put in single func
		for _, event := range tc.wantedEvents { //todo replace with gomega.And(events)
			getK8sEvents(ctx, ens, t).Should(reconcilertesting.HaveEvent(event))
		}

		//todo put in function
		getSubscriptionFromNATS(ens.natsBackend, subscriptionName, t).Should(gomega.And(tc.wantedNATSSubscription...))

		//todo put in function
		if tc.shouldTestDeletion {
			g.Expect(ens.k8sClient.Delete(ctx, subscription)).Should(gomega.BeNil())
			isSubscriptionDeleted(ctx, ens, subscription, t).Should(reconcilertesting.HaveNotFoundSubscription(true))
		}
	}
}

// isSubscriptionDeleted checks a subscription is deleted and allows making assertions on it
func isSubscriptionDeleted(ctx context.Context, ens *testEnsemble, subscription *eventingv1alpha1.Subscription, t *testing.T) gomega.AsyncAssertion {
	g := gomega.NewGomegaWithT(t)

	return g.Eventually(func() bool {
		lookupKey := types.NamespacedName{
			Namespace: subscription.Namespace,
			Name:      subscription.Name,
		}
		if err := ens.k8sClient.Get(ctx, lookupKey, subscription); err != nil {
			return k8serrors.IsNotFound(err)
		}
		return false
	}, smallTimeout, smallPollingInterval)
}

func conditionValidSubscription(msg string) eventingv1alpha1.Condition {
	return eventingv1alpha1.MakeCondition(
		eventingv1alpha1.ConditionSubscriptionActive,
		eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
		v1.ConditionTrue, msg)
}

func conditionInvalidSink(msg string) eventingv1alpha1.Condition { //todo should this be reconcilertesting
	return eventingv1alpha1.MakeCondition(
		eventingv1alpha1.ConditionSubscriptionActive,
		eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
		v1.ConditionFalse, msg)
}

func eventInvalidSink(msg string) v1.Event { // todo should this be in reoncilertesting
	return v1.Event{
		Reason:  string(events.ReasonValidationFailed),
		Message: msg,
		Type:    v1.EventTypeWarning,
	}
}

func setupTestEnsemble(ctx context.Context, eventTypePrefix string, g *gomega.GomegaWithT) *testEnsemble {
	useExistingCluster := useExistingCluster
	ens := &testEnsemble{
		defaultSubscriptionConfig: env.DefaultSubscriptionConfig{
			MaxInFlightMessages:   1,
			DispatcherRetryPeriod: time.Second,
			DispatcherMaxRetries:  1,
		},
		natsServer: startNATS(natsPort),
		testEnv: &envtest.Environment{
			CRDDirectoryPaths: []string{
				filepath.Join("../../../", "config", "crd", "bases"),
				filepath.Join("../../../", "config", "crd", "external"),
			},
			AttachControlPlaneOutput: attachControlPlaneOutput,
			UseExistingCluster:       &useExistingCluster,
		},
	}

	startTestEnv(ens, g)
	startReconciler(eventTypePrefix, ens, g)
	startSubscriberSvc(ctx, ens, g)
	return ens
}

func startTestEnv(ens *testEnsemble, g *gomega.GomegaWithT) {
	k8sCfg, err := ens.testEnv.Start()
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(k8sCfg).ToNot(gomega.BeNil())
	ens.cfg = k8sCfg
}

func startNATS(port int) *natsserver.Server {
	//todo any tests/ error handling needed here?
	natsServer := reconcilertesting.RunNatsServerOnPort(port)
	//clientURL := natsServer.ClientURL()
	log.Printf("NATS server started %v", natsServer.ClientURL())
	return natsServer
}

func startReconciler(eventTypePrefix string, ens *testEnsemble, g *gomega.GomegaWithT) *testEnsemble {
	ctx, cancel := context.WithCancel(context.Background())
	ens.cancel = cancel
	//logf.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(GinkgoWriter))) //todo

	err := eventingv1alpha1.AddToScheme(scheme.Scheme)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	syncPeriod := time.Second * 2
	k8sManager, err := ctrl.NewManager(ens.cfg, ctrl.Options{
		Scheme:             scheme.Scheme,
		SyncPeriod:         &syncPeriod,
		MetricsBindAddress: "localhost:7070",
	})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	envConf := env.NatsConfig{
		URL:             ens.natsServer.ClientURL(),
		MaxReconnects:   10,
		ReconnectWait:   time.Second,
		EventTypePrefix: eventTypePrefix,
	}

	// prepare application-lister
	app := applicationtest.NewApplication(reconcilertesting.ApplicationNameNotClean, nil)
	applicationLister := fake.NewApplicationListerOrDie(context.Background(), app)

	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	g.Expect(err).To(gomega.BeNil())

	ens.reconciler = natsreconciler.NewReconciler(
		ctx,
		k8sManager.GetClient(),
		applicationLister,
		defaultLogger,
		k8sManager.GetEventRecorderFor("eventing-controller-nats"),
		envConf,
		ens.defaultSubscriptionConfig,
	)

	err = ens.reconciler.SetupUnmanaged(k8sManager)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	ens.natsBackend = ens.reconciler.Backend.(*handlers.Nats)

	go func() { //todo is this func needed any longer?
		err = k8sManager.Start(ctx)
		g.Expect(err).ToNot(gomega.HaveOccurred())
	}()

	ens.k8sClient = k8sManager.GetClient()
	g.Expect(ens.k8sClient).ToNot(gomega.BeNil())

	return ens
}

func startSubscriberSvc(ctx context.Context, ens *testEnsemble, g *gomega.GomegaWithT) {
	ens.subscriberSvc = reconcilertesting.NewSubscriberSvc("test-subscriber", "test")
	createSubscriberSvcInK8s(ctx, ens, g)
}

func createSubscriberSvcInK8s(ctx context.Context, ens *testEnsemble, g *gomega.GomegaWithT) {
	//if the namespace is not "default" create it on the cluster
	if ens.subscriberSvc.Namespace != "default " {
		namespace := fixtureNamespace(ens.subscriberSvc.Namespace)
		if namespace.Name != "default" {
			err := ens.k8sClient.Create(ctx, namespace)
			if !k8serrors.IsAlreadyExists(err) {
				fmt.Println(err)
				g.Expect(err).ShouldNot(gomega.HaveOccurred())
			}
		}
	}

	err := ens.k8sClient.Create(ctx, ens.subscriberSvc)
	g.Expect(err).Should(gomega.BeNil())
}

func createSubscriptionInK8s(ctx context.Context, ens *testEnsemble, subscription *eventingv1alpha1.Subscription, t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	if subscription.Namespace != "default " {
		// create testing namespace
		namespace := fixtureNamespace(subscription.Namespace)
		if namespace.Name != "default" {
			err := ens.k8sClient.Create(ctx, namespace)
			if !k8serrors.IsAlreadyExists(err) {
				fmt.Println(err)
				g.Expect(err).ShouldNot(gomega.HaveOccurred())
			}
		}
	}

	// create subscription
	err := ens.k8sClient.Create(ctx, subscription)
	g.Expect(err).Should(gomega.BeNil())
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
func getSubscriptionOnK8S(ctx context.Context, ens *testEnsemble, subscription *eventingv1alpha1.Subscription, t *testing.T, intervals ...interface{}) gomega.AsyncAssertion {
	g := gomega.NewGomegaWithT(t) //todo can i just pass g from caller?

	if len(intervals) == 0 {
		intervals = []interface{}{smallTimeout, smallPollingInterval}
	}
	return g.Eventually(func() *eventingv1alpha1.Subscription {
		lookupKey := types.NamespacedName{
			Namespace: subscription.Namespace,
			Name:      subscription.Name,
		}
		if err := ens.k8sClient.Get(ctx, lookupKey, subscription); err != nil {
			return &eventingv1alpha1.Subscription{}
		}
		return subscription
	}, intervals...)
}

// getK8sEvents returns all kubernetes events for the given namespace.
// The result can be used in a gomega assertion.
func getK8sEvents(ctx context.Context, ens *testEnsemble, t *testing.T) gomega.AsyncAssertion {
	g := gomega.NewGomegaWithT(t)

	eventList := v1.EventList{}
	return g.Eventually(func() v1.EventList {
		err := ens.k8sClient.List(ctx, &eventList, client.InNamespace(ens.subscriberSvc.Namespace))
		if err != nil {
			return v1.EventList{}
		}
		return eventList
	}, smallTimeout, smallPollingInterval)
}

func getSubscriptionFromNATS(natsBackend *handlers.Nats, subscriptionName string, t *testing.T) gomega.Assertion {
	g := gomega.NewGomegaWithT(t)

	return g.Expect(func() *nats.Subscription {
		subscriptions := natsBackend.GetAllSubscriptions()
		for key, subscription := range subscriptions {
			//the key does NOT ONLY contain the subscription name
			if strings.Contains(key, subscriptionName) {
				return subscription
			}
		}
		return nil
	}()) // todo do we need func()*nats.Subscription{}()? can we remove func?
}
