package step_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/step"
	"github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

func TestParallelRunner(t *testing.T) {
	//GIVEN
	g := gomega.NewWithT(t)
	logger, hook := test.NewNullLogger()
	//logger = logrus.New()
	logf := logger.WithField("Test", "Test")
	idx := 0

	steps := step.NewSerialTestRunner(logf, "Test1",
		testStep{name: "start 0", counter: &idx, logf: logf},
		step.NewParallelRunner(logf, "Parallel Step",
			testStep{name: "step 1", counter: &idx, logf: logf},
			testStep{name: "step 2", counter: &idx, logf: logf},
			testStep{name: "step 3", counter: &idx, logf: logf},
			testStep{name: "step 4", err: errors.New("Error Attention"), counter: &idx, logf: logf},
			testStep{name: "step 5", counter: &idx, logf: logf},
		),
	)
	runner := step.NewRunner(step.WithCleanupDefault(step.CleanupModeYes), step.WithLogger(logger))

	//WHEN
	err := runner.Execute(steps)

	//THEN
	g.Expect(err).ToNot(gomega.BeNil())
	g.Expect(idx).To(gomega.Equal(2))
	g.Expect(err.Error()).To(gomega.ContainSubstring("while executing step: step 4"))

	logEntries := hook.AllEntries()
	entry, nextLogIdx := getFirstMatchingLog(logEntries, "Step: step 4, returned error: Error Attention", 0)
	g.Expect(entry).ToNot(gomega.BeNil())

	entry, nextLogIdx = getFirstMatchingLog(logEntries, "Called on Error, resource: step 4", nextLogIdx)
	g.Expect(entry).ToNot(gomega.BeNil())

	entry, nextLogIdx = getFirstMatchingLog(logEntries, "Called on Error, resource: start 0", nextLogIdx)
	g.Expect(entry).ToNot(gomega.BeNil())

	entry, nextLogIdx = getFirstMatchingLog(logEntries, "Cleanup Serial Step: Parallel:Parallel Step", nextLogIdx)
	g.Expect(entry).ToNot(gomega.BeNil())

	entry, nextLogIdx = getFirstMatchingLog(logEntries, "Cleanup Serial Step: start 0", nextLogIdx)
	g.Expect(entry).ToNot(gomega.BeNil())
	hook.Reset()
}

func getFirstMatchingLog(entries []*logrus.Entry, text string, startIdx int) (*logrus.Entry, int) {
	for i := startIdx; i < len(entries); i++ {
		if strings.Contains(entries[i].Message, text) {
			return entries[i], i
		}
	}
	return nil, -1
}

func TestMixed(t *testing.T) {
	//GIVEN
	g := gomega.NewWithT(t)
	//logger, hook := test.NewNullLogger()
	logger := logrus.New()
	logger.SetReportCaller(true)
	logf := logger.WithField("TestSuite", "Test")
	log1 := logf.WithField("Test", "suite1")
	log2 := logf.WithField("Test", "suite2")
	log3 := logf.WithField("Test", "suite3")
	idx := 0

	//TODO: Check if array of steps is ok, becasue I see doubled logs
	//TODO: Analyze the log from this output
	steps := step.NewSerialTestRunner(logf, "Suite",
		testStep{name: "outside of parallel", counter: &idx, logf: logf},
		step.NewParallelRunner(logf, "test",
			step.NewSerialTestRunner(log1, "Test Serial Runner",
				testStep{name: "step 1.1", counter: &idx, logf: log1},
				testStep{name: "step 1.2", counter: &idx, logf: log1},
				testStep{name: "step 1.3", counter: &idx, logf: log1},
				testStep{name: "step 1.4", err: errors.New("Error Attention"), counter: &idx, logf: log1},
				testStep{name: "step 1.5", counter: &idx, logf: log1},
			),
			step.NewSerialTestRunner(log2, "Test Serial Runner",
				testStep{name: "step 2.1", counter: &idx, logf: log2},
				testStep{name: "step 2.2", counter: &idx, logf: log2},
				testStep{name: "step 2.3", counter: &idx, logf: log2},
				//testStep{name: "step 4", err: errors.New("Error Attention"), counter: &idx, logf: logf},
				testStep{name: "step 2.4", counter: &idx, logf: log2},
			),
			step.NewSerialTestRunner(log3, "Test Serial Runner",
				testStep{name: "step 3.1", counter: &idx, logf: log3},
				testStep{name: "step 3.2", counter: &idx, logf: log3},
				testStep{name: "step 3.3", counter: &idx, logf: log3},
				//testStep{name: "step 4", err: errors.New("Error Attention"), counter: &idx, logf: logf},
				testStep{name: "step 3.4", counter: &idx, logf: logf},
			),
		),
		testStep{name: "outside of parallel", counter: &idx, logf: logf})
	runner := step.NewRunner(step.WithCleanupDefault(step.CleanupModeYes), step.WithLogger(logger))

	//WHEN
	err := runner.Execute(steps)

	//THEN
	g.Expect(err).ToNot(gomega.BeNil())
	g.Expect(err.Error()).To(gomega.ContainSubstring("while executing step: step 1.4"))
	g.Expect(idx).To(gomega.Equal(5))

	//errLog := getLogsContains(hook.AllEntries(), "OnError")
	//g.Expect(len(errLog)).To(gomega.Equal(4))
	//g.Expect(errLog[0].Message).To(gomega.Equal("OnError: step 4"))
	//g.Expect(errLog[1].Message).To(gomega.Equal("OnError: step 3"))
	//g.Expect(errLog[2].Message).To(gomega.Equal("OnError: step 2"))
	//g.Expect(errLog[3].Message).To(gomega.Equal("OnError: step 1"))
	//
	//hook.Reset()
}
