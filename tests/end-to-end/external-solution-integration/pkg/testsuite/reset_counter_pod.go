package testsuite

import (
	retrygo "github.com/avast/retry-go"
	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/retry"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
)

// ResetCounterPod resets the request counter deployed as a service
type ResetCounterPod struct {
	testService *testkit.TestService
	retryOpts   []retrygo.Option
}

var _ step.Step = &ResetCounterPod{}

// NewResetCounterPod returns new ResetCounterPod
func NewResetCounterPod(testService *testkit.TestService, opts ...retrygo.Option) *ResetCounterPod {
	return &ResetCounterPod{
		testService: testService,
		retryOpts:   opts,
	}
}

// Name returns name name of the step
func (s *ResetCounterPod) Name() string {
	return "Reset counter pod"
}

// Run executes the step
func (s *ResetCounterPod) Run() error {
	err := retry.Do(func() error {
		if err := s.testService.Reset(); err != nil {
			return err
		}
		return s.testService.WaitForCounterPodToUpdateValue(0)
	}, s.retryOpts...)
	if err != nil {
		return errors.Wrapf(err, "the counter pod is not updated")
	}
	return nil
}

// Cleanup removes all resources that may possibly created by the step
func (s *ResetCounterPod) Cleanup() error {
	return nil
}
