package logging

import (
	"fmt"
	"time"

	dex "github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/utils/fetch-dex-token"

	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/tests/logging/pkg/fluentbit"
	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/tests/logging/pkg/logstream"
	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/tests/logging/pkg/loki"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// LoggingTest checks if logging continues to work properly after upgrade
type LoggingTest struct {
	coreInterface kubernetes.Interface
	domainName    string
	idpConfig     dex.IdProviderConfig
}

// NewLoggingTest creates a new instance of logging upgrade test
func NewLoggingTest(coreInterface kubernetes.Interface, domainName string, dexConfig dex.IdProviderConfig) LoggingTest {
	return LoggingTest{
		coreInterface: coreInterface,
		domainName:    domainName,
		idpConfig:     dexConfig,
	}
}

// CreateResources creates resources for logging upgrade test
func (t LoggingTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	log.Println("LoggingUpgradeTest creating resources")
	// logstream.Cleanup(namespace)
	log.Println("Test if all the Loki/Fluent Bit pods are ready before upgrade")
	loki.TestPodsAreReady()
	log.Println("Test if Fluent Bit is able to find Loki before upgrade")
	fluentbit.Test()
	log.Println("Deploying test-counter-pod")
	err := t.deployDummyPod(namespace, log)
	if err != nil {
		return err
	}
	logstream.WaitForDummyPodToRun(namespace)
	log.Println("Test if logs from test-counter-pod are streamed by Loki before upgrade")
	err = t.testLogStream(namespace, log)
	if err != nil {
		return err
	}
	return nil
}

// TestResources checks if resources are working properly after upgrade
func (t LoggingTest) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	log.Println("Test if all the Loki/Fluent Bit pods are still running after upgrade")
	loki.TestPodsAreReady()
	log.Println("Test if Fluent Bit is able to find Loki after upgrade")
	fluentbit.Test()
	log.Println("Test if logs from test-counter-pod are streamed by Loki after upgrade")
	err := t.testLogStream(namespace, log)
	if err != nil {
		return err
	}
	logstream.Cleanup(namespace)
	return nil
}

func (t LoggingTest) testLogStream(namespace string, log logrus.FieldLogger) error {
	token, err := t.fetchDexToken()
	if err != nil {
		return errors.Wrap(err, "cannot fetch dex token")
	}
	authHeader := setAuthHeader(token)
	currentTime := time.Now().UnixNano()
	logstream.Test("container", "count", authHeader, currentTime)
	logstream.Test("app", "test-counter-pod", authHeader, currentTime)
	logstream.Test("namespace", namespace, authHeader, currentTime)
	log.Println("Test Logging Succeeded!")
	return nil
}

func (t LoggingTest) deployDummyPod(namespace string, log logrus.FieldLogger) error {

	labels := map[string]string{
		"app":       "test-counter-pod",
		"component": "test-counter-pod",
	}
	args := []string{"sh", "-c", "let i=1; while true; do echo \"$i: logTest-$(date)\"; let i++; sleep 2; done"}

	_, err := t.coreInterface.CoreV1().Pods(namespace).Create(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "test-counter-pod",
			Labels: labels,
		},
		Spec: corev1.PodSpec{
			RestartPolicy: "Never",
			Containers: []corev1.Container{
				{
					Name:  "count",
					Image: "alpine:3.8",
					Args:  args,
				},
			},
		},
	})
	if err != nil {
		return errors.Wrap(err, "while creating test-counter-pod")
	}
	return nil
}

func (t LoggingTest) fetchDexToken() (string, error) {
	return dex.Authenticate(t.idpConfig)
}

func setAuthHeader(token string) string {
	authHeader := fmt.Sprintf("Authorization: Bearer %s", token)
	return authHeader
}
