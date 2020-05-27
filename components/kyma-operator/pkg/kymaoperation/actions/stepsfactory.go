package actions

import (
	"log"

	errors "github.com/pkg/errors"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/config"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymahelm"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/overrides"
	helm "k8s.io/helm/pkg/proto/hapi/release"
)

const (
	defaultDeleteWaitTimeSec   = 10
	defaultRollbackWaitTimeSec = 10
)

// SourceGetterLegacySupport exist only to support legacy, now deprecated, mechanism of fetching installation sources. Remove as soon as possible.
type SourceGetterLegacySupport interface {
	SourceGetterFor(kymaURL, kymaVersion string) SourceGetter
}

//TODO: Duplicated interface, see kymasources.SourceGetter
// SourceGetter defines contract for fetching component sources.
type SourceGetter interface {
	// SrcDirFor returns a local filesystem directory path to the component sources.
	SrcDirFor(component v1alpha1.KymaComponent) (string, error)
}

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

// stepFactoryCreator is used to create StepFactory instances for installation or uninstallation.
type stepFactoryCreator struct {
	helmClient          kymahelm.ClientInterface
	sourceGetterSupport SourceGetterLegacySupport
}

// NewStepFactoryCreator returns a new stepFactoryCreator instance.
func NewStepFactoryCreator(helmClient kymahelm.ClientInterface, sgls SourceGetterLegacySupport) *stepFactoryCreator {
	return &stepFactoryCreator{
		helmClient:          helmClient,
		sourceGetterSupport: sgls,
	}
}

func (sfc *stepFactoryCreator) getInstalledReleases() (map[string]kymahelm.ReleaseStatus, error) {

	existingReleases := make(map[string]kymahelm.ReleaseStatus)

	list, err := sfc.helmClient.ListReleases()
	if err != nil {
		return nil, errors.New("Helm error: " + err.Error())
	}

	if list != nil {
		log.Println("Helm releases list:")

		for _, release := range list.Releases {
			var lastDeployedRev int32

			statusCode := release.Info.Status.Code
			if statusCode == helm.Status_DEPLOYED {
				lastDeployedRev = release.Version
			} else {
				lastDeployedRev, err = sfc.helmClient.ReleaseDeployedRevision(release.Name)
				if err != nil {
					return nil, errors.New("Helm error: " + err.Error())
				}
			}

			log.Printf("%s status: %s, last deployed revision: %d", release.Name, statusCode, lastDeployedRev)
			existingReleases[release.Name] = kymahelm.ReleaseStatus{
				StatusCode:           statusCode,
				CurrentRevision:      release.Version,
				LastDeployedRevision: lastDeployedRev,
			}
		}
	}
	return existingReleases, nil
}

// ForInstallation returns implementation of StepFactory interface used to install or upgrade Kyma
func (sfc *stepFactoryCreator) ForInstallation(installationData *config.InstallationData, overrideData overrides.OverrideData) (StepLister, error) {

	installedReleases, err := sfc.getInstalledReleases()
	if err != nil {
		return nil, err
	}

	sourceGetter := sfc.sourceGetterSupport.SourceGetterFor(installationData.URL, installationData.KymaVersion)
	return &installStepFactory{
		stepFactory{sfc.helmClient, installedReleases, installationData},
		sourceGetter,
		overrideData,
	}, nil
}

// ForUninstallation returns implementation of StepFactory interface used to uninstall Kyma
func (sfc *stepFactoryCreator) ForUninstallation(installationData *config.InstallationData) (StepLister, error) {

	installedReleases, err := sfc.getInstalledReleases()
	if err != nil {
		return nil, err
	}

	return &uninstallStepFactory{
		stepFactory{sfc.helmClient, installedReleases, installationData},
	}, nil
}

// stepList iterates over configured components and returns a list of coressponding steps.
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
	}

	inststp := installStep{
		step:              step,
		sourceGetter:      isf.sourceGetter,
		overrideData:      isf.overrideData,
		deleteWaitTimeSec: defaultDeleteWaitTimeSec,
	}

	relStatus, exists := isf.installedReleases[component.GetReleaseName()]

	if exists {

		isUpgrade, err := relStatus.IsUpgradeStep()
		if err != nil {
			return nil, errors.Wrapf(err, "unable to process release %s", component.GetReleaseName())
		}

		if isUpgrade {
			return upgradeStep{
				installStep:         inststp,
				deployedRevision:    relStatus.LastDeployedRevision,
				rollbackWaitTimeSec: defaultRollbackWaitTimeSec,
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
