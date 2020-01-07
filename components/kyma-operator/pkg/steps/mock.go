package steps

import (
	"errors"
	"time"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/config"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymasources"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/overrides"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/client/clientset/versioned/fake"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/statusmanager"
	"k8s.io/helm/pkg/proto/hapi/release"

	installationInformers "github.com/kyma-project/kyma/components/kyma-operator/pkg/client/informers/externalversions"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymahelm"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/toolkit"

	rls "k8s.io/helm/pkg/proto/hapi/services"
)

type KymaPackageMock struct {
	kymasources.KymaPackageMock
}

func (KymaPackageMock) GetChartsDirPath() string {
	return "/kymasources/charts"
}

//MockHelmClient is a fake helm client that returns no errors
type MockHelmClient struct {
	InstallReleaseCalled            bool
	InstallReleaseWithoutWaitCalled bool
	UpgradeReleaseCalled            bool
}

//InstallRelease mocks a call to helm client's InstallRelease function
func (mhc *MockHelmClient) InstallRelease(chartdir, ns, releasename, overrides string) (*rls.InstallReleaseResponse, error) {
	mhc.InstallReleaseCalled = true
	mockResponse := &rls.InstallReleaseResponse{}
	return mockResponse, nil
}

//InstallReleaseWithoutWait mocks a call to helm client's InstallReleaseWithoutWait function
func (mhc *MockHelmClient) InstallReleaseWithoutWait(chartdir, ns, releasename, overrides string) (*rls.InstallReleaseResponse, error) {
	mhc.InstallReleaseWithoutWaitCalled = true
	mockResponse := &rls.InstallReleaseResponse{}
	return mockResponse, nil
}

//UpgradeRelease mocks a call to helm client's UpgradeRelease function
func (mhc *MockHelmClient) UpgradeRelease(chartDir, releaseName, overrides string) (*rls.UpdateReleaseResponse, error) {
	mhc.UpgradeReleaseCalled = true
	mockResponse := &rls.UpdateReleaseResponse{}
	return mockResponse, nil
}

//ListReleases mocks a call to helm client's ListRelease function
func (mhc *MockHelmClient) ListReleases() (*rls.ListReleasesResponse, error) { return nil, nil }

//ReleaseStatus mocks a call to helm client's ReleaseStatus function
func (mhc *MockHelmClient) ReleaseStatus(rname string) (string, error) { return "", nil }

//InstallReleaseFromChart mocks a call to helm client's InstallReleaseFromChart function
func (mhc *MockHelmClient) InstallReleaseFromChart(chartdir, ns, releaseName, overrides string) (*rls.InstallReleaseResponse, error) {
	return nil, nil
}

//DeleteRelease mocks a call to helm client's DeleteRelease function
func (mhc *MockHelmClient) DeleteRelease(releaseName string) (*rls.UninstallReleaseResponse, error) {
	return nil, nil
}

// PrintRelease mocks a call to helm client's PrintRelease function
func (mhc *MockHelmClient) PrintRelease(release *release.Release) {}

//MockErrorHelmClient is a fake helm client that always returns an error
type MockErrorHelmClient struct {
	InstallReleaseCalled            bool
	InstallReleaseWithoutWaitCalled bool
	UpgradeReleaseCalled            bool
	ReleaseStatusCalled             bool
}

//InstallRelease mocks a call to helm client's InstallRelease function
func (mehc *MockErrorHelmClient) InstallRelease(chartdir, ns, releasename, overrides string) (*rls.InstallReleaseResponse, error) {
	mehc.InstallReleaseCalled = true
	err := errors.New("InstallRelease test error")
	return nil, err
}

//InstallReleaseWithoutWait mocks a call to helm client's InstallReleaseWithoutWait function
func (mehc *MockErrorHelmClient) InstallReleaseWithoutWait(chartdir, ns, releasename, overrides string) (*rls.InstallReleaseResponse, error) {
	mehc.InstallReleaseWithoutWaitCalled = true
	err := errors.New("InstallReleaseWithoutWait test error")
	return nil, err
}

