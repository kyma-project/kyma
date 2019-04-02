package configurer

import (
	tester "github.com/kyma-project/kyma/tests/console-backend-service"
	v1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type NamespaceConfigurer struct {
	Name string

	coreCli *corev1.CoreV1Client
}

func NewNamespace(name string, coreCli *corev1.CoreV1Client) *NamespaceConfigurer {
	return &NamespaceConfigurer{Name: name, coreCli: coreCli}
}

func (c *NamespaceConfigurer) Create() error {
	_, err := c.coreCli.Namespaces().Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: c.Name,
			Labels: map[string]string{
				tester.TestLabelKey: tester.TestLabelValue,
			},
		},
	})

	if err != nil && !apiErrors.IsAlreadyExists(err) {
		return err
	}

	return nil
}

func (c *NamespaceConfigurer) Delete() error {
	err := c.coreCli.Namespaces().Delete(c.Name, nil)
	if err != nil && !apiErrors.IsNotFound(err) {
		return err
	}

	return nil
}
