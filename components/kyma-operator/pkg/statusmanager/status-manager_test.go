package statusmanager

import (
	"errors"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/consts"

	installationv1alpha1 "github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/client/clientset/versioned/fake"
	installationInformers "github.com/kyma-project/kyma/components/kyma-operator/pkg/client/informers/externalversions"
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
			expectedComponent := "installer"
			expectedLog := "failed to do something"
			expectedError := errors.New(expectedLog)

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

			err := testStatusManager.Error(expectedComponent, expectedDescription, expectedError)

			kymaInst, _ := testStatusManager.client.InstallerV1alpha1().Installations(consts.InstNamespace).Get(consts.InstResource, metav1.GetOptions{})

			So(err, ShouldBeNil)
			So(kymaInst.Status.State, ShouldEqual, expectedState)
			So(kymaInst.Status.Description, ShouldEqual, expectedDescription)
			So(kymaInst.Status.URL, ShouldEqual, "")
			So(kymaInst.Status.KymaVersion, ShouldEqual, "")
			So(len(kymaInst.Status.ErrorLog), ShouldEqual, 1)
			So(kymaInst.Status.ErrorLog[0].Component, ShouldEqual, expectedComponent)
			So(kymaInst.Status.ErrorLog[0].Log, ShouldEqual, expectedLog)
		})

		Convey("should update state, description, append to error log and clear URL and KymaVersion", func() {
			oldState := installationv1alpha1.StateInstalled
			oldDescription := "kyma installed"
			oldURL := "installedURL"
			oldVersion := "0.0.1"
			oldErrorLog := installationv1alpha1.ErrorLogEntry{Component: "some-component", Log: "some old error", Occurrences: 3}

			testDescription := "updating kyma"
			testState := installationv1alpha1.StateError
			testURL := "fakeURL"
			testVersion := "0.0.2"
			testComponent := "other-component"
			testLog := "another error in later attempt"
			testError := errors.New(testLog)

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
					ErrorLog:    []installationv1alpha1.ErrorLogEntry{oldErrorLog},
				},
			}
			testStatusManager := getTestSetup(testInst)

			err := testStatusManager.Error(testComponent, testDescription, testError)

			kymaInst, _ := testStatusManager.client.InstallerV1alpha1().Installations(consts.InstNamespace).Get(consts.InstResource, metav1.GetOptions{})

			So(err, ShouldBeNil)
			So(kymaInst.Status.State, ShouldEqual, testState)
			So(kymaInst.Status.Description, ShouldEqual, testDescription)
			So(kymaInst.Status.URL, ShouldEqual, "")
			So(kymaInst.Status.KymaVersion, ShouldEqual, "")
			So(len(kymaInst.Status.ErrorLog), ShouldEqual, 2)
			So(kymaInst.Status.ErrorLog[0].Component, ShouldEqual, oldErrorLog.Component)
			So(kymaInst.Status.ErrorLog[0].Log, ShouldEqual, oldErrorLog.Log)
			So(kymaInst.Status.ErrorLog[0].Occurrences, ShouldEqual, oldErrorLog.Occurrences)
			So(kymaInst.Status.ErrorLog[1].Component, ShouldEqual, testComponent)
			So(kymaInst.Status.ErrorLog[1].Log, ShouldEqual, testLog)
			So(kymaInst.Status.ErrorLog[1].Occurrences, ShouldEqual, 1)
		})

		Convey("should increase the error counter if it appears again", func() {
			oldState := installationv1alpha1.StateInstalled
			oldDescription := "kyma installed"
			oldURL := "installedURL"
			oldVersion := "0.0.1"
			oldErrorLog := installationv1alpha1.ErrorLogEntry{Component: "some-component", Log: "some old error", Occurrences: 3}

			testDescription := "updating kyma"
			testState := installationv1alpha1.StateError
			testURL := "fakeURL"
			testVersion := "0.0.2"
			testComponent := oldErrorLog.Component
			testLog := oldErrorLog.Log
			testError := errors.New(testLog)

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
					ErrorLog:    []installationv1alpha1.ErrorLogEntry{oldErrorLog},
				},
			}
			testStatusManager := getTestSetup(testInst)

			err := testStatusManager.Error(testComponent, testDescription, testError)

			kymaInst, _ := testStatusManager.client.InstallerV1alpha1().Installations(consts.InstNamespace).Get(consts.InstResource, metav1.GetOptions{})

			So(err, ShouldBeNil)
			So(kymaInst.Status.State, ShouldEqual, testState)
			So(kymaInst.Status.Description, ShouldEqual, testDescription)
			So(kymaInst.Status.URL, ShouldEqual, "")
			So(kymaInst.Status.KymaVersion, ShouldEqual, "")
			So(len(kymaInst.Status.ErrorLog), ShouldEqual, 1)
			So(kymaInst.Status.ErrorLog[0].Component, ShouldEqual, oldErrorLog.Component)
			So(kymaInst.Status.ErrorLog[0].Log, ShouldEqual, oldErrorLog.Log)
			So(kymaInst.Status.ErrorLog[0].Occurrences, ShouldEqual, oldErrorLog.Occurrences+1)
		})

		Convey("should not aggregate not-following errors", func() {
			oldState := installationv1alpha1.StateInstalled
			oldDescription := "kyma installed"
			oldURL := "installedURL"
			oldVersion := "0.0.1"
			oldErrorLog1 := installationv1alpha1.ErrorLogEntry{Component: "some-component1", Log: "some old error1", Occurrences: 3}
			oldErrorLog2 := installationv1alpha1.ErrorLogEntry{Component: "some-component2", Log: "some old error1", Occurrences: 3}

			testDescription := "updating kyma"
			testState := installationv1alpha1.StateError
			testURL := "fakeURL"
			testVersion := "0.0.2"
			testComponent := oldErrorLog1.Component
			testLog := oldErrorLog1.Log
			testError := errors.New(testLog)

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
					ErrorLog:    []installationv1alpha1.ErrorLogEntry{oldErrorLog1, oldErrorLog2},
				},
			}
			testStatusManager := getTestSetup(testInst)

			err := testStatusManager.Error(testComponent, testDescription, testError)

			kymaInst, _ := testStatusManager.client.InstallerV1alpha1().Installations(consts.InstNamespace).Get(consts.InstResource, metav1.GetOptions{})

			So(err, ShouldBeNil)
			So(kymaInst.Status.State, ShouldEqual, testState)
			So(kymaInst.Status.Description, ShouldEqual, testDescription)
			So(kymaInst.Status.URL, ShouldEqual, "")
			So(kymaInst.Status.KymaVersion, ShouldEqual, "")
			So(len(kymaInst.Status.ErrorLog), ShouldEqual, 3)
			So(kymaInst.Status.ErrorLog[0].Component, ShouldEqual, oldErrorLog1.Component)
			So(kymaInst.Status.ErrorLog[0].Log, ShouldEqual, oldErrorLog1.Log)
			So(kymaInst.Status.ErrorLog[0].Occurrences, ShouldEqual, oldErrorLog1.Occurrences)
			So(kymaInst.Status.ErrorLog[1].Component, ShouldEqual, oldErrorLog2.Component)
			So(kymaInst.Status.ErrorLog[1].Log, ShouldEqual, oldErrorLog2.Log)
			So(kymaInst.Status.ErrorLog[1].Occurrences, ShouldEqual, oldErrorLog2.Occurrences)
			So(kymaInst.Status.ErrorLog[2].Component, ShouldEqual, testComponent)
			So(kymaInst.Status.ErrorLog[2].Log, ShouldEqual, testLog)
			So(kymaInst.Status.ErrorLog[2].Occurrences, ShouldEqual, 1)
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
			So(kymaInst.Status.ErrorLog, ShouldBeEmpty)
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
