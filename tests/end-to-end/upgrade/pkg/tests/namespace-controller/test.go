package namespacecontroller

import (
	"github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/informers/externalversions/gateway.kyma-project.io"
	"github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/backupe2e"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

// Test tests the API controller business logic after Kyma upgrade phase
type Test struct {
	upstream backupe2e
}

// New creates new instance of Test
func New(gatewayInterface gateway.Interface, coreInterface kubernetes.Interface, kubelessInterface kubeless.Interface, domainName string) Test {
	upstream := backupe2e.NewApiControllerTest(gatewayInterface, coreInterface, kubelessInterface, domainName)
	return Test{upstream}
}

// CreateResources creates resources for tests
func (t Test) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	return t.upstream.CreateResourcesError(namespace)
}

// TestResources tests if resources are working properly after upgrade
func (t Test) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	return t.upstream.TestResourcesError(namespace)
}
