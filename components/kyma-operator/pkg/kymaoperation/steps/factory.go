package steps

import (
	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/config"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymahelm"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/overrides"
)

//StepList is a list of steps corresponding to a list of components as defined by Installation CR
type StepList []Step

type StepLister interface {
	StepList() (StepList, error)
}

type stepFactory struct {
	helmClient        kymahelm.ClientInterface
	installedReleases map[string]kymahelm.ReleaseStatus
	installationData  *config.InstallationData
}

// StepFactory implementation for installation operation
type installStepFactory struct {
	stepFactory
	sourceGetter SourceGetter
	overrideData overrides.OverrideData
}

// StepFactory implementation for uninstallation operation
type uninstallStepFactory struct {
	stepFactory
}

// stepList iterates over configured components and returns a list of corresponding steps.
func (sf stepFactory) stepList(newStepFn func(component v1alpha1.KymaComponent) (Step, error)) (StepList, error) {
	res := StepList{}
	for _, component := range sf.installationData.Components {

		step, err := newStepFn(component)
		if err != nil {
			return nil, err
		}

		res = append(res, step)
	}

	return res, nil
}

func (isf installStepFactory) StepList() (StepList, error) {
	return isf.stepFactory.stepList(isf.newStep)
}

// newStep method returns instance of the installation/upgrade step based on component details
func (isf installStepFactory) newStep(component v1alpha1.KymaComponent) (Step, error) {
	step := step{
		helmClient: isf.helmClient,
		component:  component,
		profile:    isf.stepFactory.installationData.Profile,
	}

	inststp := installStep{
		step:         step,
		sourceGetter: isf.sourceGetter,
		overrideData: isf.overrideData,
	}

	relStatus, exists := isf.installedReleases[component.GetReleaseName()]

	if exists {

		isUpgrade, err := relStatus.IsUpgradeStep()
		if err != nil {
			return nil, errors.Wrapf(err, "unable to process release %s", component.GetReleaseName())
		}

		if isUpgrade {
			return upgradeStep{
				installStep: inststp,
			}, nil
		}
	}

	return inststp, nil
}

func (usf uninstallStepFactory) StepList() (StepList, error) {
	return usf.stepFactory.stepList(usf.newStep)
}

// NewStep method returns instance of the uninstallation step based on component details
func (usf uninstallStepFactory) newStep(component v1alpha1.KymaComponent) (Step, error) {
	step := step{
		helmClient: usf.helmClient,
		component:  component,
	}

	_, exists := usf.installedReleases[component.GetReleaseName()]

	if exists {
		return uninstallStep{
			step,
		}, nil
	}

	return noStep{
		step,
	}, nil
}
