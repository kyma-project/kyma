package backupe2e

import (
	"time"

	"github.com/avast/retry-go"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/utils/config"
	. "github.com/smartystreets/goconvey/convey"
)

const resourceQuotaObjName = "kyma-default"

type NamespaceControllerTest struct {
	coreInterface kubernetes.Interface
}

func NewNamespaceControllerTestFromEnv() (NamespaceControllerTest, error) {
	restConfig, err := config.NewRestClientConfig()
	if err != nil {
		return NamespaceControllerTest{}, err
	}

	coreClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return NamespaceControllerTest{}, err
	}

	return NamespaceControllerTest{
		coreInterface: coreClient,
	}, nil
}

func NewNamespaceControllerTest(coreInterface kubernetes.Interface) NamespaceControllerTest {
	return NamespaceControllerTest{
		coreInterface: coreInterface,
	}
}

func (n NamespaceControllerTest) CreateResources(namespace string) {
	err := n.CreateResourcesError(namespace)
	So(err, ShouldBeNil)
}

func (n NamespaceControllerTest) CreateResourcesError(namespace string) error {
	return n.labelTestNamespace(namespace)
}

func (n NamespaceControllerTest) TestResources(namespace string) {
	err := n.TestResourcesError(namespace)
	So(err, ShouldBeNil)
}

func (n NamespaceControllerTest) TestResourcesError(namespace string) error {
	return n.waitForResourceQuota(namespace)
}

func (n NamespaceControllerTest) DeleteResources(namespace string) {
	// There is not need to be implemented for this test.
}

func (n NamespaceControllerTest) labelTestNamespace(namespaceName string) error {
	namespace, err := n.coreInterface.CoreV1().Namespaces().Get(namespaceName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	labels := namespace.GetLabels()
	labels["env"] = "true"

	namespaceCopy := namespace.DeepCopy()
	namespaceCopy.SetLabels(labels)

	_, err = n.coreInterface.CoreV1().Namespaces().Update(namespaceCopy)
	return err
}

func (n NamespaceControllerTest) waitForResourceQuota(namespaceName string) error {
	return retry.Do(func() error {
		_, err := n.coreInterface.CoreV1().ResourceQuotas(namespaceName).Get(resourceQuotaObjName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		return nil
	},
		retry.Delay(500*time.Millisecond),
	)
}
