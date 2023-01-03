package eventmesh_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"

	"github.com/avast/retry-go/v3"
	"github.com/go-logr/zapr"
	apigatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	eventmeshreconciler "github.com/kyma-project/kyma/components/eventing-controller/controllers/subscriptionv2/eventmesh"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	backendbeb "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/beb"
	backendeventmesh "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/eventmesh"
	sink "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/sink/v2"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/utils"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/constants"
	eventMeshtypes "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	reconcilertestingv1 "github.com/kyma-project/kyma/components/eventing-controller/testing"
	reconcilertesting "github.com/kyma-project/kyma/components/eventing-controller/testing/v2"
)

const (
	testEnvStartDelay           = time.Minute
	testEnvStartAttempts        = 10
	beforeSuiteTimeoutInSeconds = testEnvStartAttempts * 60
	subscriptionNamespacePrefix = "test-"
	bigPollingInterval          = 3 * time.Second
	bigTimeOut                  = 40 * time.Second
	smallTimeOut                = 5 * time.Second
	smallPollingInterval        = 1 * time.Second
	domain                      = "domain.com"
)

var (
	acceptableMethods = []string{http.MethodPost, http.MethodOptions}
	k8sCancelFn       context.CancelFunc
)

var _ = Describe("Subscription Reconciliation Tests", func() {
	var namespaceName string
	var testID = 0
	var ctx context.Context

	// enable me for debugging
	// SetDefaultEventuallyTimeout(time.Minute)
	// SetDefaultEventuallyPollingInterval(time.Second)

	BeforeEach(func() {
		namespaceName = fmt.Sprintf("%s%d", subscriptionNamespacePrefix, testID)
		// we need to reset the http requests which the mock captured
		eventMeshMock.Reset()

		// Context
		ctx = context.Background()
	})

	AfterEach(func() {
		// detailed request logs
		logf.Log.V(1).Info("eventMesh requests", "number", eventMeshMock.Requests.Len())

		i := 0

		eventMeshMock.Requests.ReadEach(
			func(req *http.Request, payload interface{}) {
				reqDescription := fmt.Sprintf("method: %q, url: %q, payload object: %+v", req.Method, req.RequestURI, payload)
				fmt.Printf("request[%d]: %s\n", i, reqDescription)
				i++
			})

		// print all subscriptions in the namespace for debugging purposes
		if err := printSubscriptions(namespaceName); err != nil {
			logf.Log.Error(err, "print subscriptions failed")
		}
		testID++
	})
})

// getSubscription fetches a subscription using the lookupKey and allows making assertions on it.
func getSubscription(ctx context.Context, subscription *eventingv1alpha2.Subscription) AsyncAssertion {
	return Eventually(func() *eventingv1alpha2.Subscription {
		lookupKey := types.NamespacedName{
			Namespace: subscription.Namespace,
			Name:      subscription.Name,
		}
		if err := k8sClient.Get(ctx, lookupKey, subscription); err != nil {
			log.Printf("fetch subscription %s failed: %v", lookupKey.String(), err)
			return &eventingv1alpha2.Subscription{}
		}
		log.Printf("[Subscription] name:%s ns:%s apiRule:%s", subscription.Name, subscription.Namespace, subscription.Status.Backend.APIRuleName)
		return subscription
	}, bigTimeOut, bigPollingInterval)
}

// ensureAPIRuleStatusUpdated updates the status fof the APIRule(mocking APIGateway controller).
func ensureAPIRuleStatusUpdatedWithStatusReady(ctx context.Context, apiRule *apigatewayv1beta1.APIRule) AsyncAssertion {
	By(fmt.Sprintf("Ensuring the APIRule %q is updated", apiRule.Name))

	return Eventually(func() error {

		lookupKey := types.NamespacedName{
			Namespace: apiRule.Namespace,
			Name:      apiRule.Name,
		}
		err := k8sClient.Get(ctx, lookupKey, apiRule)
		if err != nil {
			return err
		}
		newAPIRule := apiRule.DeepCopy()
		reconcilertesting.MarkReady(newAPIRule)
		err = k8sClient.Status().Update(ctx, newAPIRule)
		if err != nil {
			return err
		}
		return nil
	}, bigTimeOut, bigPollingInterval)
}

