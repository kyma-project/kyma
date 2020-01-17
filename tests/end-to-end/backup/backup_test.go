package backuptest

import (
	"flag"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/client"
	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/tests/apicontroller"
	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/tests/eventbus"
	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/tests/function"
	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/tests/helloworld"
	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/tests/monitoring"
	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/tests/servicecatalog"
	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/tests/ui"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	testBeforeBackupActionName = "testBeforeBackup"
	testAfterRestoreActionName = "testAfterRestore"
)

type e2eTest struct {
	name       string
	backupTest client.BackupTest
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

	myFunctionTest, err := function.NewFunctionTest()
	fatalOnError(t, err, "while creating structure for Function test")

	myStatefulSetTest, err := helloworld.NewStatefulSetTest()
	fatalOnError(t, err, "while creating structure for StatefulSet test")

	myDeploymentTest, err := helloworld.NewDeploymentTest()
	fatalOnError(t, err, "while creating structure for Deployment test")

	myPrometheusTest, err := monitoring.NewPrometheusTest()
	fatalOnError(t, err, "while creating structure for Prometheus test")

	myGrafanaTest, err := monitoring.NewGrafanaTest()
	fatalOnError(t, err, "while creating structure for Grafana test")

	scAddonsTest, err := servicecatalog.NewServiceCatalogAddonsTest()
	fatalOnError(t, err, "while creating structure for ScAddons test")

	apiControllerTest, err := apicontroller.NewApiControllerTestFromEnv()
	fatalOnError(t, err, "while creating structure for ApiController test")

	myMicroFrontendTest, err := ui.NewMicrofrontendTest()
	fatalOnError(t, err, "while creating structure for MicroFrontend test")

	appBrokerTest, err := servicecatalog.NewAppBrokerTest()
	fatalOnError(t, err, "while creating structure for AppBroker test")

	helmBrokerTest, err := servicecatalog.NewHelmBrokerTest()
	fatalOnError(t, err, "while creating structure for HelmBroker test")

	myEventBusTest, err := eventbus.NewEventBusTest()
	fatalOnError(t, err, "while creating structure for EventBus test")

	//rafterTest := rafter.NewRafterTest(client)

	backupTests := []client.BackupTest{
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
			name:       name,
			backupTest: backupTest,
			namespace:  fmt.Sprintf("%s-backup-test", strings.ToLower(name)),
		}
	}

	myBackupClient, err := client.NewBackupClient()
	fatalOnError(t, err, "while creating custom client for Backup")

	switch action {
	case testBeforeBackupActionName:
		Convey("Create resources\n", t, func() {
			for _, e2eTest := range e2eTests {
				t.Logf("Creating Namespace: %s", e2eTest.namespace)
				err := myBackupClient.CreateNamespace(e2eTest.namespace)
				So(err, ShouldBeNil)
				t.Logf("[CreateResources: %s] Starting execution", e2eTest.name)
				e2eTest.backupTest.CreateResources(e2eTest.namespace)
				t.Logf("[CreateResources: %s] End with success", e2eTest.name)
			}
			for _, e2eTest := range e2eTests {
				t.Logf("[TestResources: %s] Starting execution", e2eTest.name)
				e2eTest.backupTest.TestResources(e2eTest.namespace)
				t.Logf("[TestResources: %s] End with success", e2eTest.name)
			}
		})
	case testAfterRestoreActionName:
		Convey("Test restored resources\n", t, func() {
			for _, e2eTest := range e2eTests {
				t.Logf("[TestResources: %s] Starting execution", e2eTest.name)
				e2eTest.backupTest.TestResources(e2eTest.namespace)
				t.Logf("[TestResources: %s] End with success", e2eTest.name)
			}
		})
	default:
		t.Fatalf("Unrecognized runner action. Allowed actions: %s or %s.", testBeforeBackupActionName, testAfterRestoreActionName)
	}
}

func fatalOnError(t *testing.T, err error, context string) {
	if err != nil {
		t.Fatalf("%s: %v", context, err)
	}
}
