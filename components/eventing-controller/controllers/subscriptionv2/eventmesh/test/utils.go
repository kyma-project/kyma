package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	testingv2 "github.com/kyma-project/kyma/components/eventing-controller/controllers/subscriptionv2/reconcilertesting"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/constants"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"

	"github.com/avast/retry-go/v3"
	"github.com/go-logr/zapr"
	apigatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	eventmeshreconciler "github.com/kyma-project/kyma/components/eventing-controller/controllers/subscriptionv2/eventmesh"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	backendbeb "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/beb"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"
	backendeventmesh "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/eventmesh"
	sink "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/sink/v2"
	backendutils "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/utils"
	eventMeshtypes "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	reconcilertestingv1 "github.com/kyma-project/kyma/components/eventing-controller/testing"
	reconcilertesting "github.com/kyma-project/kyma/components/eventing-controller/testing/v2"

	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/kyma-project/kyma/components/eventing-controller/utils"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type eventMeshTestEnsemble struct {
	k8sClient     client.Client
	testEnv       *envtest.Environment
	eventMeshMock *reconcilertestingv1.BEBMock
	nameMapper    backendutils.NameMapper
	envConfig     env.Config
}

const (
	useExistingCluster       = false
	attachControlPlaneOutput = false
	testEnvStartDelay        = time.Minute
	testEnvStartAttempts     = 10
	twoMinTimeOut            = 120 * time.Second
	bigPollingInterval       = 3 * time.Second
	bigTimeOut               = 40 * time.Second
	smallTimeOut             = 5 * time.Second
	smallPollingInterval     = 1 * time.Second
	domain                   = "domain.com"
	namespacePrefixLength    = 5
	syncPeriodSeconds        = 2
	eventMeshMockKeyPrefix   = "/messaging/events/subscriptions"
)

//nolint:gochecknoglobals // only used in tests
var (
	emTestEnsemble    *eventMeshTestEnsemble
	k8sCancelFn       context.CancelFunc
	acceptableMethods = []string{http.MethodPost, http.MethodOptions}
)

func setupSuite() error {
	emTestEnsemble = &eventMeshTestEnsemble{}

	// define logger
	var err error
	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	if err != nil {
		return err
	}
	logf.SetLogger(zapr.NewLogger(defaultLogger.WithContext().Desugar()))

	// setup test Env
	cfg, err := startTestEnv()
	if err != nil || cfg == nil {
		return err
	}

	// start event mesh mock
	emTestEnsemble.eventMeshMock = startNewEventMeshMock()

	// add schemes
	if err = eventingv1alpha2.AddToScheme(scheme.Scheme); err != nil {
		return err
	}

	if err = apigatewayv1beta1.AddToScheme(scheme.Scheme); err != nil {
		return err
	}
	// +kubebuilder:scaffold:scheme

	// get a free port for eventMesh manager
	metricsPort, err := reconcilertesting.GetFreePort()
	if err != nil {
		return err
	}

	// start eventMesh manager instance
	syncPeriod := syncPeriodSeconds * time.Second
	webhookInstallOptions := &emTestEnsemble.testEnv.WebhookInstallOptions
	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
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

	// setup nameMapper for EventMesh
	emTestEnsemble.nameMapper = backendutils.NewBEBSubscriptionNameMapper(domain, backendbeb.MaxBEBSubscriptionNameLength)

	// setup eventMesh reconciler
	recorder := k8sManager.GetEventRecorderFor("eventing-controller")
	sinkValidator := sink.NewValidator(context.Background(), k8sManager.GetClient(), recorder)
	credentials := &backendbeb.OAuth2ClientCredentials{
		ClientID:     "foo-client-id",
		ClientSecret: "foo-client-secret",
	}
	emTestEnsemble.envConfig = getEnvConfig()
	testReconciler := eventmeshreconciler.NewReconciler(
		context.Background(),
		k8sManager.GetClient(),
		defaultLogger,
		recorder,
		getEnvConfig(),
		cleaner.NewEventMeshCleaner(defaultLogger),
		backendeventmesh.NewEventMesh(credentials, emTestEnsemble.nameMapper, defaultLogger),
		credentials,
		emTestEnsemble.nameMapper,
		sinkValidator,
	)

	if err = testReconciler.SetupUnmanaged(k8sManager); err != nil {
		return err
	}

	// start k8s client
	go func() {
		var ctx context.Context
		ctx, k8sCancelFn = context.WithCancel(ctrl.SetupSignalHandler())
		err = k8sManager.Start(ctx)
		if err != nil {
			panic(err)
		}
	}()

	emTestEnsemble.k8sClient = k8sManager.GetClient()

	if err = testingv2.StartAndWaitForWebhookServer(k8sManager, webhookInstallOptions); err != nil {
		return err
	}

	return nil
}

