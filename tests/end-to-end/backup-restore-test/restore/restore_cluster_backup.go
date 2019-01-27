package restore


import (
. "github.com/smartystreets/goconvey/convey"
"testing"
)

func TestBackupAndRestoreCluster(t *testing.T) {

	Convey("Check cluster state", t, func() {

		t.Logf("Cluster state ...")

		Convey("Create new resources", func() {

			t.Logf("Function created ...")

			Convey("Test new created function", func() {

				t.Logf("Fucntion works properly ...")

			})

		})

	})

	Convey("Backup Cluster", t, func() {

		t.Logf("Cluster backed up ...")


		//Convey("Backup data from cluster", func() {
		//
		//	t.Logf("Data backed up ...")
		//
		//})


		Convey("Check backup status", func() {

			t.Logf("Resources deleted ...")

		})

		Convey("Delete resources from cluster", func() {

			t.Logf("Resources deleted ...")

		})

	})

	Convey("Restore Cluster", t, func() {

		t.Logf("Cluster restored ...")

		//Convey("Restore data from cluster", func() {
		//
		//	t.Logf("Data restored ...")
		//
		//})

		Convey("Test restored resources", func() {

			t.Logf("Fucntion works properly ...")

		})

	})

}