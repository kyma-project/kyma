package kymaoperation

import (
	"strings"
	"testing"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/config"
	installationConfig "github.com/kyma-project/kyma/components/kyma-operator/pkg/config"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/errors"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymaoperation/steps"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/overrides"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/statusmanager"
	. "github.com/smartystreets/goconvey/convey"
)

func TestExecutor(t *testing.T) {
	Convey("Executor.InstallKyma()", t, func() {
		Convey("should proccess all components in a happy path scenario", func() {
			//given
			msm := mockStatusManager{}
			msp := mockStepsProvider{}
			mam := mockActionManager{}
			meh := errors.ErrorHandlers{}

			extr := Executor{
				statusManager: &msm,
				stepsProvider: msp,
				actionManager: mam,
				errorHandlers: &meh,
			}

			//when
			err := extr.InstallKyma(fakeInstallationData(), fakeOverrides{})

			//then
			So(err, ShouldBeNil)
			So(msm.componentsDone, ShouldHaveLength, 3)
			So(msm.componentsDone[0], ShouldEqual, "componentOne")
			So(msm.componentsDone[1], ShouldEqual, "componentTwo")
			So(msm.componentsDone[2], ShouldEqual, "componentThree")
		})
	})
	Convey("Executor.UninstallKyma()", t, func() {
		Convey("should proccess all components in a happy path scenario", func() {
			//given
			msm := mockStatusManager{}
			msp := mockStepsProvider{}
			mam := mockActionManager{}
			meh := errors.ErrorHandlers{}
			darf := func(extr *Executor, config *DeprovisionConfig, installation installationConfig.InstallationContext) error {
				return nil
			}

			extr := Executor{
				statusManager:               &msm,
				stepsProvider:               msp,
				actionManager:               mam,
				errorHandlers:               &meh,
				deprovisionAzureResourcesFn: darf,
			}

			//when
			err := extr.UninstallKyma(fakeInstallationData())

			//then
			So(err, ShouldBeNil)
			So(msm.componentsDone, ShouldHaveLength, 3)
			So(msm.componentsDone[0], ShouldEqual, "componentOne")
			So(msm.componentsDone[1], ShouldEqual, "componentTwo")
			So(msm.componentsDone[2], ShouldEqual, "componentThree")
		})
	})
}

func fakeInstallationData() *config.InstallationData {

	return &config.InstallationData{
		Context:     config.InstallationContext{},
		KymaVersion: "",
		URL:         "",
		Components: []v1alpha1.KymaComponent{
			v1alpha1.KymaComponent{Name: "componentOne", Namespace: "testNamespaceOne"},
			v1alpha1.KymaComponent{Name: "componentTwo", Namespace: "testNamespaceTwo"},
			v1alpha1.KymaComponent{Name: "componentThree", Namespace: "testNamespaceTwo"},
		},
		Action: "install",
	}
}

type fakeOverrides struct {
}

func (fo fakeOverrides) ForRelease(releaseName string) overrides.Map {
	return overrides.Map{}
}

type mockStatusManager struct {
	statusmanager.StatusManager
	componentsDone []string
}

func (msm *mockStatusManager) InProgress(description string) error {
	const expectedPrefix = "install component "
	//TODO: weak detection of component installation - improve!
	if strings.HasPrefix(description, expectedPrefix) {
		msm.componentsDone = append(msm.componentsDone, strings.TrimPrefix(description, expectedPrefix))
	}
	return nil
}

func (msm mockStatusManager) InstallDone(url, kymaVersion string) error {
	return nil
}

func (msm mockStatusManager) UninstallDone() error {
	return nil
}

type mockStepsLister struct {
	steps.StepLister
}

func (msl mockStepsLister) StepList() (steps.StepList, error) {
	return steps.StepList{
		fakeStep{relName: "componentOne"},
		fakeStep{relName: "componentTwo"},
		fakeStep{relName: "componentThree"},
	}, nil
}

type mockStepsProvider struct {
	StepsProvider
}

func (msp mockStepsProvider) ForInstallation(*config.InstallationData, overrides.OverrideData) (steps.StepLister, error) {
	return mockStepsLister{}, nil
}

func (msp mockStepsProvider) ForUninstallation(*config.InstallationData) (steps.StepLister, error) {
	return mockStepsLister{}, nil
}

type mockActionManager struct {
}

func (mam mockActionManager) RemoveActionLabel(name string, namespace string) error {
	return nil
}

type fakeStep struct {
	steps.Step
	relName string
}

func (fs fakeStep) Run() error {
	return nil
}
func (fs fakeStep) GetReleaseName() string {
	return fs.relName
}
