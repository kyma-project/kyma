package jetstream_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/sink"

	ctrljetstream "github.com/kyma-project/kyma/components/eventing-controller/controllers/subscription/jetstream"

	"github.com/avast/retry-go/v3"
	"github.com/kyma-project/kyma/components/eventing-controller/controllers/events"
	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	cleanerv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/jetstream"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/metrics"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	eventingtesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
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
	SmallTimeout             = 20 * time.Second
	SmallPollingInterval     = 1 * time.Second
	subscriptionNameFormat   = "nats-sub-%d"
	retryAttempts            = 5
	MaxReconnects            = 10
)

//nolint:gochecknoglobals // these are required across the whole test package
var (
	k8sCancelFn    context.CancelFunc
	jsTestEnsemble *jetStreamTestEnsemble
)

type Ensemble struct {
	testID                    int
	Cfg                       *rest.Config
	K8sClient                 client.Client
	TestEnv                   *envtest.Environment
	NatsServer                *natsserver.Server
	NatsPort                  int
	DefaultSubscriptionConfig env.DefaultSubscriptionConfig
	SubscriberSvc             *corev1.Service
	Cancel                    context.CancelFunc
	Ctx                       context.Context
	G                         *gomega.GomegaWithT
	T                         *testing.T
}

type jetStreamTestEnsemble struct {
	Reconciler       *ctrljetstream.Reconciler
	jetStreamBackend *jetstream.JetStream
	JSStreamName     string
	*Ensemble
}

type Want struct {
	K8sSubscription []gomegatypes.GomegaMatcher
	K8sEvents       []corev1.Event
	// NatsSubscriptions holds gomega matchers for a NATS subscription per event-type.
	NatsSubscriptions map[string][]gomegatypes.GomegaMatcher
}

func setupSuite() error {
	ctx := context.Background()
	useExistingCluster := useExistingCluster

	natsPort, err := eventingtesting.GetFreePort()
	if err != nil {
		return err
	}
	natsServer := eventingtesting.StartDefaultJetStreamServer(natsPort)
	log.Printf("NATS server with JetStream started %v", natsServer.ClientURL())

	ens := &Ensemble{
		Ctx: ctx,
		DefaultSubscriptionConfig: env.DefaultSubscriptionConfig{
			MaxInFlightMessages: 1,
		},
		NatsPort:   natsPort,
		NatsServer: natsServer,
		TestEnv: &envtest.Environment{
			CRDDirectoryPaths: []string{
				filepath.Join("../../../", "config", "crd", "bases"),
				filepath.Join("../../../", "config", "crd", "external"),
			},
			AttachControlPlaneOutput: attachControlPlaneOutput,
			UseExistingCluster:       &useExistingCluster,
			WebhookInstallOptions: envtest.WebhookInstallOptions{
				Paths: []string{filepath.Join("../../../", "config", "webhook")},
			},
		},
	}

	jsTestEnsemble = &jetStreamTestEnsemble{
		Ensemble:     ens,
		JSStreamName: fmt.Sprintf("%s%d", eventingtesting.JSStreamName, natsPort),
	}

	if err := StartTestEnv(ens); err != nil {
		return err
	}

	if err := startReconciler(); err != nil {
		return err
	}
	return StartSubscriberSvc(ens)
}

