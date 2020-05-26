package kymainstallation

import (
	"log"

	errors "github.com/pkg/errors"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymahelm"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymasources"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/overrides"
	helm "k8s.io/helm/pkg/proto/hapi/release"
)

const (
	defaultDeleteWaitTimeSec   = 10
	defaultRollbackWaitTimeSec = 10
)

// StepFactoryCreator knows how to create an instance of the StepFactory
type StepFactoryCreator interface {
	NewInstallStepFactory(overrides.OverrideData, kymasources.LegacyKymaSourceConfig) (StepFactory, error)
	NewUninstallStepFactory() (StepFactory, error)
}

type StepList []Step

// StepFactory defines the contract for obtaining an instance of an installation/uninstallation Step
type StepFactory interface {
	GetSteps() (StepList, error)
}

type stepFactory struct {
	helmClient        kymahelm.ClientInterface
	installedReleases map[string]kymahelm.ReleaseStatus
}

// StepFactory implementation for installation operation
type installStepFactory struct {
	stepFactory
	sourceGetter kymasources.SourceGetter
	overrideData overrides.OverrideData
}

// StepFactory implementation for uninstallation operation
type uninstallStepFactory struct {
	stepFactory
}

// stepFactoryCreator is used to create StepFactory instances for installation or uninstallation.
type stepFactoryCreator struct {
	helmClient   kymahelm.ClientInterface
	kymaPackages kymasources.KymaPackages
	fsWrapper    kymasources.FilesystemWrapper
	kymaDir      string
}

// NewStepFactoryCreator returns a new StepFactoryCreator instance.
func NewStepFactoryCreator(helmClient kymahelm.ClientInterface, kymaPackages kymasources.KymaPackages, fsWrapper kymasources.FilesystemWrapper, rootDir string) StepFactoryCreator {
	return &stepFactoryCreator{
		helmClient,
		kymaPackages,
		fsWrapper,
		rootDir,
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

// NewInstallStepFactory returns implementation of StepFactory interface used to install or upgrade Kyma
func (sfc *stepFactoryCreator) NewInstallStepFactory(overrideData overrides.OverrideData, legacySourceConfig kymasources.LegacyKymaSourceConfig) (StepFactory, error) {

	installedReleases, err := sfc.getInstalledReleases()
	if err != nil {
		return nil, err
	}

	sourceGetter := kymasources.NewSourceGetterCreator(sfc.kymaPackages, sfc.fsWrapper, sfc.kymaDir).NewGetterFor(legacySourceConfig)
	return installStepFactory{
		stepFactory{sfc.helmClient, installedReleases},
		sourceGetter,
		overrideData,
	}, nil
}

// NewUninstallStepFactory returns implementation of StepFactory interface used to uninstall Kyma
func (sfc *stepFactoryCreator) NewUninstallStepFactory() (StepFactory, error) {

	installedReleases, err := sfc.getInstalledReleases()
	if err != nil {
		return nil, err
	}

	return &uninstallStepFactory{
		stepFactory{sfc.helmClient, installedReleases},
	}, nil
}

func (isf installStepFactory) GetSteps() (StepList, error) {
	//TODO: implement
	return nil, nil
}

// NewStep method returns instance of the installation/upgrade step based on component details
func (isf installStepFactory) NewStep(component v1alpha1.KymaComponent) (Step, error) {
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

func (isf uninstallStepFactory) GetSteps() (StepList, error) {
	//TODO: implement
	return nil, nil
}

// NewStep method returns instance of the uninstallation step based on component details
func (usf uninstallStepFactory) NewStep(component v1alpha1.KymaComponent) (Step, error) {
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
