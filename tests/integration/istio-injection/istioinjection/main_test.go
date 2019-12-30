package istioinjection

import (
	"os"
	"os/signal"
	"testing"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

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

func createNamespace() error {
	log.Infof("Creating namespace '%s", namespace)
	_, err := k8sClient.CoreV1().Namespaces().Create(&v1.Namespace{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: namespace,
			// Labels: map[string]string{
			// 	"istio-injection": "enabled",
			// },
		},
		Spec: v1.NamespaceSpec{},
	})
	if err != nil {
		log.Errorf("Cannot create namespace '%s': %v", namespace, err)
		return err
	}
	return nil
}

func deleteNamespace() {
	log.Infof("Deleting namespace '%s", namespace)
	var deleteImmediately int64
	err := k8sClient.CoreV1().Namespaces().Delete(namespace, &meta_v1.DeleteOptions{
		GracePeriodSeconds: &deleteImmediately,
	})
	if err != nil {
		log.Errorf("Cannot delete namespace '%s': %v", namespace, err)
	}
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

func loadKubeConfigOrDie() *rest.Config {
	if _, err := os.Stat(clientcmd.RecommendedHomeFile); os.IsNotExist(err) {
		cfg, err := rest.InClusterConfig()
		if err != nil {
			log.Errorf("Cannot create in-cluster config: %v", err)
			panic(err)
		}
		return cfg
	}

	cfg, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		log.Errorf("Cannot read kubeconfig: %s", err)
		panic(err)
	}
	return cfg
}

func testWithNamespace(m *testing.M) int {
	catchInterrupt()

	defer deleteNamespace()
	if err := createNamespace(); err != nil {
		panic(err)
	}

	return m.Run()
}
