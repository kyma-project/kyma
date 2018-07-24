package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/kyma-project/kyma/tools/stability-checker/internal"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	apiCoreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestOutputLogType marks given log entry as test output
const TestOutputLogType = "test-output"

// TestRunner executes testing script
type TestRunner struct {
	log               logrus.FieldLogger
	throttle          time.Duration
	cfgMapName        string
	cfgMapClient      configMapClient
	testingScriptPath string
}

// NewTestRunner creates new instance of TestRunner
func NewTestRunner(testThrottle time.Duration, cfgMapName, testingScriptPath string, cfgMapClient configMapClient, log logrus.FieldLogger) *TestRunner {
	return &TestRunner{
		log:               log,
		throttle:          testThrottle,
		cfgMapName:        cfgMapName,
		cfgMapClient:      cfgMapClient,
		testingScriptPath: testingScriptPath,
	}
}

// Run executes tests in a loop
func (r *TestRunner) Run(ctx context.Context) {
	for {
		testID := r.generateTestID()
		testLogger := r.log.WithField("test-run-id", testID)
		testLogger.Infof("Start test")

		startTime := time.Now()
		testOutput, err := r.executeTesting()
		if err != nil {
			testLogger.Errorf("Got internal error on executing test: [%v].", err)
		} else {
			if testOutput.result {
				testLogger.Infof("Test end with success [start time: %v, duration: %v]", startTime, testOutput.duration)
				testLogger.WithField(r.markAsTestOutput()).Debugf("%v", testOutput.output)
			} else {
				testLogger.Errorf("Test end with error [start time: %v, duration: %v]", startTime, testOutput.duration)
				testLogger.WithField(r.markAsTestOutput()).Errorf("%v", testOutput.output)
			}

			if err := r.saveTestStatus(startTime, testID, testOutput.result); err != nil {
				testLogger.Errorf("Cannot save test results, got err: %v", err)
			}
		}

		if canceled := r.throttleTest(ctx); canceled {
			return
		}
	}
}

func (r *TestRunner) throttleTest(ctx context.Context) bool {
	r.log.Debugf("Throttle test for %v", r.throttle)
	select {
	case <-ctx.Done():
		r.log.Debugf("Shutdown test runner because of %v", ctx.Err())
		return true
	case <-time.After(r.throttle):
	}

	return false
}

func (r *TestRunner) saveTestStatus(startTime time.Time, id string, pass bool) error {
	key := fmt.Sprintf("%d", startTime.Unix())

	testStatus, err := json.Marshal(internal.TestStatus{
		ID:   id,
		Pass: pass,
	})
	if err != nil {
		return errors.Wrap(err, "while marshaling test status")
	}

	cfg, err := r.cfgMapClient.Get(r.cfgMapName, metaV1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "while getting config map %q", r.cfgMapName)
	}

	return r.addStatus(cfg, key, string(testStatus))
}

func (r *TestRunner) addStatus(cfg *apiCoreV1.ConfigMap, key, testStatus string) error {
	cfgCopy := cfg.DeepCopy()

	if cfgCopy.Data == nil {
		cfgCopy.Data = make(map[string]string)
	}

	if _, exists := cfgCopy.Data[key]; exists {
		return fmt.Errorf("cannot add entry %q to ConfigMap %s beacuase it already exists", key, r.cfgMapName)
	}

	cfgCopy.Data[key] = testStatus
	if _, err := r.cfgMapClient.Update(cfgCopy); err != nil {
		return errors.Wrapf(err, "while updating config map %q", r.cfgMapName)
	}
	return nil
}

type executeTestingOutput struct {
	duration time.Duration
	output   string
	result   bool
}

func (r *TestRunner) executeTesting() (*executeTestingOutput, error) {
	startTime := time.Now()

	cmd := exec.Command(r.testingScriptPath)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	_, isExitError := err.(*exec.ExitError)
	switch {
	case err == nil:
		return &executeTestingOutput{
			result:   true,
			duration: time.Since(startTime),
			output:   out.String(),
		}, nil
	case isExitError:
		return &executeTestingOutput{
			result:   false,
			duration: time.Since(startTime),
			output:   out.String(),
		}, nil
	default:
		return nil, errors.Wrap(err, "while executing testing")
	}
}

// generateTestID generates random test ID
func (r *TestRunner) generateTestID() string {
	return uuid.NewV4().String()
}

func (r *TestRunner) markAsTestOutput() (string, string) {
	return "type", TestOutputLogType
}
