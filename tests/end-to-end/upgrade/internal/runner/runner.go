package runner

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/internal/platform/logger"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	gracefulTimeout = time.Second * 10
	runnerLabelName = "upgrade.tester.kyma-project.io"
	regexSanitize   = "[^a-z0-9]([^-a-z0-9]*[^a-z0-9])?"
)

// NamespaceCreator has methods requried to create ns.
type NamespaceCreator interface {
	Create(*v1.Namespace) (*v1.Namespace, error)
}

// TestRunner executes registered tests
type TestRunner struct {
	log       *logrus.Entry
	nsCreator NamespaceCreator

	tests               map[string]UpgradeTest
	maxConcurrencyLevel int

	sanitizeRegex *regexp.Regexp
}

// NewTestRunner is a constructor for TestRunner
func NewTestRunner(log *logrus.Entry, nsCreator NamespaceCreator, tests map[string]UpgradeTest, maxConcurrencyLevel int) (*TestRunner, error) {
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
			for task := range queue {
				failed := r.executeTask(task.CreateResources, stopCh, "CreateResources", task.Name(), true)
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
			for task := range queue {
				failed := r.executeTask(task.TestResources, stopCh, "TestResources", task.Name(), false)
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

func (r *TestRunner) executeTask(task taskFn, stopCh <-chan struct{}, header, taskName string, createNs bool) bool {
	taskLog := r.newLoggerForTask()

	fullHeader := fmt.Sprintf("[%s: %s]", header, taskName)
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
	taskLog.Logger.SetOutput(sink)
	startTime := time.Now()

	if err := task(stopCh, taskLog, nsName); err != nil {
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
				"env":             "true",
				"istio-injection": "enabled",
				runnerLabelName:   "creator",
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

func (r *TestRunner) generateTaskID() string {
	return uuid.NewV4().String()
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
		r.log.Infof("Stop channel called. Waiting %v for task to finish their job", gracefulTimeout)
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
func (r *TestRunner) newLoggerForTask() *logrus.Entry {
	taskID := r.generateTaskID()

	cfg := &logger.Config{
		Level: logger.LogLevel(r.log.Logger.Level),
	}
	taskLog := logger.New(cfg).WithField("taskID", taskID)

	return taskLog
}
