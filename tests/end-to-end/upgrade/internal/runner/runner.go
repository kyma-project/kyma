package runner

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

// TestRunner executes registered tests
type TestRunner struct {
	log                 logrus.FieldLogger
	tests               []UpgradeTest
	maxConcurrencyLevel int
}

// NewTestRunner is a constructor for TestRunner
func NewTestRunner(log logrus.FieldLogger, tests []UpgradeTest, maxConcurrencyLevel int) *TestRunner {
	return &TestRunner{
		log:                 log.WithField("service", "test:runner"),
		tests:               tests,
		maxConcurrencyLevel: maxConcurrencyLevel,
	}
}

// PrepareData executes CreateResources method for each registered test.
// Test are run in parallel with given maxConcurrencyLevel
func (r *TestRunner) PrepareData(stopCh <-chan struct{}) error {
	wg := sync.WaitGroup{}
	queue := make(chan UpgradeTest)

	// spawn workers
	for i := 0; i < r.maxConcurrencyLevel; i++ {
		wg.Add(1)
		go func() {
			for test := range queue {
				headerMsg := fmt.Sprintf("[CreateResources: %q]", test.Name())
				r.executeFunction(test.CreateResources, stopCh, headerMsg)
			}
			wg.Done()
		}()
	}

	// populate all tests
	for _, test := range r.tests {
		queue <- test
	}
	close(queue)

	// wait for the workers to finish
	if canceled := r.wgWait(stopCh, &wg); canceled {
		r.log.Infof("Stop channel called when waiting for workers to finish")
		return nil
	}

	return nil
}

// ExecuteTests executes TestResources method for each registered test.
// Test are run in parallel with given maxConcurrencyLevel
func (r *TestRunner) ExecuteTests(stopCh <-chan struct{}) error {
	wg := sync.WaitGroup{}
	queue := make(chan UpgradeTest)

	// spawn workers
	for i := 0; i < r.maxConcurrencyLevel; i++ {
		wg.Add(1)
		go func() {
			for test := range queue {
				headerMsg := fmt.Sprintf("[TestResources: %q]", test.Name())
				r.executeFunction(test.TestResources, stopCh, headerMsg)
			}
			wg.Done()
		}()
	}

	// populate all tests
	for _, test := range r.tests {
		queue <- test
	}
	close(queue)

	// wait for the workers to finish
	if canceled := r.wgWait(stopCh, &wg); canceled {
		r.log.Infof("Stop channel called when waiting for workers to finish")
		return nil
	}

	return nil
}

type m func(stopCh <-chan struct{}, log logrus.FieldLogger, namespace string) error

func (r *TestRunner) executeFunction(method m, stopCh <-chan struct{}, headerMsg string) {
	// TODO: generate namespace

	testID := r.generateTestID()
	testLogger := r.log.WithField("ID", testID).Logger
	testLogger.Infof("Starting %s", headerMsg)

	memLog := &bytes.Buffer{}
	testLogger.SetOutput(memLog)

	startTime := time.Now()
	if err := method(stopCh, testLogger, testID); err != nil {
		testLogger.Errorf("%s end with error [start time: %v, duration: %v]: %v", headerMsg, startTime, time.Since(startTime), err)
	} else {
		memLog.Reset() // if test succeeded then do not print data logged internally in that test
		testLogger.Infof("%s end with success [start time: %v, duration: %v]", headerMsg, startTime, time.Since(startTime))
	}

	fmt.Print(memLog.String())
}

// generateTestID generates random test ID
func (r *TestRunner) generateTestID() string {
	return uuid.NewV4().String()
}

func (r *TestRunner) wgWait(stopCh <-chan struct{}, wg *sync.WaitGroup) bool {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-stopCh:
		return true
	}
	return false
}
