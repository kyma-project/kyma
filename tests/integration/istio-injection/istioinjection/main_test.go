package istioinjection

import (
	"fmt"
	"testing"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

const (
	namespaceNameRoot = "istio-injection-tests"
	noSidecar         = 1
	hasSidecar        = 2
)

type TestSuite struct {
	k8sClient *kubernetes.Clientset
	namespace string
}

func TestIstioInjection(t *testing.T) {
	namespace := fmt.Sprintf("%s-%s", namespaceNameRoot, generateRandomString(8))
	if namespace == "" {
		log.Fatal("Namespace not set.")
	}

	testSuite := &TestSuite{namespace: namespace}
	testSuite.initK8sClient()

	defer testSuite.deleteNamespace()
	if err := testSuite.createNamespace(); err != nil {
		panic(err)
	}

	cases := []struct {
		description                   string
		disableInjectionForNamespace  bool
		disableInjectionForDeployment bool
		expectedResult                int
	}{
		{"no injection flag for both namespace and deployment", false, false, hasSidecar},
		{"deployment injection is disabled", false, true, noSidecar},
		{"namespace injection is disabled", true, false, noSidecar},
		{"namespace and deployment injections are disabled", true, true, noSidecar},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			testSuite.disableInjectionForNamespace(c.disableInjectionForNamespace)
			testID := generateRandomString(8)

			_, err := testSuite.createDeployment(testID, c.disableInjectionForDeployment)
			if err != nil {
				log.Panic("Cannot create deployment", err)
			}

			defer testSuite.deleteDeployment(testID)

			pods, err := testSuite.getPods(testID)

			if err != nil {
				log.Panic("There is no pod for the deployment", err)
			}

			pod := pods.Items[0]
			numberOfContainers := len(pod.Spec.Containers)

			if numberOfContainers != c.expectedResult {
				t.Errorf("pod has %d containers, but expected %d containers", numberOfContainers, c.expectedResult)
			}
		})
	}
}
