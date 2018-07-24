package steps

import (
	"errors"
	"flag"
	"time"

	"github.com/kyma-project/kyma/components/installer/pkg/config"

	"github.com/kyma-project/kyma/components/installer/pkg/client/clientset/versioned/fake"
	"github.com/kyma-project/kyma/components/installer/pkg/statusmanager"
	"k8s.io/helm/pkg/proto/hapi/release"

	installationInformers "github.com/kyma-project/kyma/components/installer/pkg/client/informers/externalversions"
	"github.com/kyma-project/kyma/components/installer/pkg/kymahelm"
	"github.com/kyma-project/kyma/components/installer/pkg/toolkit"

	rls "k8s.io/helm/pkg/proto/hapi/services"
)

//TestChartDir is a mock directory for tests
var TestChartDir = flag.String("testchartdir", "./test-kyma", "Test chart directory")

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
	TimesMockCommandExecutorCalled     int
	TimesMockBashCommandExecutorCalled int
}

//RunCommand .
func (kymaCommandExecutor *MockCommandExecutor) RunCommand(execPath string, execArgs ...string) error {
	kymaCommandExecutor.TimesMockCommandExecutorCalled++
	return nil
}

//RunBashCommand .
func (kymaCommandExecutor *MockCommandExecutor) RunBashCommand(scriptPath string, execArgs ...string) error {
	kymaCommandExecutor.TimesMockBashCommandExecutorCalled++
	return nil
}

//MockFailingCommandExecutor .
type MockFailingCommandExecutor struct {
	MockFailingCommandExecutorCalled     bool
	MockFailingBashCommandExecutorCalled bool
}

//RunCommand .
func (kymaFailingCommandExecutor *MockFailingCommandExecutor) RunCommand(execPath string, execArgs ...string) error {
	kymaFailingCommandExecutor.MockFailingCommandExecutorCalled = true
	err := errors.New("RunCommand test error")
	return err
}

//RunBashCommand .
func (kymaFailingCommandExecutor *MockFailingCommandExecutor) RunBashCommand(scriptPath string, execArgs ...string) error {
	kymaFailingCommandExecutor.MockFailingBashCommandExecutorCalled = true
	err := errors.New("RunBashCommand test error")
	return err
}

func getTestSetup() (*InstallationSteps, *config.InstallationData, *MockHelmClient, *MockCommandExecutor) {

	mockCommandExecutor := &MockCommandExecutor{}
	mockHelmClient := &MockHelmClient{}
	installationData, kymaTestSteps := getCommonTestSetup(mockHelmClient, mockCommandExecutor)

	return kymaTestSteps, installationData, mockHelmClient, mockCommandExecutor
}

func getFailingTestSetup() (*InstallationSteps, *config.InstallationData, *MockErrorHelmClient, *MockFailingCommandExecutor) {

	mockFailingCommandExecutor := &MockFailingCommandExecutor{}
	mockErrorHelmClient := &MockErrorHelmClient{}
	installationData, kymaTestSteps := getCommonTestSetup(mockErrorHelmClient, mockFailingCommandExecutor)

	return kymaTestSteps, installationData, mockErrorHelmClient, mockFailingCommandExecutor
}

func getCommonTestSetup(mockHelmClient kymahelm.ClientInterface, mockCommandExecutor toolkit.CommandExecutor) (*config.InstallationData, *InstallationSteps) {

	fakeClient := fake.NewSimpleClientset()
	informers := installationInformers.NewSharedInformerFactory(fakeClient, time.Second*0)
	mockStatusManager := statusmanager.NewKymaStatusManager(fakeClient, informers.Installer().V1alpha1().Installations().Lister())

	installationData := toolkit.NewInstallationDataCreator().GetData()
	kymaTestSteps := New(mockHelmClient, nil, nil, *TestChartDir, mockStatusManager, nil, mockCommandExecutor, nil)

	return &installationData, kymaTestSteps
}
