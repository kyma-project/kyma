package testsuite

import (
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/helpers"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreClient "k8s.io/client-go/kubernetes/typed/core/v1"
)

// CreateNamespace is a step which creates new Namespace
type CreateNamespace struct {
	namespaces coreClient.NamespaceInterface
	name       string
}

var _ step.Step = &CreateNamespace{}

// NewCreateApplication returns new CreateApplication
func NewCreateNamespace(namespaces coreClient.NamespaceInterface, name string) *CreateNamespace {
	return &CreateNamespace{
		namespaces: namespaces,
		name:       name,
	}
}

// Name returns name name of the step
func (s *CreateNamespace) Name() string {
	return "Create namespace"
}

// Run executes the step
func (s *CreateNamespace) Run() error {
	_, err := s.namespaces.Create(&core.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: s.name,
		},
	})
	return err
}

// Cleanup removes all resources that may possibly created by the step
func (s *CreateNamespace) Cleanup() error {
	err := s.namespaces.Delete(s.name, &v1.DeleteOptions{})
	if err != nil {
		return err
	}

	return helpers.AwaitResourceDeleted(func() (interface{}, error) {
		return s.namespaces.Get(s.name, v1.GetOptions{})
	})
}