func startReconciler() error {
	ctx, cancel := context.WithCancel(context.Background())
	jsTestEnsemble.Cancel = cancel

	if err := eventingv1alpha2.AddToScheme(scheme.Scheme); err != nil {
		return err
	}

	syncPeriod := syncPeriod
	webhookInstallOptions := &jsTestEnsemble.TestEnv.WebhookInstallOptions
	k8sManager, err := ctrl.NewManager(jsTestEnsemble.Cfg, ctrl.Options{
		Scheme:                 scheme.Scheme,
		SyncPeriod:             &syncPeriod,
		Host:                   webhookInstallOptions.LocalServingHost,
		Port:                   webhookInstallOptions.LocalServingPort,
		CertDir:                webhookInstallOptions.LocalServingCertDir,
		MetricsBindAddress:     "0", // disable
		HealthProbeBindAddress: "0", // disable
	})
	if err != nil {
		return err
	}

	envConf := env.NATSConfig{
		URL:                     jsTestEnsemble.NatsServer.ClientURL(),
		MaxReconnects:           MaxReconnects,
		ReconnectWait:           time.Second,
		EventTypePrefix:         eventingtesting.EventTypePrefix,
		JSStreamDiscardPolicy:   jetstream.DiscardPolicyNew,
		JSStreamName:            jsTestEnsemble.JSStreamName,
		JSSubjectPrefix:         jsTestEnsemble.JSStreamName,
		JSStreamStorageType:     jetstream.StorageTypeMemory,
		JSStreamMaxBytes:        "-1",
		JSStreamMaxMessages:     -1,
		JSStreamRetentionPolicy: "interest",
	}

	// init the metrics collector
	metricsCollector := metrics.NewCollector()

	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	if err != nil {
		return err
	}

	defaultSubConfig := env.DefaultSubscriptionConfig{}
	cleaner := cleanerv1alpha2.NewJetStreamCleaner(defaultLogger)
	jetStreamHandler := jetstream.NewJetStream(envConf, metricsCollector, cleaner, defaultSubConfig, defaultLogger)

	k8sClient := k8sManager.GetClient()
	recorder := k8sManager.GetEventRecorderFor("eventing-controller-jetstream")

	jsTestEnsemble.Reconciler = ctrljetstream.NewReconciler(ctx,
		k8sClient,
		jetStreamHandler,
		defaultLogger,
		recorder,
		cleaner,
		sink.NewValidator(ctx, k8sClient, recorder),
	)

	if err := jsTestEnsemble.Reconciler.SetupUnmanaged(k8sManager); err != nil {
		return err
	}

	jsBackend, ok := jsTestEnsemble.Reconciler.Backend.(*jetstream.JetStream)
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

	if err := StartAndWaitForWebhookServer(k8sManager, webhookInstallOptions); err != nil {
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
	StopTestEnv(jsTestEnsemble.Ensemble)

	jsCtx := jsTestEnsemble.Reconciler.Backend.GetJetStreamContext()
	if err := jsCtx.DeleteStream(jsTestEnsemble.JSStreamName); err != nil {
		return err
	}

	eventingtesting.ShutDownNATSServer(jsTestEnsemble.NatsServer)
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
	}, SmallTimeout, SmallPollingInterval).ShouldNot(gomega.HaveOccurred())
	IsSubscriptionDeletedOnK8s(g, jsTestEnsemble.Ensemble, subscription).
		Should(eventingtesting.HaveNotFoundSubscription(), "Failed to delete subscription")
}

// ensureNATSSubscriptionIsDeleted ensures that the NATS subscription is not found anymore.
// This ensures the controller did delete it correctly then the Subscription was deleted.
func ensureNATSSubscriptionIsDeleted(g *gomega.GomegaWithT,
	subscription *eventingv1alpha2.Subscription, subject string) {
	getSubscriptionFromJetStream(g, subscription, subject).
		ShouldNot(BeExistingSubscription(), "Failed to delete NATS subscription")
}

// getSubscriptionFromJetStream returns a NATS subscription for a given subscription and subject.
// NOTE: We need to give the controller enough time to react on the changes.
// Otherwise, the returned NATS subscription could have the wrong state.
// For this reason Eventually is used here.
func getSubscriptionFromJetStream(g *gomega.GomegaWithT,
	subscription *eventingv1alpha2.Subscription, subject string) gomega.AsyncAssertion {
	return g.Eventually(func() jetstream.Subscriber {
		subscriptions := jsTestEnsemble.jetStreamBackend.GetNATSSubscriptions()
		subscriptionSubject := jetstream.NewSubscriptionSubjectIdentifier(subscription, subject)
		for key, sub := range subscriptions {
			if key.ConsumerName() == subscriptionSubject.ConsumerName() {
				return sub
			}
		}
		return nil
	}, SmallTimeout, SmallPollingInterval)
}

// EventuallyUpdateSubscriptionOnK8s updates a given sub on kubernetes side.
// In order to be resilient and avoid a conflict, the update operation is retried until the update succeeds.
// To avoid a 409 conflict, the subscription CR data is read from the apiserver before a new update is performed.
// This conflict can happen if another entity such as the eventing-controller changed the sub in the meantime.
func EventuallyUpdateSubscriptionOnK8s(ctx context.Context, ens *Ensemble,
	sub *eventingv1alpha2.Subscription, updateFunc func(*eventingv1alpha2.Subscription) error) error {
	return doRetry(func() error {
		// get a fresh version of the Subscription
		lookupKey := types.NamespacedName{
			Namespace: sub.Namespace,
			Name:      sub.Name,
		}
		if err := ens.K8sClient.Get(ctx, lookupKey, sub); err != nil {
			return errors.Wrapf(err, "error while fetching subscription %s", lookupKey.String())
		}
		if err := updateFunc(sub); err != nil {
			return err
		}
		return nil
	}, "Failed to update the subscription on k8s")
}

func NewSubscription(ens *Ensemble,
	subscriptionOpts ...eventingtesting.SubscriptionOpt) *eventingv1alpha2.Subscription {
	subscriptionName := fmt.Sprintf(subscriptionNameFormat, ens.testID)
	ens.testID++
	subscription := eventingtesting.NewSubscription(subscriptionName, ens.SubscriberSvc.Namespace, subscriptionOpts...)
	return subscription
}

