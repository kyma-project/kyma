package backupe2e

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

const resourceQuotaObjName = "kyma-default"

type namespaceControllerTest struct {
	namespaceName string
	coreClient    *kubernetes.Clientset
}

func NewNamespaceControllerTest() (namespaceControllerTest, error) {

	kubeconfig := os.Getenv("KUBECONFIG")
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)

	coreClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return namespaceControllerTest{}, err
	}

	return namespaceControllerTest{
		namespaceName: "test-ns",
		coreClient:    coreClient,
	}, nil
}

func (n namespaceControllerTest) CreateResources(_ string) {
	err := n.createTestNamespace()
	So(err, ShouldBeNil)
}

func (n namespaceControllerTest) TestResources(namespace string) {
	err := n.waitForResources()
	So(err, ShouldBeNil)
}

func (n namespaceControllerTest) createTestNamespace() error {

	testNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   n.namespaceName,
			Labels: map[string]string{"env": "true"},
		},
	}

	_, err := n.coreClient.CoreV1().Namespaces().Create(testNamespace)
	return err
}

func (n namespaceControllerTest) waitForResources() error {

	timeout := time.After(10 * time.Second)
	tick := time.Tick(2 * time.Second)

	for {
		select {
		case <-tick:
			testNamespace, err := n.coreClient.CoreV1().Namespaces().Get(n.namespaceName, metav1.GetOptions{})
			if err != nil {
				continue
			}

			if testNamespace.Status.Phase != corev1.NamespaceActive {
				continue
			}

			_, err = n.coreClient.CoreV1().ResourceQuotas(n.namespaceName).Get(resourceQuotaObjName, metav1.GetOptions{})
			if err != nil {
				continue
			}

			return nil

		case <-timeout:
			return fmt.Errorf("resources could not be found: %v, %v", n.namespaceName, resourceQuotaObjName)}
	}
}
