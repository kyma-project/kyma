package kymainstallation

import (
	"github.com/kyma-project/kyma/components/installer/pkg/apis/installer/v1alpha1"
	"github.com/kyma-project/kyma/components/installer/pkg/kymahelm"
	"github.com/kyma-project/kyma/components/installer/pkg/kymasources"
	"github.com/kyma-project/kyma/components/installer/pkg/overrides"
)

// StepFactory defines contract for installation steps factory
type StepFactory interface {
	NewStep(component v1alpha1.KymaComponent) Step
}

type stepFactory struct {
	kymaPackage             kymasources.KymaPackage
	helmClient              kymahelm.ClientInterface
	legacyOverridesProvider overrides.LegacyProvider
}

// NewStep method returns instance of the step based on component details
func (sf stepFactory) NewStep(component v1alpha1.KymaComponent) Step {
	return step{
		kymaPackage:             sf.kymaPackage,
		helmClient:              sf.helmClient,
		legacyOverridesProvider: sf.legacyOverridesProvider,
		component:               component,
	}
}

// NewStepFactory returns implementation of StepFactory implementation
func NewStepFactory(kymaPackage kymasources.KymaPackage, helmClient kymahelm.ClientInterface, legacyOverridesProvider overrides.LegacyProvider) StepFactory {
	return stepFactory{
		kymaPackage:             kymaPackage,
		helmClient:              helmClient,
		legacyOverridesProvider: legacyOverridesProvider,
	}
}
