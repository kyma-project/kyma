package api_controller

import (
	"github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/backupe2e"
	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/internal/runner"
	"github.com/sirupsen/logrus"

	kubeless "github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	gateway "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/clientset/versioned"
	"k8s.io/client-go/kubernetes"
)

type test struct {
	upstream backupe2e.ApiControllerTest
}

func New(gatewayInterface gateway.Interface, coreInterface kubernetes.Interface, kubelessInterface kubeless.Interface, domainName string) (*test, error) {
	upstream, err := backupe2e.NewApiControllerTest(gatewayInterface, coreInterface, kubelessInterface, domainName)
	return &test{upstream}, err
}

func (t *test) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	return t.upstream.CreateResourcesError(namespace)
}

func (t *test) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	return t.upstream.TestResourcesError(namespace)
}

var _ runner.UpgradeTest = &test{}
