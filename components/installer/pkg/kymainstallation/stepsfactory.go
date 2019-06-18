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
	chartsDirPath     string
	helmClient        kymahelm.ClientInterface
	installedReleases map[string]bool
	overrideData      overrides.OverrideData
}

// NewStep method returns instance of the step based on component details
func (sf stepFactory) NewStep(component v1alpha1.KymaComponent) Step {
	step := step{
		chartsDirPath: sf.chartsDirPath,
		helmClient:    sf.helmClient,
		overrideData:  sf.overrideData,
		component:     component,
	}

	if sf.installedReleases[component.GetReleaseName()] {
		return upgradeStep{
			step: step,
		}
	}

	return installStep{
		step: step,
	}
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

	return stepFactory{
		chartsDirPath:     chartsDirPath,
		helmClient:        helmClient,
		installedReleases: installedReleases,
		overrideData:      overrideData,
	}, nil
}