//UpgradeRelease mocks a call to helm client's UpgradeRelease function
func (mehc *MockErrorHelmClient) UpgradeRelease(chartDir, releaseName, overrides string) (*rls.UpdateReleaseResponse, error) {
	mehc.UpgradeReleaseCalled = true
	err := errors.New("UpgradeRelease test error")
	return nil, err
}

//ReleaseStatus mocks a call to helm client's ReleaseStatus function
func (mehc *MockErrorHelmClient) ReleaseStatus(rname string) (string, error) {
	mehc.ReleaseStatusCalled = true
	return "Release test info", nil
}

//ListReleases mocks a call to helm client's ListRelease function
func (mehc *MockErrorHelmClient) ListReleases() (*rls.ListReleasesResponse, error) { return nil, nil }

//InstallReleaseFromChart mocks a call to helm client's InstallReleaseFromChart function
func (mehc *MockErrorHelmClient) InstallReleaseFromChart(chartdir, ns, releaseName, overrides string) (*rls.InstallReleaseResponse, error) {
	return nil, nil
}

//DeleteRelease mocks a call to helm client's DeleteRelease function
func (mehc *MockErrorHelmClient) DeleteRelease(releaseName string) (*rls.UninstallReleaseResponse, error) {
	return nil, nil
}

// PrintRelease mocks a call to helm client's PrintRelease function
func (mehc *MockErrorHelmClient) PrintRelease(release *release.Release) {}

//MockCommandExecutor .
type MockCommandExecutor struct {
	TimesMockCommandExecutorCalled int
}

//RunCommand .
func (kymaCommandExecutor *MockCommandExecutor) RunCommand(execPath string, execArgs ...string) error {
	kymaCommandExecutor.TimesMockCommandExecutorCalled++
	return nil
}

//MockFailingCommandExecutor .
type MockFailingCommandExecutor struct {
	MockFailingCommandExecutorCalled bool
}

//RunCommand .
func (kymaFailingCommandExecutor *MockFailingCommandExecutor) RunCommand(execPath string, execArgs ...string) error {
	kymaFailingCommandExecutor.MockFailingCommandExecutorCalled = true
	err := errors.New("RunCommand test error")
	return err
}

func getTestSetup() (*InstallationSteps, *config.InstallationData, *MockHelmClient, *MockCommandExecutor) {

	mockCommandExecutor := &MockCommandExecutor{}
	mockHelmClient := &MockHelmClient{}
	installationData, _, kymaTestSteps := getCommonTestSetup(mockHelmClient, mockCommandExecutor)

	return kymaTestSteps, installationData, mockHelmClient, mockCommandExecutor
}

func getFailingTestSetup() (*InstallationSteps, *config.InstallationData, *MockErrorHelmClient, *MockFailingCommandExecutor) {

	mockFailingCommandExecutor := &MockFailingCommandExecutor{}
	mockErrorHelmClient := &MockErrorHelmClient{}
	installationData, _, kymaTestSteps := getCommonTestSetup(mockErrorHelmClient, mockFailingCommandExecutor)

	return kymaTestSteps, installationData, mockErrorHelmClient, mockFailingCommandExecutor
}

func getCommonTestSetup(mockHelmClient kymahelm.ClientInterface, mockCommandExecutor toolkit.CommandExecutor) (*config.InstallationData, map[string]string, *InstallationSteps) {

	fakeClient := fake.NewSimpleClientset()
	informers := installationInformers.NewSharedInformerFactory(fakeClient, time.Second*0)
	mockStatusManager := statusmanager.NewKymaStatusManager(fakeClient, informers.Installer().V1alpha1().Installations().Lister())

	kymaTestSteps := New(nil, mockStatusManager, nil, nil, nil)

	installationData, overrides := toolkit.NewInstallationDataCreator().GetData()

	return &installationData, overrides, kymaTestSteps
}

type MockOverrideData struct {
	common       overrides.Map
	forComponent map[string](overrides.Map)
}

func (mod MockOverrideData) Common() overrides.Map {
	if mod.common == nil {
		return overrides.Map{}
	}
	return mod.common
}

func (mod MockOverrideData) ForComponent(componentName string) overrides.Map {

	if mod.forComponent == nil {
		return overrides.Map{}
	}
	res := mod.forComponent[componentName]
	if res == nil {
		return overrides.Map{}
	}
	return res
}
