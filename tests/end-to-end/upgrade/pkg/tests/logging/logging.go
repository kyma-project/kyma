package logging

import (
	"net/http"

	dex "github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/fetch-dex-token"

	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/tests/logging/pkg/jwt"
	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/tests/logging/pkg/logstream"
	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/tests/logging/pkg/request"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

// LoggingTest checks if logging continues to work properly after upgrade
type LoggingTest struct {
	coreInterface kubernetes.Interface
	domainName    string
	idpConfig     dex.IdProviderConfig
	httpClient    *http.Client
}

// NewLoggingTest creates a new instance of logging upgrade test
func NewLoggingTest(coreInterface kubernetes.Interface, domainName string, dexConfig dex.IdProviderConfig) LoggingTest {
	return LoggingTest{
		coreInterface: coreInterface,
		domainName:    domainName,
		idpConfig:     dexConfig,
		httpClient:    request.GetHttpClient(),
	}
}

// CreateResources creates resources for logging upgrade test
func (t LoggingTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	log.Println("Cleaning up before creating resources")
	if err := logstream.Cleanup(namespace, t.coreInterface); err != nil {
		return err
	}
	log.Println("Deploying test-counter-pod")
	if err := logstream.DeployDummyPod(namespace, t.coreInterface); err != nil {
		return err
	}
	log.Println("Waiting for test-counter-pod to run...")
	if err := logstream.WaitForDummyPodToRun(namespace, t.coreInterface); err != nil {
		return err
	}
	log.Println("Test if logs from test-counter-pod are streamed by Loki before upgrade")
	if err := t.testLogStream(namespace); err != nil {
		return err
	}
	log.Println("Checking that a JWT token is required for accessing Loki before upgrade")
	if err := t.checkTokenIsRequired(); err != nil {
		return err
	}
	log.Println("Checking that sending a request to Loki with a wrong path is forbidden before upgrade")
	if err := t.checkWrongPathIsForbidden(); err != nil {
		return err
	}
	return nil
}

// TestResources checks if resources are working properly after upgrade
func (t LoggingTest) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	log.Println("Test if new logs from test-counter-pod are streamed by Loki after upgrade")
	if err := t.testLogStream(namespace); err != nil {
		return err
	}
	log.Println("Deleting test-counter-pod")
	if err := logstream.Cleanup(namespace, t.coreInterface); err != nil {
		return err
	}
	log.Println("Checking that a JWT token is required for accessing Loki after upgrade")
	if err := t.checkTokenIsRequired(); err != nil {
		return err
	}
	log.Println("Checking that sending a request to Loki with a wrong path is forbidden after upgrade")
	if err := t.checkWrongPathIsForbidden(); err != nil {
		return err
	}
	return nil
}

func (t LoggingTest) testLogStream(namespace string) error {
	token, err := jwt.GetToken(t.idpConfig)
	if err != nil {
		return errors.Wrap(err, "cannot fetch dex token")
	}
	authHeader := jwt.SetAuthHeader(token)
	if err := logstream.Test("container", "count", authHeader, t.httpClient); err != nil {
		return err
	}
	if err := logstream.Test("app", "test-counter-pod", authHeader, t.httpClient); err != nil {
		return err
	}
	if err := logstream.Test("namespace", namespace, authHeader, t.httpClient); err != nil {
		return err
	}
	return nil
}

func (t LoggingTest) checkTokenIsRequired() error {
	lokiURL := "http://logging-loki.kyma-system:3100/api/prom"
	// sending a request to Loki wihtout a JWT token in the header
	respStatus, _, err := request.DoGet(t.httpClient, lokiURL, "")
	if err != nil {
		return errors.Wrap(err, "cannot send request to Loki")
	}
	if respStatus != http.StatusUnauthorized {
		return errors.Errorf("received status code %d instead of %d when accessing Loki wihout a JWT token", respStatus, http.StatusUnauthorized)
	}
	return nil
}

func (t LoggingTest) checkWrongPathIsForbidden() error {
	token, err := jwt.GetToken(t.idpConfig)
	if err != nil {
		return errors.Wrap(err, "cannot fetch dex token")
	}
	authHeader := jwt.SetAuthHeader(token)
	lokiURL := "http://logging-loki.kyma-system:3100/api/wrongPath"
	// sending a request with a wrong path
	respStatus, _, err := request.DoGet(t.httpClient, lokiURL, authHeader)
	if err != nil {
		return errors.Wrap(err, "cannot send request to Loki")
	}
	if respStatus != http.StatusForbidden {
		return errors.Errorf("received status code %d instead of %d when sending a request to Loki with a wrong path", respStatus, http.StatusForbidden)
	}
	return nil
}
