package event_mesh_evaluate

import (
	"time"

	"github.com/avast/retry-go"
	"k8s.io/apimachinery/pkg/runtime/schema"
	coreClient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/helpers"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testsuite"
)

const (
	retryAttemptsCount = 240
	retryDelay         = 1 * time.Second
)

var (
	apiRuleRes = schema.GroupVersionResource{Group: "gateway.kyma-project.io", Version: "v1alpha1", Resource: "apirules"}
	retry_opts = []retry.Option{
		retry.Attempts(retryAttemptsCount),
		retry.Delay(retryDelay),
	}
)

// Steps return scenario steps
func (s *Scenario) Steps(config *rest.Config) ([]step.Step, error) {
	coreClientset := coreClient.NewForConfigOrDie(config)
	testService := testkit.NewTestService(
		internal.NewHTTPClient(internal.WithSkipSSLVerification(s.SkipSSLVerify)),
		nil,
		nil,
		nil,
		s.Domain,
		s.TestID,
	)

	// lambdaEndpoint := helpers.LambdaInClusterEndpoint(s.testID, s.testID, helpers.LambdaPort)
	state := s.NewState()
	dataStore := testkit.NewDataStore(coreClientset, s.TestID)
	state.SetDataStore(dataStore)

	return []step.Step{
		testsuite.NewReuseApplication(state),
		testsuite.NewResetCounterPod(testService),
		testsuite.NewSendEventToMesh(s.TestID, helpers.FunctionPayload, state),
		testsuite.NewCheckCounterPod(testService, 1, retry_opts...),
		testsuite.NewSendEventToCompatibilityLayer(s.TestID, helpers.FunctionPayload, state),
		testsuite.NewCheckCounterPod(testService, 2, retry_opts...),
	}, nil
}
