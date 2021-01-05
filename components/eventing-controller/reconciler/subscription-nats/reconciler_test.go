package subscription_nats

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"testing"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/nats-io/nats-server/v2/server"
	natsserver "github.com/nats-io/nats-server/v2/server"
	natstestserver "github.com/nats-io/nats-server/v2/test"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	apigatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	reconcilertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	bigTimeOut           = 40 * time.Second
	bigPollingInterval   = 3 * time.Second
	smallTimeOut         = 5 * time.Second
	smallPollingInterval = 1 * time.Second
)

var _ = Describe("NATS Subscription Reconciliation Tests", func() {
	var namespaceName = "test"

	// enable me for debugging
	// SetDefaultEventuallyTimeout(time.Minute)
	// SetDefaultEventuallyPollingInterval(time.Second)

	BeforeEach(func() {
		//namespaceName = "test"
		// we need to reset the http requests which the mock captured
		//beb.Reset()
	})

	AfterEach(func() {
		// print all subscriptions in the namespace for debugging purposes
		if err := printSubscriptions(namespaceName); err != nil {
			logf.Log.Error(err, "error while printing subscriptions")
		}
	})

	When("Creating/deleting a Subscription", func() {
		It("Should create/delete a subscription in NATS", func() {
			ctx := context.Background()
			subscriptionName := "sub"

			// Create subscription
			givenSubscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName,
				reconcilertesting.WithFilterForNats, reconcilertesting.WithWebhookForNats)
			givenSubscription.Spec.Sink = "http://valid.sink"
			ensureSubscriptionCreated(givenSubscription, ctx)

			getSubscription(givenSubscription, ctx).Should(And(
				reconcilertesting.HaveSubscriptionName(subscriptionName),
				reconcilertesting.HaveCondition(eventingv1alpha1.MakeCondition(
					eventingv1alpha1.ConditionSubscriptionActive,
					eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
					v1.ConditionTrue, "")),
			))

			Expect(k8sClient.Delete(ctx, givenSubscription)).Should(BeNil())
			isSubscriptionDeleted(givenSubscription, ctx).Should(reconcilertesting.HaveNotFoundSubscription(true))
		})
	})

	When("Creating a Subscription with invalid sink", func() {
		It("Should mark the subscription as not ready", func() {
			ctx := context.Background()
			subscriptionName := "sub"

			// Create subscription
			givenSubscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName,
				reconcilertesting.WithFilterForNats, reconcilertesting.WithWebhookForNats)
			givenSubscription.Spec.Sink = "invalid"
			ensureSubscriptionCreated(givenSubscription, ctx)

			getSubscription(givenSubscription, ctx).Should(And(
				reconcilertesting.HaveSubscriptionName(subscriptionName),
				reconcilertesting.HaveCondition(eventingv1alpha1.MakeCondition(
					eventingv1alpha1.ConditionSubscriptionActive,
					eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
					v1.ConditionFalse, "parse \"invalid\": invalid URI for request")),
			))
		})
	})

	When("Creating a Subscription with invalid protocol", func() {
		It("Should mark the subscription as not ready", func() {
			ctx := context.Background()
			subscriptionName := "invalid-sub-protocol"

			// Create subscription
			givenSubscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName,
				reconcilertesting.WithFilterForNats, reconcilertesting.WithWebhookForNats)
			reconcilertesting.WithValidSink("foo", "bar", givenSubscription)
			givenSubscription.Spec.Protocol = "invalid"
			ensureSubscriptionCreated(givenSubscription, ctx)

			getSubscription(givenSubscription, ctx).Should(And(
				reconcilertesting.HaveSubscriptionName(subscriptionName),
				reconcilertesting.HaveCondition(eventingv1alpha1.MakeCondition(
					eventingv1alpha1.ConditionSubscriptionActive,
					eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
					v1.ConditionFalse, "invalid protocol: invalid")),
			))
		})
	})

	When("Creating a Subscription with empty event type", func() {
		It("Should mark the subscription as not ready", func() {
			ctx := context.Background()
			subscriptionName := "invalid-sub-event-type"

			// Create subscription
			givenSubscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName,
				reconcilertesting.WithEmptyEventTypeFilterForNats, reconcilertesting.WithWebhookForNats)
			reconcilertesting.WithValidSink("foo", "bar", givenSubscription)
			ensureSubscriptionCreated(givenSubscription, ctx)

			getSubscription(givenSubscription, ctx).Should(And(
				reconcilertesting.HaveSubscriptionName(subscriptionName),
				reconcilertesting.HaveCondition(eventingv1alpha1.MakeCondition(
					eventingv1alpha1.ConditionSubscriptionActive,
					eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
					v1.ConditionFalse, "nats: invalid subject")),
			))
		})
	})

	PWhen("Creating a Subscription and NATS is unavailable", func() {
		It("Should mark the subscription as not ready", func() {

		})
	})

})

func ensureSubscriptionCreated(subscription *eventingv1alpha1.Subscription, ctx context.Context) {

	By(fmt.Sprintf("Ensuring the test namespace %q is created", subscription.Namespace))
	if subscription.Namespace != "default " {
		// create testing namespace
		namespace := fixtureNamespace(subscription.Namespace)
		if namespace.Name != "default" {
			err := k8sClient.Create(ctx, namespace)
			if !k8serrors.IsAlreadyExists(err) {
				fmt.Println(err)
				Expect(err).ShouldNot(HaveOccurred())
			}
		}
	}

	By(fmt.Sprintf("Ensuring the subscription %q is created", subscription.Name))
	// create subscription
	err := k8sClient.Create(ctx, subscription)
	Expect(err).Should(BeNil())
}

