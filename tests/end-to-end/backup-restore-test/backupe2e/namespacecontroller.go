package backupe2e

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"

	. "github.com/smartystreets/goconvey/convey"
)

type namespaceControllerTest struct {
	namespaceName string
	coreClient    *kubernetes.Clientset
}

func newNamespaceControllerTest() (namespaceControllerTest, error) {

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
	//todo: implement
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
