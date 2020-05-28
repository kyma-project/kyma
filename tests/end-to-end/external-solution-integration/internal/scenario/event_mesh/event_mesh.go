package event_mesh

import (
	"k8s.io/client-go/rest"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario/event_mesh_evaluate"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario/event_mesh_prepare"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
)

// Steps return scenario steps
func (s *Scenario) Steps(config *rest.Config) ([]step.Step, error) {
	s.prepare = event_mesh_prepare.Scenario{
		Domain:            s.domain,
		TestID:            s.testID,
		SkipSSLVerify:     s.skipSSLVerify,
		ApplicationTenant: s.applicationTenant,
		ApplicationGroup:  s.applicationGroup,
		TestServiceImage:  s.testServiceImage,
	}
	s.evaluate = event_mesh_evaluate.Scenario{
		Domain:        s.domain,
		TestID:        s.testID,
		SkipSSLVerify: s.skipSSLVerify,
	}

	prepareSteps, err := s.prepare.Steps(config)
	if err != nil {
		return nil, err
	}
	evalSteps, err := s.evaluate.Steps(config)
	if err != nil {
		return nil, err
	}
	return append(prepareSteps, evalSteps...), nil
}