func startTestEnv() (*rest.Config, error) {
	useExistingCluster := useExistingCluster
	emTestEnsemble.testEnv = &envtest.Environment{
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
	}

	var cfg *rest.Config
	err := retry.Do(func() error {
		defer func() {
			if r := recover(); r != nil {
				log.Println("panic recovered:", r)
			}
		}()

		cfgLocal, startErr := emTestEnsemble.testEnv.Start()
		cfg = cfgLocal
		return startErr
	},
		retry.Delay(testEnvStartDelay),
		retry.DelayType(retry.FixedDelay),
		retry.Attempts(testEnvStartAttempts),
		retry.OnRetry(func(n uint, err error) {
			log.Printf("[%v] try failed to start testenv: %s", n, err)
			if stopErr := emTestEnsemble.testEnv.Stop(); stopErr != nil {
				log.Printf("failed to stop testenv: %s", stopErr)
			}
		}),
	)
	return cfg, err
}

func getEnvConfig() env.Config {
	return env.Config{
		BEBAPIURL:                emTestEnsemble.eventMeshMock.MessagingURL,
		ClientID:                 "foo-id",
		ClientSecret:             "foo-secret",
		TokenEndpoint:            emTestEnsemble.eventMeshMock.TokenURL,
		WebhookActivationTimeout: 0,
		WebhookTokenEndpoint:     "foo-token-endpoint",
		Domain:                   domain,
		EventTypePrefix:          reconcilertesting.EventMeshPrefix,
		BEBNamespace:             reconcilertesting.EventMeshNamespaceNS,
		Qos:                      string(eventMeshtypes.QosAtLeastOnce),
		EnableNewCRDVersion:      true,
	}
}

func tearDownSuite() error {
	if k8sCancelFn != nil {
		k8sCancelFn()
	}
	err := emTestEnsemble.testEnv.Stop()
	emTestEnsemble.eventMeshMock.Stop()
	return err
}

func startNewEventMeshMock() *reconcilertestingv1.BEBMock {
	emMock := reconcilertestingv1.NewBEBMock()
	emMock.Start()
	return emMock
}

func getTestNamespace() string {
	return fmt.Sprintf("ns-%s", utils.GetRandString(namespacePrefixLength))
}