// ensureSubscriptionCreated creates a Subscription in the k8s cluster. If a custom namespace is used, it will be created as well.
func ensureSubscriptionCreated(ctx context.Context, subscription *eventingv1alpha2.Subscription) {
	By(fmt.Sprintf("Ensuring the test namespace %q is created", subscription.Namespace))
	if subscription.Namespace != "default " {
		// create testing namespace
		namespace := fixtureNamespace(subscription.Namespace)
		err := k8sClient.Create(ctx, namespace)
		if !k8serrors.IsAlreadyExists(err) {
			Expect(err).ShouldNot(HaveOccurred())
		}
	}

	By(fmt.Sprintf("Ensuring the subscription %q is created", subscription.Name))
	Expect(k8sClient.Create(ctx, subscription)).Should(Succeed())
}

// ensureSubscriberSvcCreated creates a Service in the k8s cluster. If a custom namespace is used, it will be created as well.
func ensureSubscriberSvcCreated(ctx context.Context, svc *corev1.Service) {
	By(fmt.Sprintf("Ensuring the test namespace %q is created", svc.Namespace))
	if svc.Namespace != "default " {
		// create testing namespace
		namespace := fixtureNamespace(svc.Namespace)
		err := k8sClient.Create(ctx, namespace)
		if !k8serrors.IsAlreadyExists(err) {
			Expect(err).ShouldNot(HaveOccurred())
		}
	}

	By(fmt.Sprintf("Ensuring the subscriber service %q is created", svc.Name))
	Expect(k8sClient.Create(ctx, svc)).Should(Succeed())
}

func fixtureNamespace(name string) *corev1.Namespace {
	namespace := corev1.Namespace{
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

// printSubscriptions prints all subscriptions in the given namespace.
func printSubscriptions(namespace string) error {
	// print subscription details
	ctx := context.TODO()
	subscriptionList := eventingv1alpha2.SubscriptionList{}
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

func getAPIRule(ctx context.Context, apiRule *apigatewayv1beta1.APIRule) AsyncAssertion {
	return Eventually(func() apigatewayv1beta1.APIRule {
		lookUpKey := types.NamespacedName{
			Namespace: apiRule.Namespace,
			Name:      apiRule.Name,
		}
		if err := k8sClient.Get(ctx, lookUpKey, apiRule); err != nil {
			log.Printf("fetch APIRule %s failed: %v", lookUpKey.String(), err)
			return apigatewayv1beta1.APIRule{}
		}
		return *apiRule
	}, bigTimeOut, bigPollingInterval)
}

func filterAPIRulesForASvc(apiRules *apigatewayv1beta1.APIRuleList, svc *corev1.Service) apigatewayv1beta1.APIRule {
	if len(apiRules.Items) == 1 && *apiRules.Items[0].Spec.Service.Name == svc.Name {
		return apiRules.Items[0]
	}
	return apigatewayv1beta1.APIRule{}
}

func getAPIRules(ctx context.Context, svc *corev1.Service) *apigatewayv1beta1.APIRuleList {
	labels := map[string]string{
		constants.ControllerServiceLabelKey:  svc.Name,
		constants.ControllerIdentityLabelKey: constants.ControllerIdentityLabelValue,
	}
	apiRules := &apigatewayv1beta1.APIRuleList{}
	err := k8sClient.List(ctx, apiRules, &client.ListOptions{
		LabelSelector: k8slabels.SelectorFromSet(labels),
		Namespace:     svc.Namespace,
	})
	Expect(err).Should(BeNil())
	return apiRules
}

func getAPIRuleForASvc(ctx context.Context, svc *corev1.Service) AsyncAssertion {
	return Eventually(func() apigatewayv1beta1.APIRule {
		apiRules := getAPIRules(ctx, svc)
		return filterAPIRulesForASvc(apiRules, svc)
	}, smallTimeOut, smallPollingInterval)
}

// //////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Test Suite setup ////////////////////////////////////////////////////////////////////////////////////////////////////
// //////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// These tests use Ginkgo (BDD-style Go controllertesting framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

// TODO: make configurable
// but how?
const (
	useExistingCluster       = false
	attachControlPlaneOutput = false
)

var (
	k8sClient     client.Client
	testEnv       *envtest.Environment
	eventMeshMock *reconcilertestingv1.BEBMock
	nameMapper    utils.NameMapper
	mock          *reconcilertestingv1.BEBMock
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	By("bootstrapping test environment")
	useExistingCluster := useExistingCluster
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("../../../", "config", "crd", "bases", "eventing.kyma-project.io_eventingbackends.yaml"),
			filepath.Join("../../../", "config", "crd", "basesv1alpha2"),
			filepath.Join("../../../", "config", "crd", "external"),
		},
		AttachControlPlaneOutput: attachControlPlaneOutput,
		UseExistingCluster:       &useExistingCluster,
	}

	var err error
	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	Expect(err).To(BeNil())
	logf.SetLogger(zapr.NewLogger(defaultLogger.WithContext().Desugar()))

	var cfg *rest.Config
	err = retry.Do(func() error {
		defer func() {
			if r := recover(); r != nil {
				log.Println("panic recovered:", r)
			}
		}()

		cfg, err = testEnv.Start()
		return err
	},
		retry.Delay(testEnvStartDelay),
		retry.DelayType(retry.FixedDelay),
		retry.Attempts(testEnvStartAttempts),
		retry.OnRetry(func(n uint, err error) {
			log.Printf("[%v] try failed to start testenv: %s", n, err)
			if stopErr := testEnv.Stop(); stopErr != nil {
				log.Printf("failed to stop testenv: %s", stopErr)
			}
		}),
	)

	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	err = eventingv1alpha2.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = apigatewayv1beta1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	// +kubebuilder:scaffold:scheme

	mock = startEventMeshMock()
	// client, err := client.New()
	// Source: https://book.kubebuilder.io/cronjob-tutorial/writing-tests.html
	syncPeriod := time.Second * 2
	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme.Scheme,
		SyncPeriod:         &syncPeriod,
		MetricsBindAddress: "localhost:9095",
	})
	Expect(err).ToNot(HaveOccurred())
	envConf := env.Config{
		BEBAPIURL:                mock.MessagingURL,
		ClientID:                 "foo-id",
		ClientSecret:             "foo-secret",
		TokenEndpoint:            mock.TokenURL,
		WebhookActivationTimeout: 0,
		WebhookTokenEndpoint:     "foo-token-endpoint",
		Domain:                   domain,
		EventTypePrefix:          reconcilertesting.EventMeshPrefix,
		BEBNamespace:             "/default/ns",
		Qos:                      string(eventMeshtypes.QosAtLeastOnce),
		EnableNewCRDVersion:      true,
	}

	credentials := &backendbeb.OAuth2ClientCredentials{
		ClientID:     "foo-client-id",
		ClientSecret: "foo-client-secret",
	}

	// prepare
	eventMeshCleaner := cleaner.NewEventMeshCleaner(defaultLogger)
	nameMapper = utils.NewBEBSubscriptionNameMapper(domain, backendbeb.MaxBEBSubscriptionNameLength)
	eventMeshHandler := backendeventmesh.NewEventMesh(credentials, nameMapper, defaultLogger)

	recorder := k8sManager.GetEventRecorderFor("eventing-controller")
	sinkValidator := sink.NewValidator(context.Background(), k8sManager.GetClient(), recorder)
	err = eventmeshreconciler.NewReconciler(context.Background(), k8sManager.GetClient(), defaultLogger,
		recorder, envConf, eventMeshCleaner, eventMeshHandler, credentials, nameMapper, sinkValidator).SetupUnmanaged(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		var ctx context.Context
		ctx, k8sCancelFn = context.WithCancel(ctrl.SetupSignalHandler())
		err = k8sManager.Start(ctx)
		Expect(err).ToNot(HaveOccurred())
	}()

	k8sClient = k8sManager.GetClient()
	Expect(k8sClient).ToNot(BeNil())

	close(done)
}, beforeSuiteTimeoutInSeconds)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	if k8sCancelFn != nil {
		k8sCancelFn()
	}
	err := testEnv.Stop()
	mock.Stop()
	Expect(err).ToNot(HaveOccurred())
})

