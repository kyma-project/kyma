package restore

import (
	"testing"
	"time"

	"github.com/google/uuid"
	. "github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/restore/backupe2e"
	"github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/restore/utils"
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
	backupTests := []BackupTest{myFunctionTest}

	e2eTests := make([]e2eTest, len(backupTests))

	for idx, backupTest := range backupTests {
		testUUID := uuid.New()
		e2eTests[idx] = e2eTest{
			backupTest: backupTest,
			namespace:  "test-" + testUUID.String(),
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
			e2eTest.backupTest.TestResources()
		}
	})

	Convey("Backup Cluster", t, func() {
		excludedNamespaces := []string{"default", "kube-public", "kube-system", "heptio-ark", "kyma-system", "kyma-installer", "kyma-integration", "istio-system", "natts"}
		excludedResources := []string{}
		myBackupClient.CreateBackup(backupName, []string{}, excludedNamespaces, []string{}, excludedResources)

		Convey("Check backup status", func() {
			err := myBackupClient.WaitForBackupToBeCreated(backupName, 5*time.Minute)
			So(err, ShouldBeNil)
			for _, e2eTest := range e2eTests {
				Convey("Delete resources from cluster", func() {
					err := myBackupClient.DeleteNamespace(e2eTest.namespace)
					So(err, ShouldBeNil)
					err = myBackupClient.WaitForNamespaceToBeDeleted(e2eTest.namespace, 120*time.Second)
					So(err, ShouldBeNil)
				})
			}
		})
	})

	Convey("Restore Cluster", t, func() {
		err := myBackupClient.RestoreBackup(backupName)
		So(err, ShouldBeNil)
		err = myBackupClient.WaitForBackupToBeRestored(backupName, 5*time.Minute)
		So(err, ShouldBeNil)
		Convey("Test restored resources", func() {
			for _, backupTest := range backupTests {
				backupTest.TestResources()
			}
		})
	})
}
