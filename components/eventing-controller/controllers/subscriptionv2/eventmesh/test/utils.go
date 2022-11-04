package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/constants"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"

	"github.com/stretchr/testify/assert"

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
}

const (
	useExistingCluster       = false
	attachControlPlaneOutput = false
	testEnvStartDelay        = time.Minute
	testEnvStartAttempts     = 10
	bigPollingInterval       = 3 * time.Second
	bigTimeOut               = 40 * time.Second
	smallTimeOut             = 5 * time.Second
	smallPollingInterval     = 1 * time.Second
	domain                   = "domain.com"
	namespacePrefixLength    = 5
)

var (
	emTestEnsemble    *eventMeshTestEnsemble
	acceptableMethods = []string{http.MethodPost, http.MethodOptions}
	k8sCancelFn       context.CancelFunc
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
	useExistingCluster := useExistingCluster
	emTestEnsemble.testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("../../../../", "config", "crd", "bases", "eventing.kyma-project.io_eventingbackends.yaml"),
			filepath.Join("../../../../", "config", "crd", "basesv1alpha2"),
			filepath.Join("../../../../", "config", "crd", "external"),
		},
		AttachControlPlaneOutput: attachControlPlaneOutput,
		UseExistingCluster:       &useExistingCluster,
	}

	var cfg *rest.Config
	err = retry.Do(func() error {
		defer func() {
			if r := recover(); r != nil {
				log.Println("panic recovered:", r)
			}
		}()

		cfgLocal, err := emTestEnsemble.testEnv.Start()
		cfg = cfgLocal
		return err
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

	if err != nil || cfg == nil {
		return err
	}

	err = eventingv1alpha2.AddToScheme(scheme.Scheme)
	if err != nil {
		return err
	}

	err = apigatewayv1beta1.AddToScheme(scheme.Scheme)
	if err != nil {
		return err
	}
	// +kubebuilder:scaffold:scheme

	// start event mesh mock
	emTestEnsemble.eventMeshMock = startNewEventMeshMock()

	// client, err := client.New()
	// Source: https://book.kubebuilder.io/cronjob-tutorial/writing-tests.html

	var metricsPort int
	metricsPort, err = reconcilertesting.GetFreePort()
	if err != nil {
		return err
	}

	syncPeriod := time.Second * 2
	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme.Scheme,
		SyncPeriod:         &syncPeriod,
		MetricsBindAddress: fmt.Sprintf("localhost:%v", metricsPort),
	})
	if err != nil {
		return err
	}

	envConf := env.Config{
		BEBAPIURL:                emTestEnsemble.eventMeshMock.MessagingURL,
		ClientID:                 "foo-id",
		ClientSecret:             "foo-secret",
		TokenEndpoint:            emTestEnsemble.eventMeshMock.TokenURL,
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
	emTestEnsemble.nameMapper = backendutils.NewBEBSubscriptionNameMapper(domain, backendbeb.MaxBEBSubscriptionNameLength)
	eventMeshHandler := backendeventmesh.NewEventMesh(credentials, emTestEnsemble.nameMapper, defaultLogger)

	recorder := k8sManager.GetEventRecorderFor("eventing-controller")
	sinkValidator := sink.NewValidator(context.Background(), k8sManager.GetClient(), recorder, defaultLogger)
	fmt.Printf("starting emreconciler")
	err = eventmeshreconciler.NewReconciler(context.Background(), k8sManager.GetClient(), defaultLogger,
		recorder, envConf, eventMeshCleaner, eventMeshHandler, credentials, emTestEnsemble.nameMapper, sinkValidator).SetupUnmanaged(k8sManager)
	if err != nil {
		return err
	}

	go func() {
		var ctx context.Context
		ctx, k8sCancelFn = context.WithCancel(ctrl.SetupSignalHandler())
		err = k8sManager.Start(ctx)
		if err != nil {
			panic(err)
		}
	}()

	emTestEnsemble.k8sClient = k8sManager.GetClient()
	return nil
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
		assert.NoError(t, err)
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
	assert.NoError(t, emTestEnsemble.k8sClient.Create(ctx, obj))
}

func ensureK8sResourceUpdated(ctx context.Context, t *testing.T, obj client.Object) {
	assert.NoError(t, emTestEnsemble.k8sClient.Update(ctx, obj))
}

// ensureAPIRuleStatusUpdatedWithStatusReady updates the status fof the APIRule (mocking APIGateway controller).
func ensureAPIRuleStatusUpdatedWithStatusReady(ctx context.Context, t *testing.T, apiRule *apigatewayv1beta1.APIRule) {
	assert.Eventually(t, func() bool {
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
//nolint:unparam
func countEventMeshRequests(subscriptionName, eventType string) (countGet, countPost, countDelete int) {
	countGet, countPost, countDelete = 0, 0, 0
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
