package istioinjection

import (
	"fmt"
	"os"
	"os/signal"
	"testing"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

const namespaceNameRoot = "istio-injection-tests"

type TestSuite struct {
	k8sClient *kubernetes.Clientset
	namespace string
}

var testSuite *TestSuite

func TestMain(m *testing.M) {
	namespace := fmt.Sprintf("%s-%s", namespaceNameRoot, generateRandomString(8))
	if namespace == "" {
		log.Error("Namespace not set.")
		os.Exit(1)
	}

	kubeConfig := loadKubeConfigOrDie()
	k8sClient := kubernetes.NewForConfigOrDie(kubeConfig)

	testSuite = &TestSuite{k8sClient, namespace}

	os.Exit(testSuite.testWithNamespace(m))
}

func (r *TestSuite) testWithNamespace(m *testing.M) int {
	r.catchInterrupt()

	defer deleteNamespace()
	if err := createNamespace(); err != nil {
		panic(err)
	}

	return m.Run()
}

func (r *TestSuite) catchInterrupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		deleteNamespace()
		os.Exit(1)
	}()
}
