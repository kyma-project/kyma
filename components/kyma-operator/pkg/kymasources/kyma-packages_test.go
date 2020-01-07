package kymasources

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestHasBundledSources(t *testing.T) {
	Convey("HasBundledSources", t, func() {
		Convey("returns true when sources are injected", func() {
			fsWrapperMock := &fsWrapperMockedForExistingSources{}
			kymaPackages := NewKymaPackages(fsWrapperMock, nil, "/kyma")

			sourcesAreBundled := kymaPackages.HasBundledSources()

			So(sourcesAreBundled, ShouldEqual, true)
		})

		Convey("returns false when sources are not injected", func() {
			fsWrapperMock := &fsWrapperMockedForNotExistingSources{}
			kymaPackages := NewKymaPackages(fsWrapperMock, nil, "/kyma")

			sourcesAreBundled := kymaPackages.HasBundledSources()

			So(sourcesAreBundled, ShouldEqual, false)
		})
	})

	Convey("GetBundledPackage", t, func() {
		Convey("returns KymaPackage instance for injected package", func() {
			fsWrapperMock := &fsWrapperMockedForExistingSources{}
			kymaPackages := NewKymaPackages(fsWrapperMock, nil, "/kyma")

			injectedPackage, err := kymaPackages.GetBundledPackage()

			So(err, ShouldBeNil)
			So(injectedPackage, ShouldNotBeNil)
		})

		Convey("returns error when there is no injected package", func() {
			fsWrapperMock := &fsWrapperMockedForNotExistingSources{}
			kymaPackages := NewKymaPackages(fsWrapperMock, nil, "/kyma")

			injectedPackage, err := kymaPackages.GetBundledPackage()

			So(err, ShouldNotBeNil)
			So(injectedPackage, ShouldBeNil)
		})
	})

	Convey("GetPackage", t, func() {
		Convey("returns KymaPackage instance when package exist", func() {
			fsWrapperMocked := &fsWrapperMockedForExistingSources{}
			kymaPackages := NewKymaPackages(fsWrapperMocked, nil, "/kyma")

			kymaPackage, err := kymaPackages.GetPackage("v1.0.0")

			So(err, ShouldBeNil)
			So(kymaPackage, ShouldNotBeNil)
		})

		Convey("returns error when the package does not exist", func() {
			fsWrapperMocked := &fsWrapperMockedForNotExistingSources{}
			kymaPackages := NewKymaPackages(fsWrapperMocked, nil, "/kyma")

			kymaPackage, err := kymaPackages.GetPackage("v1.0.0")

			So(err, ShouldNotBeNil)
			So(kymaPackage, ShouldBeNil)
		})
	})

	Convey("FetchPackage", t, func() {
		Convey("fetches the package from server and extracts it", func() {
			cmdExecutorMock := newKymaCommandExecutor().
				pushCommand("curl", "-Lks", "http://domain.local/package.tar.gz", "-o", "/kyma/legacy-deprecated/v1.0.0.tar.gz").
				pushCommand("tar", "xz", "-C", "/kyma/legacy-deprecated/v1.0.0", "--strip-components=1", "-f", "/kyma/legacy-deprecated/v1.0.0.tar.gz")
			fsWrapperMock := &fsWrapperForCreatingDir{}
			kymaPackages := NewKymaPackages(fsWrapperMock, cmdExecutorMock, "/kyma")

			err := kymaPackages.FetchPackage("http://domain.local/package.tar.gz", "v1.0.0")

			So(err, ShouldBeNil)
		})
	})
}

type fsWrapperMockedForExistingSources struct {
	FilesystemWrapperMock
}

func (fsWrapperMockedForExistingSources) Exists(path string) bool { return true }

type fsWrapperMockedForNotExistingSources struct {
	FilesystemWrapper
}

func (fsWrapperMockedForNotExistingSources) Exists(path string) bool { return false }

type fsWrapperForCreatingDir struct {
	FilesystemWrapperMock
}

func (fsWrapperForCreatingDir) Exists(path string) bool               { return false }
func (fsWrapperForCreatingDir) CreateDir(packageDirPath string) error { return nil }

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

func (kymaCommandExecutor *mockCommandExecutor) expectNoCalls() *mockCommandExecutor {
	kymaCommandExecutor.noCallsExpected = true
	return kymaCommandExecutor
}

type command struct {
	execPath string
	execArgs []string
}
