package kymainstallation

import (
	"errors"
	"log"

	"github.com/kyma-project/kyma/components/installer/pkg/apis/installer/v1alpha1"
	"github.com/kyma-project/kyma/components/installer/pkg/kymahelm"
	"github.com/kyma-project/kyma/components/installer/pkg/overrides"
	rls "k8s.io/helm/pkg/proto/hapi/release"
)

// StepFactory defines contract for installation steps factory
type StepFactory interface {
	NewStep(component v1alpha1.KymaComponent) Step
}

type stepFactory struct {
	helmClient        kymahelm.ClientInterface
	installedReleases map[string]bool
}

type installStepFactory struct {
	stepFactory
	chartsDirPath string
	overrideData  overrides.OverrideData
}

type uninstallStepFactory struct {
	stepFactory
}

// NewStep method returns instance of the step based on component details
func (isf installStepFactory) NewStep(component v1alpha1.KymaComponent) Step {
	step := step{
		chartsDirPath: isf.chartsDirPath,
		helmClient:    isf.helmClient,
		overrideData:  isf.overrideData,
		component:     component,
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

	return nil
}

// NewStepFactory returns implementation of StepFactory implementation
func NewStepFactory(chartsDirPath string, helmClient kymahelm.ClientInterface, overrideData overrides.OverrideData) (StepFactory, error) {
	installedReleases := make(map[string]bool)

	relesesRes, err := helmClient.ListReleases()
	if err != nil {
		return nil, errors.New("Helm error: " + err.Error())
	}

	if relesesRes != nil {
		log.Println("Helm releases list:")
		for _, release := range relesesRes.Releases {
			statusCode := release.Info.Status.Code
			log.Printf("%s status: %s", release.Name, statusCode)
			if statusCode == rls.Status_DEPLOYED {
				installedReleases[release.Name] = true
			}
		}
	}

	sf := stepFactory{
		helmClient:        helmClient,
		installedReleases: installedReleases,
	}

	if chartsDirPath != "" && overrideData != nil {
		return installStepFactory{
			sf,
			chartsDirPath,
			overrideData,
		}, nil
	}

	return uninstallStepFactory{
		sf,
	}, nil
}
