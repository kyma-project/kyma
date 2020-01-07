package kymainstallation

import (
	"errors"
	"log"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymahelm"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/overrides"
	rls "k8s.io/helm/pkg/proto/hapi/release"
)

// StepFactory defines contract for installation steps factory
type StepFactory interface {
	NewStep(component v1alpha1.KymaComponent) Step
}

// SourceGetter defines contract for fetching component sources
type SourceGetter interface {
	// Get returns a local directory path to the component sources.
	// If the component is configured with external `Source.URL`, it's sources are downloaded to a local directory.
	// Otherwise component sources bundled with kyma-operator Docker image are used.
	Get(component v1alpha1.KymaComponent) (string, error)
}

type stepFactory struct {
	helmClient        kymahelm.ClientInterface
	installedReleases map[string]bool
}

type installStepFactory struct {
	stepFactory
	sourceGetter SourceGetter
	overrideData overrides.OverrideData
}

type uninstallStepFactory struct {
	stepFactory
}

// NewStep method returns instance of the installation/upgrade step based on component details
func (isf installStepFactory) NewStep(component v1alpha1.KymaComponent) Step {
	step := step{
		helmClient:   isf.helmClient,
		sourceGetter: isf.sourceGetter,
		overrideData: isf.overrideData,
		component:    component,
	}

	if isf.installedReleases[component.GetReleaseName()] {
		return upgradeStep{
			step,
		}
	}

	return installStep{
		step,
	}
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

// NewInstallStepFactory returns implementation of StepFactory interface used to install or upgrade Kyma
func NewInstallStepFactory(sourceGetter SourceGetter, helmClient kymahelm.ClientInterface, overrideData overrides.OverrideData) (StepFactory, error) {

	stepFactory, err := newStepsFactory(helmClient)
	if err != nil {
		return nil, err
	}

	return installStepFactory{
		*stepFactory,
		sourceGetter,
		overrideData,
	}, nil
}

// NewUninstallStepFactory returns implementation of StepFactory interface used to uninstall Kyma
func NewUninstallStepFactory(helmClient kymahelm.ClientInterface) (StepFactory, error) {

	stepFactory, err := newStepsFactory(helmClient)
	if err != nil {
		return nil, err
	}

	return uninstallStepFactory{
		*stepFactory,
	}, nil
}

func newStepsFactory(helmClient kymahelm.ClientInterface) (*stepFactory, error) {

	installedReleases := make(map[string]bool)

	list, err := helmClient.ListReleases()
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

	return &stepFactory{
		helmClient:        helmClient,
		installedReleases: installedReleases,
	}, nil
}
