package testsuite

import (
	"github.com/avast/retry-go"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/testkit"
	"github.com/pkg/errors"
)

type CheckCounterPod struct {
	testService *testkit.TestService
}

var _ step.Step = &CheckCounterPod{}

func NewCheckCounterPod(testService *testkit.TestService) *CheckCounterPod {
	return &CheckCounterPod{
		testService: testService,
	}
}

func (s *CheckCounterPod) Name() string {
	return "Check counter pod"
}

func (s *CheckCounterPod) Run() error {
	err := retry.Do(func() error {
		return s.testService.WaitForCounterPodToUpdateValue(1)
	})
	if err != nil {
		return errors.Wrapf(err, "the counter pod is not updated")
	}
	return nil
}

func (s *CheckCounterPod) Cleanup() error {
	return nil
}
