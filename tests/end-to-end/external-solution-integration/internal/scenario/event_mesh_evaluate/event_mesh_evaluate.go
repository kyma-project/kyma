package event_mesh_evaluate

import (
	"time"

	"github.com/avast/retry-go"
	coreclient "k8s.io/client-go/kubernetes"
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
	retryOpts = []retry.Option{
		retry.Attempts(retryAttemptsCount),
		retry.Delay(retryDelay),
	}
)

// Steps return scenario steps
func (s *Scenario) Steps(config *rest.Config) ([]step.Step, error) {
	coreClientset := coreclient.NewForConfigOrDie(config)
	testService := testkit.NewTestService(
		internal.NewHTTPClient(internal.WithSkipSSLVerification(s.SkipSSLVerify)),
		nil,
		nil,
		nil,
		s.Domain,
		s.TestID,
		"", // no need for an image as we just want to reuse the existing service
	)

	state := s.NewState()
	dataStore := testkit.NewDataStore(coreClientset, s.TestID)

	return []step.Step{
		testsuite.NewLoadStoredCertificates(dataStore, state),
		step.Retry(
			testsuite.NewSendEventToMeshAndCheckEventId(
				s.TestID,
				helpers.FunctionPayload,
				state,
				testService,
				retryOpts...),
		).WithRetryOptions(
			retry.Attempts(3),
			retry.DelayType(retry.FixedDelay),
			retry.Delay(500*time.Millisecond)),
		step.Retry(
			testsuite.NewSendEventToCompatibilityLayerAndCheckEventId(
				s.TestID,
				helpers.FunctionPayload,
				state,
				testService,
				retryOpts...),
		).WithRetryOptions(
			retry.Attempts(3),
			retry.DelayType(retry.FixedDelay),
			retry.Delay(500*time.Millisecond)),
	}, nil
}
