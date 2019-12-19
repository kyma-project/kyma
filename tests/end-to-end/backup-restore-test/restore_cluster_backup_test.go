package backupandrestore

import (
	"flag"
	"fmt"
	"reflect"
	"strings"
	"testing"

	. "github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/backupe2e"
	. "github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/backupe2e/service-catalog"
	backupClient "github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/utils/backup"
	"github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

var log = logrus.WithField("test", "backup-restore")

const (
	testBeforeBackupActionName = "testBeforeBackup"
	testAfterRestoreActionName = "testAfterRestore"
)

type e2eTest struct {
	backupTest BackupTest
	namespace  string
}

var action string

func init() {
	actionUsage := fmt.Sprintf("Define what kind of action runner should execute. Possible values: %s or %s", testBeforeBackupActionName, testAfterRestoreActionName)
	flag.StringVar(&action, "action", "", actionUsage)
	flag.Parse()
}

func TestBackupAndRestoreCluster(t *testing.T) {
	//cfg, err := config.NewRestClientConfig()
	//fatalOnError(t, err, "while creating rest client")
	//
	//client, err := dynamic.NewForConfig(cfg)
	//fatalOnError(t, err, "while creating dynamic client")

	myFunctionTest, err := NewFunctionTest()
	fatalOnError(t, err, "while creating structure for Function test")

	myStatefulSetTest, err := NewStatefulSetTest()
	fatalOnError(t, err, "while creating structure for StatefulSet test")

	myDeploymentTest, err := NewDeploymentTest()
	fatalOnError(t, err, "while creating structure for Deployment test")

	myPrometheusTest, err := NewPrometheusTest()
	fatalOnError(t, err, "while creating structure for Prometheus test")

	myGrafanaTest, err := NewGrafanaTest()
	fatalOnError(t, err, "while creating structure for Grafana test")

	scAddonsTest, err := NewServiceCatalogAddonsTest()
	fatalOnError(t, err, "while creating structure for ScAddons test")

	apiControllerTest, err := NewApiControllerTestFromEnv()
	fatalOnError(t, err, "while creating structure for ApiController test")

	myMicroFrontendTest, err := NewMicrofrontendTest()
	fatalOnError(t, err, "while creating structure for MicroFrontend test")

	appBrokerTest, err := NewAppBrokerTest()
	fatalOnError(t, err, "while creating structure for AppBroker test")

	helmBrokerTest, err := NewHelmBrokerTest()
	fatalOnError(t, err, "while creating structure for HelmBroker test")

	myEventBusTest, err := NewEventBusTest()
	fatalOnError(t, err, "while creating structure for EventBus test")

	//rafterTest := NewRafterTest(client)

	backupTests := []BackupTest{
		myPrometheusTest,
		myGrafanaTest,
		myFunctionTest,
		myDeploymentTest,
		myStatefulSetTest,
		scAddonsTest,
		apiControllerTest,
		myMicroFrontendTest,
		appBrokerTest,
		helmBrokerTest,
		myEventBusTest,
		// Rafter is not enabled yet in Kyma
		// rafterTest,
	}
	e2eTests := make([]e2eTest, len(backupTests))

	for idx, backupTest := range backupTests {

		name := string("")
		if t := reflect.TypeOf(backupTest); t.Kind() == reflect.Ptr {
			name = t.Elem().Name()
		} else {
			name = t.Name()
		}

		e2eTests[idx] = e2eTest{
			backupTest: backupTest,
			namespace:  strings.ToLower(name) + "-backup-test",
		}
	}

	myBackupClient, err := backupClient.NewBackupClient()
	fatalOnError(t, err, "while creating custom client for Backup")

	switch action {
	case testBeforeBackupActionName:
		Convey("Create resources\n", t, func() {
			for _, e2eTest := range e2eTests {
				log.Infof("Creating Namespace: %s", e2eTest.namespace)
				err := myBackupClient.CreateNamespace(e2eTest.namespace)
				So(err, ShouldBeNil)
				e2eTest.backupTest.CreateResources(e2eTest.namespace)
			}
			for _, e2eTest := range e2eTests {
				log.Infof("Testing resources in namespace: %s", e2eTest.namespace)
				t.Logf("Testing resources in namespace: %s", e2eTest.namespace)
				e2eTest.backupTest.TestResources(e2eTest.namespace)
				t.Log(e2eTest.namespace + " is done!")
			}
		})
	case testAfterRestoreActionName:
		Convey("Test restored resources\n", t, func() {
			for _, e2eTest := range e2eTests {
				log.Infof("Testing resources in namespace: %s", e2eTest.namespace)
				t.Logf("Testing resources in namespace: %s", e2eTest.namespace)
				e2eTest.backupTest.TestResources(e2eTest.namespace)
				t.Log(e2eTest.namespace + " is done!")
			}
		})
	default:
		logrus.Fatalf("Unrecognized runner action. Allowed actions: %s or %s.", testBeforeBackupActionName, testAfterRestoreActionName)
	}
}

func fatalOnError(t *testing.T, err error, context string) {
	if err != nil {
		t.Fatalf("%s: %v", context, err)
	}
}
