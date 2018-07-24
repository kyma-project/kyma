package steps

import (
	"strings"
	"testing"
	"time"

	fake "github.com/kyma-project/kyma/components/installer/pkg/client/clientset/versioned/fake"
	installationInformers "github.com/kyma-project/kyma/components/installer/pkg/client/informers/externalversions"
	"github.com/kyma-project/kyma/components/installer/pkg/config"
	statusmanager "github.com/kyma-project/kyma/components/installer/pkg/statusmanager"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDownloadKyma(t *testing.T) {

	Convey("DownloadKyma function", t, func() {

		Convey("should download kyma package in case of remote installation", func() {
			testInst := &config.InstallationData{
				URL:         "doesnotexists",
				KymaVersion: "test",
			}

			fakeClient := fake.NewSimpleClientset()
			informers := installationInformers.NewSharedInformerFactory(fakeClient, time.Second*0)
			mockStatusManager := statusmanager.NewKymaStatusManager(fakeClient, informers.Installer().V1alpha1().Installations().Lister())

			mockCommandExecutor := newKymaCommandExecutor().
				pushCommand("curl", "-Lks", testInst.URL, "-o", testInst.KymaVersion+".tar.gz").
				pushCommand("tar", "xz", "-C", "./test-kyma", "--strip-components=1", "-f", testInst.KymaVersion+".tar.gz")

			mockKymaPackage := &mockKymaPackageClientForDownload{}
			kymaTestSteps := New(nil, nil, nil, *TestChartDir, mockStatusManager, nil, mockCommandExecutor, mockKymaPackage)

			err := kymaTestSteps.DownloadKyma(testInst)

			So(mockCommandExecutor.TimesMockCommandExecutorCalled, ShouldEqual, 2)
			So(err, ShouldBeNil)
		})

		Convey("should not create directory nor download kyma package in case of local installation", func() {

			testInst := &config.InstallationData{}

			fakeClient := fake.NewSimpleClientset()
			informers := installationInformers.NewSharedInformerFactory(fakeClient, time.Second*0)
			mockStatusManager := statusmanager.NewKymaStatusManager(fakeClient, informers.Installer().V1alpha1().Installations().Lister())

			mockCommandExecutor := newKymaCommandExecutor().
				expectNoCalls()

			mockKymaPackage := &mockKymaPackageClientForLocal{}

			kymaTestSteps := New(nil, nil, nil, *TestChartDir, mockStatusManager, nil, mockCommandExecutor, mockKymaPackage)

			err := kymaTestSteps.DownloadKyma(testInst)

			So(mockCommandExecutor.TimesMockCommandExecutorCalled, ShouldEqual, 0)
			So(err, ShouldBeNil)
		})

		Convey("should return error if url is not set", func() {
			testInst := &config.InstallationData{}

			fakeClient := fake.NewSimpleClientset()
			informers := installationInformers.NewSharedInformerFactory(fakeClient, time.Second*0)
			mockStatusManager := statusmanager.NewKymaStatusManager(fakeClient, informers.Installer().V1alpha1().Installations().Lister())

			mockCommandExecutor := newKymaCommandExecutor().
				expectNoCalls()

			mockKymaPackage := &mockKymaPackageClientForDownload{}
			kymaTestSteps := New(nil, nil, nil, *TestChartDir, mockStatusManager, nil, mockCommandExecutor, mockKymaPackage)

			err := kymaTestSteps.DownloadKyma(testInst)

			So(mockCommandExecutor.TimesMockCommandExecutorCalled, ShouldEqual, 0)
			So(err, ShouldNotBeNil)
		})

	})
}

type mockKymaPackageClientForLocal struct {
}

// NeedDownload .
func (kymaPackageClient *mockKymaPackageClientForLocal) NeedDownload(kymaPath string) bool {
	So(kymaPath, ShouldEqual, "./test-kyma")

	return false
}

// CreateDir .
func (kymaPackageClient *mockKymaPackageClientForLocal) CreateDir(kymaPath string) error {
	panic("CreateDir shouldn't be called!")
}

// RemoveDir .
func (kymaPackageClient *mockKymaPackageClientForLocal) RemoveDir(kymaPath string) error {
	So(kymaPath, ShouldEqual, "./test-kyma")

	return nil
}

type mockKymaPackageClientForDownload struct {
}

// NeedDownload .
func (kymaPackageClient *mockKymaPackageClientForDownload) NeedDownload(kymaPath string) bool {
	So(kymaPath, ShouldEqual, "./test-kyma")

	return true
}

// CreateDir .
func (kymaPackageClient *mockKymaPackageClientForDownload) CreateDir(kymaPath string) error {
	So(kymaPath, ShouldEqual, "./test-kyma")

	return nil
}

// RemoveDir .
func (kymaPackageClient *mockKymaPackageClientForDownload) RemoveDir(kymaPath string) error {
	So(kymaPath, ShouldEqual, "./test-kyma")

	return nil
}

type mockCommandExecutor struct {
	TimesMockCommandExecutorCalled int
	commands                       []command
	noCallsExpected                bool
}

func newKymaCommandExecutor() *mockCommandExecutor {
	return &mockCommandExecutor{}
}

func (kymaCommandExecutor *mockCommandExecutor) pushCommand(execPath string, execArgs ...string) *mockCommandExecutor {
	command := &command{
		execPath: execPath,
		execArgs: execArgs,
	}
	kymaCommandExecutor.commands = append(kymaCommandExecutor.commands, *command)
	return kymaCommandExecutor
}

func (kymaCommandExecutor *mockCommandExecutor) popCommand() command {
	var cmd command
	cmd, kymaCommandExecutor.commands = kymaCommandExecutor.commands[0], kymaCommandExecutor.commands[1:]

	return cmd
}

//RunCommand .
func (kymaCommandExecutor *mockCommandExecutor) RunCommand(execPath string, execArgs ...string) error {
	if kymaCommandExecutor.noCallsExpected {
		panic("Shouldn't be called!")
	}

	cmd := kymaCommandExecutor.popCommand()

	So(execPath, ShouldEqual, cmd.execPath)
	So(strings.Join(execArgs, ""), ShouldEqual, strings.Join(cmd.execArgs, ""))

	kymaCommandExecutor.TimesMockCommandExecutorCalled++

	return nil
}

//RunBashCommand .
func (kymaCommandExecutor *mockCommandExecutor) RunBashCommand(execPath string, execArgs ...string) error {
	panic("RunBashCommand shouldn't be called!")
}

func (kymaCommandExecutor *mockCommandExecutor) expectNoCalls() *mockCommandExecutor {
	kymaCommandExecutor.noCallsExpected = true
	return kymaCommandExecutor
}

type command struct {
	execPath string
	execArgs []string
}