// startEventMeshMock starts the EventMesh mock and configures the controller process to use it.
func startEventMeshMock() *reconcilertestingv1.BEBMock {
	By("Preparing EventMesh Mock")
	eventMeshMock = reconcilertestingv1.NewBEBMock()
	eventMeshMock.Start()
	return eventMeshMock
}

// countEventMeshRequests returns how many requests for a given subscription are sent for each HTTP method
//
//nolint:unparam
func countEventMeshRequests(subscriptionName, eventType string) (countGet, countPost, countDelete int) {
	countGet, countPost, countDelete = 0, 0, 0
	eventMeshMock.Requests.ReadEach(
		func(request *http.Request, payload interface{}) {
			switch method := request.Method; method {
			case http.MethodGet:
				if strings.Contains(request.URL.Path, subscriptionName) {
					countGet++
				}
			case http.MethodPost:
				if sub, ok := payload.(eventMeshtypes.Subscription); ok {
					if len(sub.Events) > 0 {
						for _, event := range sub.Events {
							if event.Type == eventType && sub.Name == subscriptionName {
								countPost++
							}
						}
					}
				}
			case http.MethodDelete:
				if strings.Contains(request.URL.Path, subscriptionName) {
					countDelete++
				}
			}
		})
	return countGet, countPost, countDelete
}
