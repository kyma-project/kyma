package backupAndRestore

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	. "github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/backupe2e"
	"github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/utils"
	. "github.com/smartystreets/goconvey/convey"
)

var testUUID = uuid.New()
var backupName = "test-" + testUUID.String()

type e2eTest struct {
	backupTest BackupTest
	namespace  string
	testUUID   string
}

func TestBackupAndRestoreCluster(t *testing.T) {

	myFunctionTest, err := NewFunctionTest()

	if err != nil {
		t.Fatalf("%v", err)
	}

	myStatefulSetTest, err := NewStatefulSetTest()

	if err != nil {
		t.Fatalf("%v", err)
	}

	myDeploymentTest, err := NewDeploymentTest()

	if err != nil {
		t.Fatalf("%v", err)
	}

	myPrometheusTest, err := NewPrometheusTest()

	if err != nil {
		t.Fatalf("%v", err)
	}

	myNamespaceControllerTest, err := NewNamespaceControllerTest()

	if err != nil {
		t.Fatalf("%v", err)
	}

	apiControllerTest, err := NewApiControllerTest()

	if err != nil {
		t.Fatalf("%v", err)
	}

	myGrafanaTest, err := NewGrafanaTest()

	if err != nil {
		t.Fatalf("%v", err)
	}

	myMicrofrontendTest, err := NewMicrofrontendTest()

	if err != nil {
		t.Fatalf("%v", err)
	}

	backupTests := []BackupTest{
		myPrometheusTest,
		myFunctionTest,
		myDeploymentTest,
		myStatefulSetTest,
		myNamespaceControllerTest,
		apiControllerTest,
		myGrafanaTest,
		myMicrofrontendTest,
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

	myBackupClient, err := utils.NewBackupClient()
	if err != nil {
		t.Fatalf("%v", err)
	}

	Convey("Create resources", t, func() {
		for _, e2eTest := range e2eTests {

			err := myBackupClient.CreateNamespace(e2eTest.namespace)
			So(err, ShouldBeNil)
			e2eTest.backupTest.CreateResources(e2eTest.namespace)
		}
		for _, e2eTest := range e2eTests {
			e2eTest.backupTest.TestResources(e2eTest.namespace)
		}
	})

	Convey("Backup Cluster", t, func() {
		systemBackupSpecFile := "/system-backup.yaml"
		allBackupSpecFile := "/all-backup.yaml"
		allBackupName := "all-" + backupName
		systemBackupName := "system-" + backupName
		err := myBackupClient.CreateBackup(allBackupName, allBackupSpecFile)
		So(err, ShouldBeNil)
		err = myBackupClient.CreateBackup(systemBackupName, systemBackupSpecFile)
		So(err, ShouldBeNil)

		Convey("Check backup status", func() {
			err := myBackupClient.WaitForBackupToBeCreated(allBackupName, 20*time.Minute)
			myBackupClient.DescribeBackup(allBackupName)
			So(err, ShouldBeNil)
			err = myBackupClient.WaitForBackupToBeCreated(systemBackupName, 20*time.Minute)
			myBackupClient.DescribeBackup(systemBackupName)
			So(err, ShouldBeNil)
			Convey("Delete resources from cluster", func() {
				for _, e2eTest := range e2eTests {
					e2eTest.backupTest.DeleteResources()
					err := myBackupClient.DeleteNamespace(e2eTest.namespace)
					So(err, ShouldBeNil)
					err = myBackupClient.WaitForNamespaceToBeDeleted(e2eTest.namespace, 2*time.Minute)
					So(err, ShouldBeNil)
				}
				Convey("Restore Cluster", func() {
					err := myBackupClient.RestoreBackup(allBackupName)
					So(err, ShouldBeNil)
					err = myBackupClient.RestoreBackup(systemBackupName)
					So(err, ShouldBeNil)
					err = myBackupClient.WaitForBackupToBeRestored(allBackupName, 15*time.Minute)
					myBackupClient.DescribeRestore(allBackupName)
					So(err, ShouldBeNil)
					err = myBackupClient.WaitForBackupToBeRestored(systemBackupName, 15*time.Minute)
					myBackupClient.DescribeRestore(systemBackupName)
					So(err, ShouldBeNil)
					Convey("Test restored resources", func() {
						for _, e2eTest := range e2eTests {
							e2eTest.backupTest.TestResources(e2eTest.namespace)
						}
					})
				})
			})
		})
	})
}
