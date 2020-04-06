package common

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/client"
	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/tests/eventmesh"
	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/tests/function"
	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/tests/helloworld"
	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/tests/monitoring"
	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/tests/ory"
	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/tests/servicecatalog"
	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/tests/ui"
)

type TestMode int

const (
	TestBeforeBackup TestMode = iota
	TestAfterRestore
)

type e2eTest struct {
	enabled    bool
	name       string
	backupTest client.BackupTest
	namespace  string
}

// RunTest executes a series of different tests either before or after a Backup is taken
func RunTest(t *testing.T, mode TestMode) {
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

	myMicroFrontendTest, err := ui.NewMicrofrontendTest()
	fatalOnError(t, err, "while creating structure for MicroFrontend test")

	appBrokerTest, err := servicecatalog.NewAppBrokerTest()
	fatalOnError(t, err, "while creating structure for AppBroker test")

	helmBrokerTest, err := servicecatalog.NewHelmBrokerTest()
	fatalOnError(t, err, "while creating structure for HelmBroker test")

	myOryScenarioTest, err := ory.NewHydraOathkeeperTest()
	fatalOnError(t, err, "while creating structure for Ory test")

	myApiGatewayScenarioTest, err := ory.NewApiGatewayTest()
	fatalOnError(t, err, "while creating structure for Api-Gateway test")

	myEventMeshTest, err := eventmesh.NewEventMeshTest()
	fatalOnError(t, err, "while creating structure for event mesh test")

	//rafterTest := rafter.NewRafterTest(client)

	backupTests := []e2eTest{
		{enabled: true, backupTest: myPrometheusTest},
		{enabled: false, backupTest: myGrafanaTest}, //disabled due to flakiness
		{enabled: true, backupTest: myFunctionTest},
		{enabled: true, backupTest: myDeploymentTest},
		{enabled: true, backupTest: myStatefulSetTest},
		{enabled: true, backupTest: scAddonsTest},
		{enabled: true, backupTest: myMicroFrontendTest},
		{enabled: true, backupTest: appBrokerTest},
		{enabled: true, backupTest: helmBrokerTest},
		{enabled: true, backupTest: myOryScenarioTest},
		{enabled: true, backupTest: myApiGatewayScenarioTest},
		{enabled: true, backupTest: myEventMeshTest},
		// Rafter is not enabled yet in Kyma
		// rafterTest,
	}
	e2eTests := make([]e2eTest, len(backupTests))

	for idx, backupTest := range backupTests {

		name := string("")
		if t := reflect.TypeOf(backupTest.backupTest); t.Kind() == reflect.Ptr {
			name = t.Elem().Name()
		} else {
			name = t.Name()
		}

		e2eTests[idx] = e2eTest{
			backupTest: backupTest.backupTest,
			enabled:    backupTest.enabled,
			name:       name,
			namespace:  fmt.Sprintf("%s-backup-test", strings.ToLower(name)),
		}
	}

	myBackupClient, err := client.NewBackupClient()
	fatalOnError(t, err, "while creating custom client for Backup")

	switch mode {
	case TestBeforeBackup:
		for _, e2eTest := range e2eTests {
			t.Run(e2eTest.name, func(t *testing.T) {
				if !e2eTest.enabled {
					t.Skip("Test disabled")
				}
				t.Logf("[CreateResources: %s] Creating Namespace: %s", e2eTest.name, e2eTest.namespace)
				err := myBackupClient.CreateNamespace(e2eTest.namespace)
				require.NoError(t, err)
				t.Logf("[CreateResources: %s] Starting execution", e2eTest.name)
				e2eTest.backupTest.CreateResources(t, e2eTest.namespace)
				t.Logf("[CreateResources: %s] End with success", e2eTest.name)
				t.Logf("[TestResources: %s] Starting execution", e2eTest.name)
				e2eTest.backupTest.TestResources(t, e2eTest.namespace)
				t.Logf("[TestResources: %s] End with success", e2eTest.name)
			})
		}
	case TestAfterRestore:
		for _, e2eTest := range e2eTests {
			t.Run(e2eTest.name, func(t *testing.T) {
				if !e2eTest.enabled {
					t.Skip("Test disabled")
				}
				t.Logf("[TestResources: %s] Starting execution", e2eTest.name)
				e2eTest.backupTest.TestResources(t, e2eTest.namespace)
				t.Logf("[TestResources: %s] End with success", e2eTest.name)
			})
		}
	default:
		t.Fatalf("Unrecognized mode")
	}
}

func fatalOnError(t *testing.T, err error, context string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", context, err)
	}
}
