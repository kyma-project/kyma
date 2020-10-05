package step_test

import (
	"errors"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/step"
	"github.com/onsi/gomega"
	"github.com/sirupsen/logrus/hooks/test"
	"testing"
)

func TestParallelRunner(t *testing.T) {
	//GIVEN
	g := gomega.NewWithT(t)
	logger, hook := test.NewNullLogger()
	logf := logger.WithField("Test", "Test")
	idx := 0

	steps := []step.Step{
		step.Parallel(logf,
			testStep{name: "step 1", counter: &idx, logf: logf},
			testStep{name: "step 2", counter: &idx, logf: logf},
			testStep{name: "step 3", counter: &idx, logf: logf},
			testStep{name: "step 4", err: errors.New("Error Attention"), counter: &idx, logf: logf},
			testStep{name: "step 3", counter: &idx, logf: logf},
		),
	}
	runner := step.NewRunner(step.WithCleanupDefault(step.CleanupModeYes), step.WithLogger(logger))

	//WHEN
	err := runner.Execute(steps)

	//THEN
	g.Expect(err).ToNot(gomega.BeNil())
	g.Expect(err.Error()).To(gomega.ContainSubstring("while executing step: step 4"))
	g.Expect(idx).To(gomega.Equal(1))

	errLog := getLogsContains(hook.AllEntries(), "OnError")
	g.Expect(len(errLog)).To(gomega.Equal(1))
	g.Expect(errLog[0].Message).To(gomega.Equal("OnError: step 4"))

	hook.Reset()
}
