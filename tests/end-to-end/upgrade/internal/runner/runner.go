package runner

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/internal/platform/logger"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
)

const (
	gracefulTimeout    = time.Second * 10
	createdByLabelName = "kyma-project.io/created-by"
	// negative value for regex used for name validation in k8s:
	// https://github.com/kubernetes/apimachinery/blob/98853ca904e81a25e2000cae7f077dc30f81b85f/pkg/util/validation/validation.go#L110
	regexSanitize = "[^a-z0-9]([^-a-z0-9]*[^a-z0-9])?"
)

// NamespaceCreator has methods required to create k8s ns.
type NamespaceCreator interface {
	Create(*v1.Namespace) (*v1.Namespace, error)
}

// TestRunner executes registered tests
type TestRunner struct {
	log                 *logrus.Entry
	nsCreator           NamespaceCreator
	tests               map[string]UpgradeTest
	maxConcurrencyLevel int
	sanitizeRegex       *regexp.Regexp
	verbose             bool
}

// NewTestRunner is a constructor for TestRunner
func NewTestRunner(log *logrus.Entry, nsCreator NamespaceCreator, tests map[string]UpgradeTest, maxConcurrencyLevel int, verbose bool) (*TestRunner, error) {
	sanitizeRegex, err := regexp.Compile(regexSanitize)
	if err != nil {
		return nil, errors.Wrap(err, "while compiling sanitize regexp")
	}

	return &TestRunner{
		log:                 log.WithField("service", "test:runner"),
		nsCreator:           nsCreator,
		tests:               tests,
		maxConcurrencyLevel: maxConcurrencyLevel,
		sanitizeRegex:       sanitizeRegex,
		verbose:             verbose,
	}, nil
}

// PrepareData executes CreateResources method for each registered test.
// Test are run in parallel with given maxConcurrencyLevel
func (r *TestRunner) PrepareData(stopCh <-chan struct{}) error {
	wg := sync.WaitGroup{}
	queue := make(chan task)

	failedTaskCnt := 0
	// spawn workers
	for i := 0; i < r.maxConcurrencyLevel; i++ {
		wg.Add(1)
		go func() {
			for test := range queue {
				failed := r.executeTaskFunc(test.CreateResources, stopCh, "CreateResources", test.Name(), true)
				if failed {
					failedTaskCnt++
				}
			}
			wg.Done()
		}()
	}

	// populate all tests
	for name, test := range r.tests {
		queue <- task{name, test}
	}
	close(queue)

	// wait for the workers to finish
	r.wgWait(stopCh, &wg)

	if failedTaskCnt > 0 {
		return fmt.Errorf("executed %d task and %d of them failed", len(r.tests), failedTaskCnt)
	}
	return nil
}

// ExecuteTests executes TestResources method for each registered test.
// Test are run in parallel with given maxConcurrencyLevel
func (r *TestRunner) ExecuteTests(stopCh <-chan struct{}) error {
	wg := sync.WaitGroup{}
	queue := make(chan task)

	failedTaskCnt := 0
	// spawn workers
	for i := 0; i < r.maxConcurrencyLevel; i++ {
		wg.Add(1)
		go func() {
			for test := range queue {
				failed := r.executeTaskFunc(test.TestResources, stopCh, "TestResources", test.Name(), false)
				if failed {
					failedTaskCnt++
				}
			}
			wg.Done()
		}()
	}

	// populate all tests
	for name, test := range r.tests {
		queue <- task{name, test}
	}
	close(queue)

	// wait for the workers to finish
	r.wgWait(stopCh, &wg)

	if failedTaskCnt > 0 {
		return fmt.Errorf("executed %d task and %d of them failed", len(r.tests), failedTaskCnt)
	}

	return nil
}

func (r *TestRunner) executeTaskFunc(taskHandler taskFn, stopCh <-chan struct{}, header, taskName string, createNs bool) bool {
	fullHeader := fmt.Sprintf("[%s: %s]", header, taskName)

	taskLog, err := r.newLoggerForTask()
	if err != nil {
		r.log.Errorf("%s Cannot create uuid, the task won't be started, got err: %v", fullHeader, err)
		return true
	}

	if r.shutdownRequested(stopCh) {
		taskLog.Debugf("Stop channel called. Not executing %s", fullHeader)
		return true
	}

	taskLog.Infof("%s Starting execution", fullHeader)

	nsName := r.sanitizedNamespaceName(taskName)
	if createNs {
		if err := r.ensureNamespaceExists(nsName); err != nil {
			taskLog.Errorf("Cannot create namespace %q: %v", nsName, err)
			return true
		}
	}

	sink := &bytes.Buffer{}
	originalOutput := taskLog.Logger.Out
	// verbose means "do not suppress logs"
	if !r.verbose {
		taskLog.Logger.SetOutput(sink)
	}
	startTime := time.Now()

	if err := taskHandler(stopCh, taskLog, nsName); err != nil {
		taskLog.Errorf("%s End with error [start time: %v, duration: %v]: %v", fullHeader, startTime.Format(time.RFC1123), time.Since(startTime), err)
		fmt.Fprint(originalOutput, sink.String())
		return true
	}

	sink.Reset() // if task succeeded then do not print data logged internally in that test
	taskLog.Infof("%s End with success [start time: %v, duration: %v]", fullHeader, startTime.Format(time.RFC1123), time.Since(startTime))
	fmt.Fprint(originalOutput, sink.String())

	return false
}

// sanitizedNamespaceName returns sanitized name base on rules from this site:
// https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
func (r *TestRunner) sanitizedNamespaceName(nameToSanitize string) string {
	nsName := strings.ToLower(nameToSanitize)
	nsName = r.sanitizeRegex.ReplaceAllString(nsName, "-")

	if len(nsName) > 253 {
		nsName = nsName[:253]
	}

	return nsName
}

func (r *TestRunner) ensureNamespaceExists(name string) error {
	_, err := r.nsCreator.Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"env":              "true",
				"istio-injection":  "enabled",
				createdByLabelName: "e2e-upgrade-test-runner",
			},
		},
	})

	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	return nil
}

func (r *TestRunner) shutdownRequested(stopCh <-chan struct{}) bool {
	select {
	case <-stopCh:
		return true
	default:
	}
	return false
}

func (r *TestRunner) generateTaskID() (string, error) {
	uuidInstance, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	return uuidInstance.String(), nil
}

// wgWait waits for wg with respection to stopCh
func (r *TestRunner) wgWait(stopCh <-chan struct{}, wg *sync.WaitGroup) {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		r.log.Debug("All task finished.")
	case <-stopCh:
		r.log.Infof("Stop channel called. Waiting %v for task to finish their job.", gracefulTimeout)
		select {
		case <-done:
		case <-time.After(gracefulTimeout):
			r.log.Errorf("Task didn't finished in %v after calling stop channel.", gracefulTimeout)
		}
	}
}

// newLoggerForTask returns new logger which can be used in given task.
// We need to create new instance of logger otherwise we will start
// mix logs between each task cause they will share same instance.
func (r *TestRunner) newLoggerForTask() (*logrus.Entry, error) {
	taskID, err := r.generateTaskID()
	if err != nil {
		return nil, errors.Wrap(err, "while generating ID for task logger")
	}

	cfg := &logger.Config{
		Level: logger.LogLevel(r.log.Logger.Level),
	}
	taskLog := logger.New(cfg).WithField("taskID", taskID)

	return taskLog, nil
}
