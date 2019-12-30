package istioinjection

import (
	"os"
	"os/signal"
	"testing"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const namespaceNameRoot = "istio-injection-tests"

var kubeConfig *rest.Config
var k8sClient *kubernetes.Clientset
var namespace string

func TestMain(m *testing.M) {
	namespace = namespaceNameRoot + "-" + generateRandomString(8)
	if namespace == "" {
		log.Error("Namespace not set.")
		os.Exit(2)
	}

	kubeConfig = loadKubeConfigOrDie()
	k8sClient = kubernetes.NewForConfigOrDie(kubeConfig)

	os.Exit(testWithNamespace(m))
}

func testWithNamespace(m *testing.M) int {
	catchInterrupt()

	defer deleteNamespace()
	if err := createNamespace(); err != nil {
		panic(err)
	}

	return m.Run()
}

func catchInterrupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		deleteNamespace()
		os.Exit(2)
	}()
}