func ensureNamespaceCreated(ctx context.Context, t *testing.T, namespace string) {
	if namespace == "default" {
		return
	}
	// create namespace
	ns := fixtureNamespace(namespace)
	err := emTestEnsemble.k8sClient.Create(ctx, ns)
	if !k8serrors.IsAlreadyExists(err) {
		require.NoError(t, err)
	}
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

func ensureK8sResourceCreated(ctx context.Context, t *testing.T, obj client.Object) {
	require.NoError(t, emTestEnsemble.k8sClient.Create(ctx, obj))
}

func ensureK8sResourceNotCreated(ctx context.Context, t *testing.T, obj client.Object, err error) {
	require.Equal(t, emTestEnsemble.k8sClient.Create(ctx, obj), err)
}

func ensureK8sResourceUpdated(ctx context.Context, t *testing.T, obj client.Object) {
	require.NoError(t, emTestEnsemble.k8sClient.Update(ctx, obj))
}

func ensureK8sResourceDeleted(ctx context.Context, t *testing.T, obj client.Object) {
	require.NoError(t, emTestEnsemble.k8sClient.Delete(ctx, obj))
}

func ensureK8sSubscriptionUpdated(ctx context.Context, t *testing.T, subscription *eventingv1alpha2.Subscription) {
	require.Eventually(t, func() bool {
		latestSubscription := &eventingv1alpha2.Subscription{}
		lookupKey := types.NamespacedName{
			Namespace: subscription.Namespace,
			Name:      subscription.Name,
		}
		require.NoError(t, emTestEnsemble.k8sClient.Get(ctx, lookupKey, latestSubscription))
		require.NotEmpty(t, latestSubscription.Name)
		latestSubscription.Spec = subscription.Spec
		latestSubscription.Labels = subscription.Labels
		require.NoError(t, emTestEnsemble.k8sClient.Update(ctx, latestSubscription))
		return true
	}, bigTimeOut, bigPollingInterval)
}

// ensureAPIRuleStatusUpdatedWithStatusReady updates the status fof the APIRule (mocking APIGateway controller).
func ensureAPIRuleStatusUpdatedWithStatusReady(ctx context.Context, t *testing.T, apiRule *apigatewayv1beta1.APIRule) {
	require.Eventually(t, func() bool {
		fetchedAPIRule, err := getAPIRule(ctx, apiRule)
		if err != nil {
			return false
		}

		newAPIRule := fetchedAPIRule.DeepCopy()
		// mark the ApiRule status as ready
		reconcilertesting.MarkReady(newAPIRule)

		// update ApiRule status on k8s
		err = emTestEnsemble.k8sClient.Status().Update(ctx, newAPIRule)
		return err == nil
	}, bigTimeOut, bigPollingInterval)
}

// ensureAPIRuleNotFound ensures that a APIRule does not exists (or deleted).
func ensureAPIRuleNotFound(ctx context.Context, t *testing.T, apiRule *apigatewayv1beta1.APIRule) {
	require.Eventually(t, func() bool {
		apiRuleKey := client.ObjectKey{
			Namespace: apiRule.Namespace,
			Name:      apiRule.Name,
		}

		apiRule2 := new(apigatewayv1beta1.APIRule)
		err := emTestEnsemble.k8sClient.Get(ctx, apiRuleKey, apiRule2)
		return k8serrors.IsNotFound(err)
	}, bigTimeOut, bigPollingInterval)
}

func getAPIRulesList(ctx context.Context, svc *corev1.Service) (*apigatewayv1beta1.APIRuleList, error) {
	labels := map[string]string{
		constants.ControllerServiceLabelKey:  svc.Name,
		constants.ControllerIdentityLabelKey: constants.ControllerIdentityLabelValue,
	}
	apiRules := &apigatewayv1beta1.APIRuleList{}
	err := emTestEnsemble.k8sClient.List(ctx, apiRules, &client.ListOptions{
		LabelSelector: k8slabels.SelectorFromSet(labels),
		Namespace:     svc.Namespace,
	})
	return apiRules, err
}

func getAPIRule(ctx context.Context, apiRule *apigatewayv1beta1.APIRule) (*apigatewayv1beta1.APIRule, error) {
	lookUpKey := types.NamespacedName{
		Namespace: apiRule.Namespace,
		Name:      apiRule.Name,
	}
	err := emTestEnsemble.k8sClient.Get(ctx, lookUpKey, apiRule)
	return apiRule, err
}

func filterAPIRulesForASvc(apiRules *apigatewayv1beta1.APIRuleList, svc *corev1.Service) apigatewayv1beta1.APIRule {
	if len(apiRules.Items) == 1 && *apiRules.Items[0].Spec.Service.Name == svc.Name {
		return apiRules.Items[0]
	}
	return apigatewayv1beta1.APIRule{}
}

// countEventMeshRequests returns how many requests for a given subscription are sent for each HTTP method
//
//nolint:gocognit
func countEventMeshRequests(subscriptionName, eventType string) (int, int, int) {
	countGet, countPost, countDelete := 0, 0, 0
	emTestEnsemble.eventMeshMock.Requests.ReadEach(
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

func getEventMeshSubFromMock(subscriptionName, subscriptionNamespace string) *eventMeshtypes.Subscription {
	key := getEventMeshSubKeyForMock(subscriptionName, subscriptionNamespace)
	return emTestEnsemble.eventMeshMock.Subscriptions.GetSubscription(key)
}

func getEventMeshSubKeyForMock(subscriptionName, subscriptionNamespace string) string {
	nm1 := emTestEnsemble.nameMapper.MapSubscriptionName(subscriptionName, subscriptionNamespace)
	return fmt.Sprintf("%s/%s", eventMeshMockKeyPrefix, nm1)
}

func getEventMeshKeyForMock(name string) string {
	return fmt.Sprintf("%s/%s", eventMeshMockKeyPrefix, name)
}

// ensureK8sEventReceived checks if a certain event have triggered for the given namespace.
func ensureK8sEventReceived(t *testing.T, event corev1.Event, namespace string) {
	ctx := context.TODO()
	require.Eventually(t, func() bool {
		// get all events from k8s for namespace
		eventList := &corev1.EventList{}
		err := emTestEnsemble.k8sClient.List(ctx, eventList, client.InNamespace(namespace))
		require.NoError(t, err)

		// find the desired event
		var receivedEvent *corev1.Event
		for i, e := range eventList.Items {
			if e.Reason == event.Reason {
				receivedEvent = &eventList.Items[i]
				break
			}
		}

		// check the received event
		require.NotNil(t, receivedEvent)
		require.Equal(t, receivedEvent.Reason, event.Reason)
		require.Equal(t, receivedEvent.Message, event.Message)
		require.Equal(t, receivedEvent.Type, event.Type)
		return true
	}, bigTimeOut, bigPollingInterval)
}
