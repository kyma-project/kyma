package testsuite

import (
	"github.com/avast/retry-go"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
	"github.com/pkg/errors"
	"time"
)

// CheckCounterPod is a step which checks if counter has been updated in test pod
type CheckCounterPod struct {
	testService *testkit.TestService
}

var _ step.Step = &CheckCounterPod{}

// NewCheckCounterPod returns new CheckCounterPod
func NewCheckCounterPod(testService *testkit.TestService) *CheckCounterPod {
	return &CheckCounterPod{
		testService: testService,
	}
}

// Name returns name name of the step
func (s *CheckCounterPod) Name() string {
	return "Check counter pod"
}

// Run executes the step
func (s *CheckCounterPod) Run() error {
	err := retry.Do(func() error {
		return s.testService.WaitForCounterPodToUpdateValue(1)
	}, retry.Attempts(2), retry.Delay(time.Second), retry.DelayType(retry.FixedDelay))
	if err != nil {
		return errors.Wrapf(err, "the counter pod is not updated")
	}
	return nil
}

// Cleanup removes all resources that may possibly created by the step
func (s *CheckCounterPod) Cleanup() error {
	return nil
}
