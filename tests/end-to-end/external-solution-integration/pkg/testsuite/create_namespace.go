package testsuite

import (
	"fmt"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/helpers"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
)

// CreateNamespace is a step which creates new Namespace
type CreateNamespace struct {
	namespaces corev1client.NamespaceInterface
	name       string
}

var _ step.Step = &CreateNamespace{}

// NewCreateApplication returns new CreateApplication
func NewCreateNamespace(name string, namespaces corev1client.NamespaceInterface) *CreateNamespace {
	return &CreateNamespace{
		namespaces: namespaces,
		name:       name,
	}
}

// Name returns name name of the step
func (s *CreateNamespace) Name() string {
	return fmt.Sprintf("Create namespace %s", s.name)
}

// Run executes the step
func (s *CreateNamespace) Run() error {
	_, err := s.namespaces.Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: s.name,
		},
	})
	return err
}

// Cleanup removes all resources that may possibly created by the step
func (s *CreateNamespace) Cleanup() error {
	err := s.namespaces.Delete(s.name, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return helpers.AwaitResourceDeleted(func() (interface{}, error) {
		return s.namespaces.Get(s.name, metav1.GetOptions{})
	})
}