func CreateSubscription(t *testing.T, ens *Ensemble,
	subscriptionOpts ...eventingtesting.SubscriptionOpt) *eventingv1alpha2.Subscription {
	subscription := NewSubscription(ens, subscriptionOpts...)
	EnsureNamespaceCreatedForSub(t, ens, subscription)
	require.NoError(t, ensureSubscriptionCreated(ens, subscription))
	return subscription
}

func CheckSubscriptionOnK8s(g *gomega.WithT, ens *Ensemble, subscription *eventingv1alpha2.Subscription,
	expectations ...gomegatypes.GomegaMatcher) {
	description := "Failed to match the eventing subscription"
	expectations = append(expectations, eventingtesting.HaveSubscriptionName(subscription.Name))
	getSubscriptionOnK8S(g, ens, subscription).Should(gomega.And(expectations...), description)
}

func CheckEventsOnK8s(g *gomega.WithT, ens *Ensemble, expectations ...corev1.Event) {
	for _, event := range expectations {
		getK8sEvents(g, ens).Should(eventingtesting.HaveEvent(event), "Failed to match k8s events")
	}
}

func ValidSinkURL(ens *Ensemble, additions ...string) string {
	url := eventingtesting.ValidSinkURL(ens.SubscriberSvc.Namespace, ens.SubscriberSvc.Name)
	for _, a := range additions {
		url = fmt.Sprintf("%s%s", url, a)
	}
	return url
}

// IsSubscriptionDeletedOnK8s checks a subscription is deleted and allows making assertions on it.
func IsSubscriptionDeletedOnK8s(g *gomega.WithT, ens *Ensemble,
	subscription *eventingv1alpha2.Subscription) gomega.AsyncAssertion {
	return g.Eventually(func() bool {
		lookupKey := types.NamespacedName{
			Namespace: subscription.Namespace,
			Name:      subscription.Name,
		}
		if err := ens.K8sClient.Get(ens.Ctx, lookupKey, subscription); err != nil {
			return k8serrors.IsNotFound(err)
		}
		return false
	}, SmallTimeout, SmallPollingInterval)
}

func ConditionInvalidSink(msg string) eventingv1alpha2.Condition {
	return eventingv1alpha2.MakeCondition(
		eventingv1alpha2.ConditionSubscriptionActive,
		eventingv1alpha2.ConditionReasonNATSSubscriptionNotActive,
		corev1.ConditionFalse, msg)
}

func EventInvalidSink(msg string) corev1.Event {
	return corev1.Event{
		Reason:  string(events.ReasonValidationFailed),
		Message: msg,
		Type:    corev1.EventTypeWarning,
	}
}

func StartTestEnv(ens *Ensemble) error {
	var err error
	var k8sCfg *rest.Config

	err = retry.Do(func() error {
		defer func() {
			if r := recover(); r != nil {
				log.Println("panic recovered:", r)
			}
		}()

		k8sCfg, err = ens.TestEnv.Start()
		return err
	},
		retry.Delay(time.Minute),
		retry.DelayType(retry.FixedDelay),
		retry.Attempts(retryAttempts),
		retry.OnRetry(func(n uint, err error) {
			log.Printf("[%v] try failed to start testenv: %s", n, err)
			if stopErr := ens.TestEnv.Stop(); stopErr != nil {
				log.Printf("failed to stop testenv: %s", stopErr)
			}
		}),
	)

	if err != nil {
		return err
	}
	if k8sCfg == nil {
		return errors.New("k8s config cannot be nil")
	}
	ens.Cfg = k8sCfg

	return nil
}

func StopTestEnv(ens *Ensemble) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("panic recovered:", r)
		}
	}()

	if stopErr := ens.TestEnv.Stop(); stopErr != nil {
		log.Printf("failed to stop testenv: %s", stopErr)
	}
}

func StartSubscriberSvc(ens *Ensemble) error {
	ens.SubscriberSvc = eventingtesting.NewSubscriberSvc("test-subscriber", "test")
	return createSubscriberSvcInK8s(ens)
}

// createSubscriberSvcInK8s ensures the subscriber service in the k8s cluster is created successfully.
// The subscriber service is taken from the Ensemble struct and should not be nil.
// If the namespace of the subscriber service does not exist, it will be created.
func createSubscriberSvcInK8s(ens *Ensemble) error {
	// if the namespace is not "default" create it on the cluster
	if ens.SubscriberSvc.Namespace != "default " {
		namespace := fixtureNamespace(ens.SubscriberSvc.Namespace)
		err := doRetry(func() error {
			if err := ens.K8sClient.Create(ens.Ctx, namespace); !k8serrors.IsAlreadyExists(err) {
				return err
			}
			return nil
		}, "Failed to to create the namespace for the subscriber")
		if err != nil {
			return err
		}
	}

	return doRetry(func() error {
		return ens.K8sClient.Create(ens.Ctx, ens.SubscriberSvc)
	}, "Failed to create the subscriber service")
}

