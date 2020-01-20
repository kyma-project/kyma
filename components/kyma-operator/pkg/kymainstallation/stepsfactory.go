package kymainstallation

import (
	"errors"
	"log"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymahelm"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymasources"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/overrides"
	rls "k8s.io/helm/pkg/proto/hapi/release"
)

// StepFactoryCreator knows how to create an instance of the StepFactory
type StepFactoryCreator interface {
	NewInstallStepFactory(overrides.OverrideData, kymasources.LegacyKymaSourceConfig) (StepFactory, error)
	NewUninstallStepFactory() (StepFactory, error)
}

// StepFactory defines the contract for obtaining an instance of an installation/uninstallation Step
type StepFactory interface {
	NewStep(component v1alpha1.KymaComponent) Step
}

type stepFactory struct {
	helmClient        kymahelm.ClientInterface
	installedReleases map[string]bool
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

func (sfc *stepFactoryCreator) getInstalledReleases() (map[string]bool, error) {

	installedReleases := make(map[string]bool)

	list, err := sfc.helmClient.ListReleases()
	if err != nil {
		return nil, errors.New("Helm error: " + err.Error())
	}

	if list != nil {
		log.Println("Helm releases list:")
		for _, release := range list.Releases {
			statusCode := release.Info.Status.Code
			log.Printf("%s status: %s", release.Name, statusCode)
			if statusCode == rls.Status_DEPLOYED {
				installedReleases[release.Name] = true
			}
		}
	}
	return installedReleases, nil
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

// NewStep method returns instance of the installation/upgrade step based on component details
func (isf installStepFactory) NewStep(component v1alpha1.KymaComponent) Step {
	step := step{
		helmClient: isf.helmClient,
		component:  component,
	}

	inststp := installStep{
		step:         step,
		sourceGetter: isf.sourceGetter,
		overrideData: isf.overrideData,
	}

	if isf.installedReleases[component.GetReleaseName()] {
		return upgradeStep{
			inststp,
		}
	}

	return inststp
}

// NewStep method returns instance of the uninstallation step based on component details
func (usf uninstallStepFactory) NewStep(component v1alpha1.KymaComponent) Step {
	step := step{
		helmClient: usf.helmClient,
		component:  component,
	}

	if usf.installedReleases[component.GetReleaseName()] {
		return uninstallStep{
			step,
		}
	}

	return noStep{
		step,
	}
}
