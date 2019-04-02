package backupe2e

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/utils/config"
	. "github.com/smartystreets/goconvey/convey"
)

const resourceQuotaObjName = "kyma-default"

type namespaceControllerTest struct {
	coreClient *kubernetes.Clientset
}

func NewNamespaceControllerTest() (namespaceControllerTest, error) {
	restConfig, err := config.NewRestClientConfig()
	if err != nil {
		return namespaceControllerTest{}, err
	}

	coreClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return namespaceControllerTest{}, err
	}

	return namespaceControllerTest{
		coreClient: coreClient,
	}, nil
}

func (n namespaceControllerTest) CreateResources(namespace string) {
	err := n.labelTestNamespace(namespace)
	So(err, ShouldBeNil)
}

func (n namespaceControllerTest) TestResources(namespace string) {
	err := n.waitForResourceQuota(namespace)
	So(err, ShouldBeNil)
}

func (n namespaceControllerTest) DeleteResources(namespace string) {
	// There is not need to be implemented for this test.
}

func (n namespaceControllerTest) labelTestNamespace(namespaceName string) error {
	namespace, err := n.coreClient.CoreV1().Namespaces().Get(namespaceName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	labels := namespace.GetLabels()
	labels["env"] = "true"

	namespaceCopy := namespace.DeepCopy()
	namespaceCopy.SetLabels(labels)

	_, err = n.coreClient.CoreV1().Namespaces().Update(namespaceCopy)
	return err
}

func (n namespaceControllerTest) waitForResourceQuota(namespaceName string) error {
	timeout := time.After(5 * time.Minute)
	tick := time.Tick(1 * time.Second)

	var messages string

	for {
		select {
		case <-tick:
			_, err := n.coreClient.CoreV1().ResourceQuotas(namespaceName).Get(resourceQuotaObjName, metav1.GetOptions{})
			if err != nil {
				messages += fmt.Sprintf("%+v\n", err)
				continue
			}

			return nil

		case <-timeout:
			return fmt.Errorf("unable to fetch resourcequota:\n %v", messages)
		}
	}
}
