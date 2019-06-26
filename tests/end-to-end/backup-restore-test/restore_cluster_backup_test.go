package backupAndRestore

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	. "github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/backupe2e"

	. "github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/backupe2e/asset-store"
	. "github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/backupe2e/cms"

	. "github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/backupe2e/service-catalog"

	backupClient "github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/utils/backup"
	"github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/vrischmann/envconfig"
)

var testUUID = uuid.New()
var backupName = "test-" + testUUID.String()
var log = logrus.WithField("test", "backup-restore")

type config struct {
	AllBackupConfigurationFile    string `envconfig:"default=/all-backup.yaml"`
	SystemBackupConfigurationFile string `envconfig:"default=/system-backup.yaml"`
}

type e2eTest struct {
	backupTest BackupTest
	namespace  string
	testUUID   string
}

func TestBackupAndRestoreCluster(t *testing.T) {
	cfg, err := loadConfig()
	fatalOnError(t, err, "while reading configuration from environment variables")

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

	myAssetStoreTest, err := NewAssetStoreTest(t)
	fatalOnError(t, err, "while creating structure for AssetStore test")

	myMicroFrontendTest, err := NewMicrofrontendTest()
	fatalOnError(t, err, "while creating structure for MicroFrontend test")

	appBrokerTest, err := NewAppBrokerTest()
	fatalOnError(t, err, "while creating structure for AppBroker test")

	// helmBrokerTest, err := NewHelmBrokerTest()
	// fatalOnError(t, err, "while creating structure for HelmBroker test")

	myEventBusTest, err := NewEventBusTest()
	fatalOnError(t, err, "while creating structure for EventBus test")

	myCmsTest, err := NewCmsTest(t)
	fatalOnError(t, err, "while creating structure for Cms test")

	backupTests := []BackupTest{
		myPrometheusTest,
		myFunctionTest,
		myDeploymentTest,
		myStatefulSetTest,
		myGrafanaTest,
		scAddonsTest,
		myCmsTest,
		myAssetStoreTest,
		apiControllerTest,
		myMicroFrontendTest,
		appBrokerTest,
		// helmBrokerTest,
		myEventBusTest,
	}
	e2eTests := make([]e2eTest, len(backupTests))

	for idx, backupTest := range backupTests {
		testUUID := uuid.New()

		name := string("")
		if t := reflect.TypeOf(backupTest); t.Kind() == reflect.Ptr {
			name = t.Elem().Name()
		} else {
			name = t.Name()
		}

		e2eTests[idx] = e2eTest{
			backupTest: backupTest,
			namespace:  strings.ToLower(name) + "-" + testUUID.String(),
			testUUID:   testUUID.String(),
		}
	}

	myBackupClient, err := backupClient.NewBackupClient()
	fatalOnError(t, err, "while creating custom client for Backup")

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

	Convey("Backup Cluster", t, func() {
		allBackupSpecFile := cfg.AllBackupConfigurationFile
		allBackupName := "all-" + backupName

		systemBackupSpecFile := cfg.SystemBackupConfigurationFile
		systemBackupName := "system-" + backupName

		err = myBackupClient.CreateBackup(systemBackupName, systemBackupSpecFile)
		So(err, ShouldBeNil)
		err := myBackupClient.CreateBackup(allBackupName, allBackupSpecFile)
		So(err, ShouldBeNil)

		Convey("Check backup status", func() {

			err = myBackupClient.WaitForBackupToBeCreated(systemBackupName, 25*time.Minute)
			myBackupClient.DescribeBackup(systemBackupName)
			So(err, ShouldBeNil)

			err := myBackupClient.WaitForBackupToBeCreated(allBackupName, 25*time.Minute)
			myBackupClient.DescribeBackup(allBackupName)
			So(err, ShouldBeNil)

			Convey("Delete resources from cluster\n", func() {
				for _, e2eTest := range e2eTests {
					e2eTest.backupTest.DeleteResources(e2eTest.namespace)

					err := myBackupClient.DeleteNamespace(e2eTest.namespace)
					So(err, ShouldBeNil)

					err = myBackupClient.WaitForNamespaceToBeDeleted(e2eTest.namespace, 2*time.Minute)
					So(err, ShouldBeNil)
				}

				Convey("Restore Cluster", func() {
					err = myBackupClient.RestoreBackup(systemBackupName)
					So(err, ShouldBeNil)
					err := myBackupClient.RestoreBackup(allBackupName)
					So(err, ShouldBeNil)

					err = myBackupClient.WaitForBackupToBeRestored(systemBackupSpecFile, 25*time.Minute)
					So(err, ShouldBeNil)

					err = myBackupClient.WaitForBackupToBeRestored(allBackupName, 20*time.Minute)
					myBackupClient.DescribeRestore(allBackupName)
					So(err, ShouldBeNil)

					Convey("Test restored resources\n", func() {
						for _, e2eTest := range e2eTests {
							e2eTest.backupTest.TestResources(e2eTest.namespace)
						}
					})
				})
			})
		})
	})
}

func loadConfig() (config, error) {
	var cfg config
	err := envconfig.Init(&cfg)
	if err != nil {
		return config{}, err
	}

	return cfg, nil
}

func fatalOnError(t *testing.T, err error, context string) {
	if err != nil {
		t.Fatalf("%s: %v", context, err)
	}
}
