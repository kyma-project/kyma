package testsuite

import (
	retrygo "github.com/avast/retry-go"
	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/retry"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
)

// CheckEventId is a step which checks if the correct EventId has been received
type CheckEventId struct {
	testService *testkit.TestService
	expectedId  string
	retryOpts   []retrygo.Option
}

var _ step.Step = &CheckEventId{}

// NewCheckEventId returns new CheckEventId
func NewCheckEventId(testService *testkit.TestService, expectedId string, opts ...retrygo.Option) *CheckEventId {
	return &CheckEventId{
		testService: testService,
		expectedId:  expectedId,
		retryOpts:   opts,
	}
}

// Name returns name name of the step
func (s *CheckEventId) Name() string {
	return "Check counter pod"
}

// Run executes the step
func (s *CheckEventId) Run() error {
	err := retry.Do(func() error {
		return s.testService.CheckTestId(s.expectedId) //s.testService.WaitForCounterPodToUpdateValue(s.expectedValue)
	}, s.retryOpts...)
	if err != nil {
		return errors.Wrapf(err, "the counter pod is not updated")
	}
	return nil
}

// Cleanup removes all resources that may possibly created by the step
func (s *CheckEventId) Cleanup() error {
	return nil
}
