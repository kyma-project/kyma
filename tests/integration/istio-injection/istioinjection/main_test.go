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

var testSuite *TestSuite

func TestIstioInjection(t *testing.T) {
	namespace := fmt.Sprintf("%s-%s", namespaceNameRoot, generateRandomString(8))
	if namespace == "" {
		log.Fatal("Namespace not set.")
	}

	kubeConfig := loadKubeConfigOrDie()
	k8sClient := kubernetes.NewForConfigOrDie(kubeConfig)
	testSuite = &TestSuite{k8sClient, namespace}

	defer deleteNamespace()
	if err := createNamespace(); err != nil {
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
			disableInjectionForNamespace(c.disableInjectionForNamespace)
			testID := generateRandomString(8)

			_, err := createDeployment(testID, c.disableInjectionForDeployment)
			if err != nil {
				log.Errorf("Cannot create deployment '%s': %v", testID, err)
			}

			defer deleteDeployment(testID)

			pods, err := getPods(testID)

			if err != nil {
				log.Fatal("There is no pods for the deployment", err)
			}

			pod := pods.Items[0]
			numberOfContainers := len(pod.Spec.Containers)

			if numberOfContainers != c.expectedResult {
				t.Errorf("pod has %d containers, but expected %d containers", numberOfContainers, c.expectedResult)
			}
		})
	}
}
