package step_test

import (
	"errors"
	"fmt"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/step"
	"github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"strings"
	"testing"
)

func TestSerialTestRunner(t *testing.T) {
	//GIVEN
	g := gomega.NewWithT(t)
	logger, hook := test.NewNullLogger()
	logf := logger.WithField("Test", "Test")
	idx := 0

	steps := []step.Step{
		step.NewSerialTestRunner(logf, "Test Serial Runner",
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
	g.Expect(idx).To(gomega.Equal(4))

	errLog := getLogsContains(hook.AllEntries(), "OnError")
	g.Expect(len(errLog)).To(gomega.Equal(4))
	g.Expect(errLog[0].Message).To(gomega.Equal("OnError: step 4"))
	g.Expect(errLog[1].Message).To(gomega.Equal("OnError: step 3"))
	g.Expect(errLog[2].Message).To(gomega.Equal("OnError: step 2"))
	g.Expect(errLog[3].Message).To(gomega.Equal("OnError: step 1"))

	hook.Reset()
}


func getLogsContains(entries []*logrus.Entry, text string) []*logrus.Entry {
	filteredEntries := []*logrus.Entry{}

	for _, entry := range entries {
		if strings.Contains(entry.Message, text) {
			filteredEntries = append(filteredEntries, entry)
		}
	}

	return filteredEntries
}

type testStep struct {
	err     error
	name    string
	counter *int
	logf    *logrus.Entry
}

func (e testStep) Name() string {
	return e.name
}

func (e testStep) Run() error {
	return e.err
}

func (e testStep) Cleanup() error {
	return nil
}

func (e testStep) OnError(cause error) error {
	*e.counter++
	e.logf.Info(fmt.Sprintf("OnError: %s", e.name))
	return nil
}