func fixtureNamespace(name string) *v1.Namespace {
	namespace := v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	return &namespace
}

// getSubscription fetches a subscription using the lookupKey and allows to make assertions on it
func getSubscription(subscription *eventingv1alpha1.Subscription, ctx context.Context) AsyncAssertion {
	return Eventually(func() eventingv1alpha1.Subscription {
		lookupKey := types.NamespacedName{
			Namespace: subscription.Namespace,
			Name:      subscription.Name,
		}
		if err := k8sClient.Get(ctx, lookupKey, subscription); err != nil {
			log.Printf("failed to fetch subscription(%s): %v", lookupKey.String(), err)
			return eventingv1alpha1.Subscription{}
		}
		log.Printf("[Subscription] name:%s ns:%s status:%v", subscription.Name, subscription.Namespace,
			subscription.Status)
		return *subscription
	}, smallTimeOut, smallPollingInterval)
}

// isSubscriptionDeleted checks a subscription is deleted and allows to make assertions on it
func isSubscriptionDeleted(subscription *eventingv1alpha1.Subscription, ctx context.Context) AsyncAssertion {
	return Eventually(func() bool {
		lookupKey := types.NamespacedName{
			Namespace: subscription.Namespace,
			Name:      subscription.Name,
		}
		if err := k8sClient.Get(ctx, lookupKey, subscription); err != nil {
			log.Printf("failed to fetch subscription(%s): %v", lookupKey.String(), err)
			return k8serrors.IsNotFound(err)
		}
		log.Printf("[Subscription] name:%s ns:%s status:%v", subscription.Name, subscription.Namespace,
			subscription.Status)
		return false
	}, smallTimeOut, smallPollingInterval)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Test Suite setup ////////////////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// These tests use Ginkgo (BDD-style Go controllertesting framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

// TODO: make configurable
const (
	useExistingCluster       = false
	attachControlPlaneOutput = false
)

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment

//var beb *reconcilertesting.BebMock

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t, "NATS Controller Suite", []Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(GinkgoWriter)))

	By("bootstrapping test environment")
	useExistingCluster := useExistingCluster

	s := RunDefaultServer()
	log.Printf("started test Nats server: %v", s.ClientURL())

	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("../../", "config", "crd", "bases"),
			filepath.Join("../../", "config", "crd", "external"),
		},
		AttachControlPlaneOutput: attachControlPlaneOutput,
		UseExistingCluster:       &useExistingCluster,
	}

	var err error

	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	err = eventingv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = apigatewayv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	// +kubebuilder:scaffold:scheme

	syncPeriod := time.Second
	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme.Scheme,
		SyncPeriod:         &syncPeriod,
		MetricsBindAddress: ":7070",
	})
	Expect(err).ToNot(HaveOccurred())
	envConf := env.NatsConfig{
		Url:           nats.DefaultURL,
		MaxReconnects: 10,
		ReconnectWait: time.Second,
	}
	err = NewReconciler(
		k8sManager.GetClient(),
		k8sManager.GetCache(),
		ctrl.Log.WithName("nats-reconciler").WithName("Subscription"),
		k8sManager.GetEventRecorderFor("eventing-controller-nats"),
		envConf,
	).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(ctrl.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred())
	}()

	k8sClient = k8sManager.GetClient()
	Expect(k8sClient).ToNot(BeNil())

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})

// printSubscriptions prints all subscriptions in the given namespace
func printSubscriptions(namespace string) error {
	// print subscription details
	ctx := context.TODO()
	subscriptionList := eventingv1alpha1.SubscriptionList{}
	if err := k8sClient.List(ctx, &subscriptionList, client.InNamespace(namespace)); err != nil {
		logf.Log.V(1).Info("error while getting subscription list", "error", err)
		return err
	}
	subscriptions := make([]string, 0)
	for _, sub := range subscriptionList.Items {
		subscriptions = append(subscriptions, sub.Name)
	}
	log.Printf("subscriptions: %v", subscriptions)
	return nil
}

// NATS tests helpers
func NewDefaultConnection() *nats.Conn {
	return NewConnection(nats.DefaultPort)
}

// NewConnection forms connection on a given port.
func NewConnection(port int) *nats.Conn {
	url := fmt.Sprintf("nats://127.0.0.1:%d", port)
	nc, err := nats.Connect(url)
	if err != nil {
		log.Fatal("Failed to create default connection")
	}
	if nc.Status() != nats.CONNECTED {
		log.Fatal("Failed to create default connection")
	}
	log.Printf("Connection to Nats status: %v", nc.Status())
	return nc
}

// RunDefaultServer will run a server on the default port.
func RunDefaultServer() *server.Server {
	return RunServerOnPort(nats.DefaultPort)
}

// RunServerOnPort will run a server on the given port.
func RunServerOnPort(port int) *server.Server {
	opts := natstestserver.DefaultTestOptions
	opts.Port = port
	//opts.Cluster.Name = "testing"
	return RunServerWithOptions(opts)
}

// RunServerWithOptions will run a server with the given options.
func RunServerWithOptions(opts natsserver.Options) *server.Server {
	return natstestserver.RunServer(&opts)
}