// EnsureNamespaceCreatedForSub creates the namespace for subscription if it does not exist.
func EnsureNamespaceCreatedForSub(t *testing.T, ens *Ensemble, subscription *eventingv1alpha2.Subscription) {
	// create subscription on cluster
	if subscription.Namespace != "default " {
		// create testing namespace
		namespace := fixtureNamespace(subscription.Namespace)
		err := ens.K8sClient.Create(ens.Ctx, namespace)
		if !k8serrors.IsAlreadyExists(err) {
			require.NoError(t, err)
		}
	}
}

// ensureSubscriptionCreated creates a Subscription on the K8s client of the testEnsemble. All the reconciliation
// happening will be reflected in the subscription.
func ensureSubscriptionCreated(ens *Ensemble, subscription *eventingv1alpha2.Subscription) error {
	// create subscription on cluster
	return doRetry(func() error {
		return ens.K8sClient.Create(ens.Ctx, subscription)
	}, "failed to create a subscription")
}

// EnsureK8sResourceNotCreated ensures that the obj creation in K8s fails.
func EnsureK8sResourceNotCreated(t *testing.T, ens *Ensemble, obj client.Object, err error) {
	require.Equal(t, ens.K8sClient.Create(ens.Ctx, obj), err)
}

func fixtureNamespace(name string) *corev1.Namespace {
	namespace := corev1.Namespace{
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

// getSubscriptionOnK8S fetches a subscription using the lookupKey and allows making assertions on it.
func getSubscriptionOnK8S(g *gomega.WithT, ens *Ensemble,
	subscription *eventingv1alpha2.Subscription) gomega.AsyncAssertion {
	return g.Eventually(func() *eventingv1alpha2.Subscription {
		lookupKey := types.NamespacedName{
			Namespace: subscription.Namespace,
			Name:      subscription.Name,
		}
		if err := ens.K8sClient.Get(ens.Ctx, lookupKey, subscription); err != nil {
			return &eventingv1alpha2.Subscription{}
		}
		return subscription
	}, SmallTimeout, SmallPollingInterval)
}

// getK8sEvents returns all kubernetes events for the given namespace.
// The result can be used in a gomega assertion.
func getK8sEvents(g *gomega.WithT, ens *Ensemble) gomega.AsyncAssertion {
	eventList := corev1.EventList{}
	return g.Eventually(func() corev1.EventList {
		err := ens.K8sClient.List(ens.Ctx, &eventList, client.InNamespace(ens.SubscriberSvc.Namespace))
		if err != nil {
			return corev1.EventList{}
		}
		return eventList
	}, SmallTimeout, SmallPollingInterval)
}

func NewUncleanEventType(ending string) string {
	return fmt.Sprintf("%s%s", eventingtesting.OrderCreatedEventTypeNotClean, ending)
}

func NewCleanEventType(ending string) string {
	return fmt.Sprintf("%s%s", eventingtesting.OrderCreatedEventType, ending)
}

func GenerateInvalidSubscriptionError(subName, errType string, path *field.Path) error {
	webhookErr := "admission webhook \"vsubscription.kb.io\" denied the request: "
	givenError := k8serrors.NewInvalid(
		eventingv1alpha2.GroupKind, subName,
		field.ErrorList{eventingv1alpha2.MakeInvalidFieldError(path, subName, errType)})
	givenError.ErrStatus.Message = webhookErr + givenError.ErrStatus.Message
	return givenError
}

func StartAndWaitForWebhookServer(k8sManager manager.Manager, webhookInstallOpts *envtest.WebhookInstallOptions) error {
	if err := (&eventingv1alpha2.Subscription{}).SetupWebhookWithManager(k8sManager); err != nil {
		return err
	}
	// wait for the webhook server to get ready
	dialer := &net.Dialer{Timeout: time.Second}
	addrPort := fmt.Sprintf("%s:%d", webhookInstallOpts.LocalServingHost, webhookInstallOpts.LocalServingPort)
	err := retry.Do(func() error {
		conn, connErr := tls.DialWithDialer(dialer, "tcp", addrPort, &tls.Config{InsecureSkipVerify: true})
		if connErr != nil {
			return connErr
		}
		return conn.Close()
	}, retry.Attempts(MaxReconnects))
	return err
}

func doRetry(function func() error, errString string) error {
	err := retry.Do(function,
		retry.Delay(time.Minute),
		retry.Attempts(retryAttempts),
		retry.OnRetry(func(n uint, err error) {
			log.Printf("[%v] %s: %s", n, errString, err)
		}),
	)
	if err != nil {
		return err
	}
	return nil
}
