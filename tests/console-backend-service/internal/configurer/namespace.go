package configurer

import (
	tester "github.com/kyma-project/kyma/tests/console-backend-service"
	v1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type NamespaceConfigurer struct {
	NamePrefix    string
	generatedName *string
	coreCli       *corev1.CoreV1Client
}

func NewNamespace(name string, coreCli *corev1.CoreV1Client) *NamespaceConfigurer {
	return &NamespaceConfigurer{NamePrefix: name, coreCli: coreCli}
}

func (c *NamespaceConfigurer) Create() error {
	namespace, err := c.coreCli.Namespaces().Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: c.NamePrefix,
			Labels: map[string]string{
				tester.TestLabelKey: tester.TestLabelValue,
			},
		},
	})

	if err != nil && !apiErrors.IsAlreadyExists(err) {
		return err
	}

	c.generatedName = &namespace.Name
	return nil
}

func (c *NamespaceConfigurer) Delete() error {
	if c.generatedName == nil {
		return nil
	}

	err := c.coreCli.Namespaces().Delete(*c.generatedName, nil)
	if err != nil && !apiErrors.IsNotFound(err) {
		return err
	}

	return nil
}

func (c *NamespaceConfigurer) Name() string {
	return *c.generatedName
}
