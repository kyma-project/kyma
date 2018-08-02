package statusmanager

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/installer/pkg/consts"

	installationv1alpha1 "github.com/kyma-project/kyma/components/installer/pkg/apis/installer/v1alpha1"
	fake "github.com/kyma-project/kyma/components/installer/pkg/client/clientset/versioned/fake"
	installationInformers "github.com/kyma-project/kyma/components/installer/pkg/client/informers/externalversions"
	. "github.com/smartystreets/goconvey/convey"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestStatusManager(t *testing.T) {
	Convey("Status Manager InProgress", t, func() {
		Convey("should return error if kyma installation is not found", func() {
			testStatusManager := getTestSetup()

			err := testStatusManager.InProgress("installing kyma")

			So(err, ShouldNotBeNil)
		})

		Convey("should update state and description only", func() {
			expectedStatus := installationv1alpha1.StateInProgress
			expectedDescription := "installing kyma"

			givenURL := "fakeURL"
			givenVersion := "0.0.1"

			testInst := &installationv1alpha1.Installation{
				ObjectMeta: metav1.ObjectMeta{
					Name:      consts.InstResource,
					Namespace: consts.InstNamespace,
				},
				Spec: installationv1alpha1.InstallationSpec{
					URL:         givenURL,
					KymaVersion: givenVersion,
				},
			}
			testStatusManager := getTestSetup(testInst)

			err := testStatusManager.InProgress(expectedDescription)

			kymaInst, _ := testStatusManager.client.InstallerV1alpha1().Installations(consts.InstNamespace).Get(consts.InstResource, metav1.GetOptions{})

			So(err, ShouldBeNil)
			So(kymaInst.Status.State, ShouldEqual, expectedStatus)
			So(kymaInst.Status.Description, ShouldEqual, expectedDescription)
			So(kymaInst.Status.URL, ShouldEqual, "")
			So(kymaInst.Status.KymaVersion, ShouldEqual, "")
		})

		Convey("should update state and description only leaving old url and version", func() {
			oldState := installationv1alpha1.StateInstalled
			oldDescription := "Kyma installed"
			oldURL := "installedURL"
			oldVersion := "0.0.1"

			testState := installationv1alpha1.StateInProgress
			testDescription := "installing kyma"
			testURL := "fakeURL"
			testVersion := "0.0.2"

			testInst := &installationv1alpha1.Installation{
				ObjectMeta: metav1.ObjectMeta{
					Name:      consts.InstResource,
					Namespace: consts.InstNamespace,
				},
				Spec: installationv1alpha1.InstallationSpec{
					URL:         testURL,
					KymaVersion: testVersion,
				},
				Status: installationv1alpha1.InstallationStatus{
					State:       oldState,
					Description: oldDescription,
					URL:         oldURL,
					KymaVersion: oldVersion,
				},
			}
			testStatusManager := getTestSetup(testInst)

			err := testStatusManager.InProgress(testDescription)

			kymaInst, _ := testStatusManager.client.InstallerV1alpha1().Installations(consts.InstNamespace).Get(consts.InstResource, metav1.GetOptions{})

			So(err, ShouldBeNil)
			So(kymaInst.Status.State, ShouldEqual, testState)
			So(kymaInst.Status.Description, ShouldEqual, testDescription)
			So(kymaInst.Status.URL, ShouldEqual, oldURL)
			So(kymaInst.Status.KymaVersion, ShouldEqual, oldVersion)
		})
	})

	Convey("Status Manager Error", t, func() {

		Convey("should update state and description only", func() {
			expectedState := installationv1alpha1.StateError
			expectedDescription := "installing kyma"

			givenURL := "fakeURL"
			givenVersion := "0.0.1"

			testInst := &installationv1alpha1.Installation{
				ObjectMeta: metav1.ObjectMeta{
					Name:      consts.InstResource,
					Namespace: consts.InstNamespace,
				},
				Spec: installationv1alpha1.InstallationSpec{
					URL:         givenURL,
					KymaVersion: givenVersion,
				},
			}
			testStatusManager := getTestSetup(testInst)

			err := testStatusManager.Error(expectedDescription)

			kymaInst, _ := testStatusManager.client.InstallerV1alpha1().Installations(consts.InstNamespace).Get(consts.InstResource, metav1.GetOptions{})

			So(err, ShouldBeNil)
			So(kymaInst.Status.State, ShouldEqual, expectedState)
			So(kymaInst.Status.Description, ShouldEqual, expectedDescription)
			So(kymaInst.Status.URL, ShouldEqual, "")
			So(kymaInst.Status.KymaVersion, ShouldEqual, "")
		})

		Convey("should update state, description only and clear URL and KymaVersion", func() {
			oldState := installationv1alpha1.StateInstalled
			oldDescription := "kyma installed"
			oldURL := "installedURL"
			oldVersion := "0.0.1"

			testDescription := "updating kyma"
			testState := installationv1alpha1.StateError
			testURL := "fakeURL"
			testVersion := "0.0.2"

			testInst := &installationv1alpha1.Installation{
				ObjectMeta: metav1.ObjectMeta{
					Name:      consts.InstResource,
					Namespace: consts.InstNamespace,
				},
				Spec: installationv1alpha1.InstallationSpec{
					URL:         testURL,
					KymaVersion: testVersion,
				},
				Status: installationv1alpha1.InstallationStatus{
					State:       oldState,
					Description: oldDescription,
					URL:         oldURL,
					KymaVersion: oldVersion,
				},
			}
			testStatusManager := getTestSetup(testInst)

			err := testStatusManager.Error(testDescription)

			kymaInst, _ := testStatusManager.client.InstallerV1alpha1().Installations(consts.InstNamespace).Get(consts.InstResource, metav1.GetOptions{})

			So(err, ShouldBeNil)
			So(kymaInst.Status.State, ShouldEqual, testState)
			So(kymaInst.Status.Description, ShouldEqual, testDescription)
			So(kymaInst.Status.URL, ShouldEqual, "")
			So(kymaInst.Status.KymaVersion, ShouldEqual, "")
		})
	})

	Convey("Status Manager Done", t, func() {

		Convey("should update state, description, url and kyma version after installation", func() {
			oldState := installationv1alpha1.StateInProgress
			oldDescription := "installing kyma"
			oldURL := "installedURL"
			oldVersion := "0.0.1"

			testState := installationv1alpha1.StateInstalled
			testDescription := "Kyma installed"
			testURL := "fakeURL"
			testVersion := "0.0.2"

			testInst := &installationv1alpha1.Installation{
				ObjectMeta: metav1.ObjectMeta{
					Name:      consts.InstResource,
					Namespace: consts.InstNamespace,
				},
				Spec: installationv1alpha1.InstallationSpec{
					URL:         testURL,
					KymaVersion: testVersion,
				},
				Status: installationv1alpha1.InstallationStatus{
					State:       oldState,
					Description: oldDescription,
					URL:         oldURL,
					KymaVersion: oldVersion,
				},
			}
			testStatusManager := getTestSetup(testInst)

			err := testStatusManager.InstallDone(testURL, testVersion)

			kymaInst, _ := testStatusManager.client.InstallerV1alpha1().Installations(consts.InstNamespace).Get(consts.InstResource, metav1.GetOptions{})

			So(err, ShouldBeNil)
			So(kymaInst.Status.State, ShouldEqual, testState)
			So(kymaInst.Status.Description, ShouldEqual, testDescription)
			So(kymaInst.Status.URL, ShouldEqual, testURL)
			So(kymaInst.Status.KymaVersion, ShouldEqual, testVersion)
		})

		Convey("should update state, description, url and kyma version after update", func() {
			oldState := installationv1alpha1.StateInstalled
			oldDescription := "installing kyma"
			oldURL := "installedURL"
			oldVersion := "0.0.1"

			testState := installationv1alpha1.StateUpdated
			testDescription := "Kyma updated"
			testURL := "fakeURL"
			testVersion := "0.0.2"

			testInst := &installationv1alpha1.Installation{
				ObjectMeta: metav1.ObjectMeta{
					Name:      consts.InstResource,
					Namespace: consts.InstNamespace,
				},
				Spec: installationv1alpha1.InstallationSpec{
					URL:         testURL,
					KymaVersion: testVersion,
				},
				Status: installationv1alpha1.InstallationStatus{
					State:       oldState,
					Description: oldDescription,
					URL:         oldURL,
					KymaVersion: oldVersion,
				},
			}
			testStatusManager := getTestSetup(testInst)

			err := testStatusManager.UpdateDone(testURL, testVersion)

			kymaInst, _ := testStatusManager.client.InstallerV1alpha1().Installations(consts.InstNamespace).Get(consts.InstResource, metav1.GetOptions{})

			So(err, ShouldBeNil)
			So(kymaInst.Status.State, ShouldEqual, testState)
			So(kymaInst.Status.Description, ShouldEqual, testDescription)
			So(kymaInst.Status.URL, ShouldEqual, testURL)
			So(kymaInst.Status.KymaVersion, ShouldEqual, testVersion)
		})
	})
}

func getTestSetup(installations ...runtime.Object) *statusManager {
	fakeClient := fake.NewSimpleClientset(installations...)
	informers := installationInformers.NewSharedInformerFactory(fakeClient, time.Second*0)
	installationLister := informers.Installer().V1alpha1().Installations().Lister()

	if len(installations) > 0 {
		for ind := range installations {
			informers.Installer().V1alpha1().Installations().Informer().GetIndexer().Add(installations[ind])
		}
	}

	testStatusManager := &statusManager{
		client: fakeClient,
		lister: installationLister,
	}

	return testStatusManager
}
