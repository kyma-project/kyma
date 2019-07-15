package testsuite

import (
	acApi "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	acClient "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/consts"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/step"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CreateMapping struct {
	mappings acClient.ApplicationMappingInterface
}

var _ step.Step = &CreateMapping{}

func NewCreateMapping(mappings acClient.ApplicationMappingInterface) *CreateMapping {
	return &CreateMapping{
		mappings: mappings,
	}
}

func (s *CreateMapping) Name() string {
	return "Create mapping"
}

func (s *CreateMapping) Run() error {
	am := &acApi.ApplicationMapping{
		TypeMeta:   metav1.TypeMeta{Kind: "ApplicationMapping", APIVersion: acApi.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{Name: consts.AppName},
	}

	_, err := s.mappings.Create(am)
	return err
}

func (s *CreateMapping) Cleanup() error {
	return s.mappings.Delete(consts.AppName, &metav1.DeleteOptions{})
}
