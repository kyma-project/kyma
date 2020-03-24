package testsuite

import (
	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/retry"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
)

// CheckCounterPod is a step which checks if counter has been updated in test pod
type CheckCounterPod struct {
	testService   *testkit.TestService
	expectedValue int
}

var _ step.Step = &CheckCounterPod{}

// NewCheckCounterPod returns new CheckCounterPod
func NewCheckCounterPod(testService *testkit.TestService, expectedValue int) *CheckCounterPod {
	return &CheckCounterPod{
		testService:   testService,
		expectedValue: expectedValue,
	}
}

// Name returns name name of the step
func (s *CheckCounterPod) Name() string {
	return "Check counter pod"
}

// Run executes the step
func (s *CheckCounterPod) Run() error {
	err := retry.Do(func() error {
		return s.testService.WaitForCounterPodToUpdateValue(s.expectedValue)
	})
	if err != nil {
		return errors.Wrapf(err, "the counter pod is not updated")
	}
	return nil
}

// Cleanup removes all resources that may possibly created by the step
func (s *CheckCounterPod) Cleanup() error {
	return nil
}
