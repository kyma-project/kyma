package testsuite

import (
	acApi "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	acClient "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/consts"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateEventActivation is a step which creates new EventActivation
type CreateEventActivation struct {
	eventActivations acClient.EventActivationInterface
}

var _ step.Step = &CreateEventActivation{}

// NewCreateEventActivation returns new CreateEventActivation
func NewCreateEventActivation(eventActivations acClient.EventActivationInterface) *CreateEventActivation {
	return &CreateEventActivation{
		eventActivations: eventActivations,
	}
}

// Name returns name name of the step
func (s *CreateEventActivation) Name() string {
	return "Create event activation"
}

// Run executes the step
func (s *CreateEventActivation) Run() error {
	eaSpec := acApi.EventActivationSpec{
		DisplayName: "Commerce-events",
		SourceID:    consts.AppName,
	}

	ea := &acApi.EventActivation{
		TypeMeta:   metav1.TypeMeta{Kind: "EventActivation", APIVersion: acApi.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{Name: consts.AppName},

		Spec: eaSpec,
	}

	_, err := s.eventActivations.Create(ea)
	return err
}

// Cleanup removes all resources that may possibly created by the step
func (s *CreateEventActivation) Cleanup() error {
	return s.eventActivations.Delete(consts.AppName, &metav1.DeleteOptions{})
}
