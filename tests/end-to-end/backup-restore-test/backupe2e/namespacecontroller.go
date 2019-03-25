package backupe2e

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

const resourceQuotaObjName = "kyma-default"

type namespaceControllerTest struct {
	coreClient *kubernetes.Clientset
}

func NewNamespaceControllerTest() (namespaceControllerTest, error) {

	kubeconfig := os.Getenv("KUBECONFIG")
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)

	coreClient, err := kubernetes.NewForConfig(cfg)
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

	timeout := time.After(5 * time.Second)
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
			return fmt.Errorf("unable to fetch resourcequota:\n %v", messages)}
	}
}
