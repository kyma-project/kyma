package testsuite

import (
	"fmt"

	acApi "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	acClient "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/helpers"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateMapping is a step which creates new Mapping
type CreateMapping struct {
	mappings acClient.ApplicationMappingInterface
	name     string
}

var _ step.Step = &CreateMapping{}

// NewCreateMapping returns new CreateMapping
func NewCreateMapping(name string, mappings acClient.ApplicationMappingInterface) *CreateMapping {
	return &CreateMapping{
		mappings: mappings,
		name:     name,
	}
}

// Name returns name name of the step
func (s *CreateMapping) Name() string {
	return fmt.Sprintf("Create mapping %s", s.name)
}

// Run executes the step
func (s *CreateMapping) Run() error {
	am := &acApi.ApplicationMapping{
		TypeMeta:   metav1.TypeMeta{Kind: "ApplicationMapping", APIVersion: acApi.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{Name: s.name},
	}

	_, err := s.mappings.Create(am)
	return err
}

// Cleanup removes all resources that may possibly created by the step
func (s *CreateMapping) Cleanup() error {
	err := s.mappings.Delete(s.name, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return helpers.AwaitResourceDeleted(func() (interface{}, error) {
		return s.mappings.Get(s.name, v1.GetOptions{})
	})
}
