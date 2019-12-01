package apicontroller

import (
	"github.com/michal-hudy/kyma/tests/end-to-end/backup-restore-test/backupe2e"
	dex "github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/utils/fetch-dex-token"
	"github.com/sirupsen/logrus"

	kubeless "github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	gateway "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/clientset/versioned"
	"k8s.io/client-go/kubernetes"
)

// Test tests the API controller business logic after Kyma upgrade phase
type Test struct {
	upstream backupe2e.ApiControllerTest
}

// New creates new instance of Test
func New(gatewayInterface gateway.Interface, coreInterface kubernetes.Interface, kubelessInterface kubeless.Interface, domainName string, dexConfig dex.IdProviderConfig) Test {
	upstream := backupe2e.NewApiControllerTest(gatewayInterface, coreInterface, kubelessInterface, domainName, dexConfig)
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
