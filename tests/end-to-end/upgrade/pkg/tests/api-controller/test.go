package apicontroller

import (
	"github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/backupe2e"
	"github.com/sirupsen/logrus"

	kubeless "github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	gateway "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/clientset/versioned"
	"k8s.io/client-go/kubernetes"
)

type Test struct {
	upstream backupe2e.ApiControllerTest
}

func New(gatewayInterface gateway.Interface, coreInterface kubernetes.Interface, kubelessInterface kubeless.Interface, domainName string) Test {
	upstream := backupe2e.NewApiControllerTest(gatewayInterface, coreInterface, kubelessInterface, domainName)
	return Test{upstream}
}

func (t Test) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	return t.upstream.CreateResourcesError(namespace)
}

func (t Test) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	return t.upstream.TestResourcesError(namespace)
}
