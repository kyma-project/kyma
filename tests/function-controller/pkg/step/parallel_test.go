package step_test

import (
	"errors"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/step"
	"github.com/onsi/gomega"
	"github.com/sirupsen/logrus/hooks/test"
)

func TestParallelRunner(t *testing.T) {
	//GIVEN
	g := gomega.NewWithT(t)
	logger, hook := test.NewNullLogger()
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

	entry, nextLogIdx = getFirstMatchingLog(logEntries, "Cleanup Serial Step: Parallel: Parallel Step", nextLogIdx)
	g.Expect(entry).ToNot(gomega.BeNil())

	entry, nextLogIdx = getFirstMatchingLog(logEntries, "Cleanup Serial Step: start 0", nextLogIdx)
	g.Expect(entry).ToNot(gomega.BeNil())
	hook.Reset()
}

func TestMixedRunners(t *testing.T) {
	//GIVEN
	g := gomega.NewWithT(t)
	logger, hook := test.NewNullLogger()
	logf := logger.WithField("TestSuite", "Test")
	log1 := logf.WithField("Test", "suite1")
	log2 := logf.WithField("Test", "suite2")
	log3 := logf.WithField("Test", "suite3")
	idx := 0

	steps := step.NewSerialTestRunner(logf, "Suite",
		testStep{name: "Init Step", counter: &idx, logf: logf},
		step.NewParallelRunner(logf, "test",
			step.NewSerialTestRunner(log1, "Test Serial Runner",
				testStep{name: "step 1.1", counter: &idx, logf: log1},
				testStep{name: "step 1.2", counter: &idx, logf: log1},
				testStep{name: "step 1.3", counter: &idx, logf: log1},
				testStep{name: "step 1.4", counter: &idx, logf: log1},
			),
			step.NewSerialTestRunner(log2, "Fault Serial Runner",
				testStep{name: "step 2.1", counter: &idx, logf: log2},
				testStep{name: "step 2.2", counter: &idx, logf: log2},
				testStep{name: "step 2.3", counter: &idx, logf: log2},
				testStep{name: "step 2.4", err: errors.New("Error Attention"), counter: &idx, logf: log2},
				testStep{name: "step 2.5", counter: &idx, logf: log2},
			),
			step.NewSerialTestRunner(log3, "Test Serial Runner",
				testStep{name: "step 3.1", counter: &idx, logf: log3},
				testStep{name: "step 3.2", counter: &idx, logf: log3},
				testStep{name: "step 3.3", counter: &idx, logf: log3},
				testStep{name: "step 3.4", counter: &idx, logf: logf},
			),
		),
		testStep{name: "Finish Step", counter: &idx, logf: logf})
	runner := step.NewRunner(step.WithCleanupDefault(step.CleanupModeYes), step.WithLogger(logger))

	//WHEN
	err := runner.Execute(steps)

	//THEN
	g.Expect(err).ToNot(gomega.BeNil())
	g.Expect(err.Error()).To(gomega.ContainSubstring("while executing step: step 2.4"))
	g.Expect(idx).To(gomega.Equal(5))

	allLogs := hook.AllEntries()

	step2Logs := getLogs(allLogs, "Test", "suite2")
	g.Expect(len(step2Logs)).To(gomega.Equal(13))
	g.Expect(step2Logs[0].Message).To(gomega.Equal("Running Step 0: step 2.1"))
	g.Expect(step2Logs[1].Message).To(gomega.Equal("Running Step 1: step 2.2"))
	g.Expect(step2Logs[2].Message).To(gomega.Equal("Running Step 2: step 2.3"))
	g.Expect(step2Logs[3].Message).To(gomega.Equal("Running Step 3: step 2.4"))

	g.Expect(step2Logs[4].Message).To(gomega.Equal("Error in step 2.4, error: Error Attention"))

	g.Expect(step2Logs[5].Message).To(gomega.Equal("Called on Error, resource: step 2.4"))
	g.Expect(step2Logs[6].Message).To(gomega.Equal("Called on Error, resource: step 2.3"))
	g.Expect(step2Logs[7].Message).To(gomega.Equal("Called on Error, resource: step 2.2"))
	g.Expect(step2Logs[8].Message).To(gomega.Equal("Called on Error, resource: step 2.1"))

	g.Expect(step2Logs[9].Message).To(gomega.Equal("Cleanup Serial Step: step 2.4"))
	g.Expect(step2Logs[10].Message).To(gomega.Equal("Cleanup Serial Step: step 2.3"))
	g.Expect(step2Logs[11].Message).To(gomega.Equal("Cleanup Serial Step: step 2.2"))
	g.Expect(step2Logs[12].Message).To(gomega.Equal("Cleanup Serial Step: step 2.1"))

	errorLogs := getLogsWithLevel(allLogs, logrus.ErrorLevel)

	g.Expect(len(errorLogs)).To(gomega.Equal(3))
	g.Expect(errorLogs[0].Message).To(gomega.Equal("Error in step 2.4, error: Error Attention"))
	g.Expect(errorLogs[1].Message).To(gomega.ContainSubstring("Fault Serial Runner, Steps: 0:step 2.1, 1:step 2.2, 2:step 2.3, 3:step 2.4, 4:step 2.5., returned error: while executing step: step 2.4: Error Attention"))
	g.Expect(errorLogs[2].Message).To(gomega.ContainSubstring("Error in Parallel: test"))

	hook.Reset()
}
