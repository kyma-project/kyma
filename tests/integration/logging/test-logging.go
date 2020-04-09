package main

import (
	"log"
	"os"

	"github.com/kyma-project/kyma/tests/integration/logging/pkg/jwt"
	"github.com/kyma-project/kyma/tests/integration/logging/pkg/logstream"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const namespace = "kyma-system"

func main() {
	kubeConfig, err := loadKubeConfigOrDie()
	if err != nil {
		log.Fatal(err)
	}
	k8sClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		log.Fatalf("cannot create k8s clientset: %v", err)
	}
	log.Println("Cleaning up before starting logging test")
	err = logstream.Cleanup(namespace, k8sClient)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Deploying test-counter-pod")
	err = logstream.DeployDummyPod(namespace, k8sClient)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Waiting for test-counter-pod to run...")
	err = logstream.WaitForDummyPodToRun(namespace, k8sClient)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Test if logs from test-counter-pod are streamed by Loki")
	err = testLogStream(namespace)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Deleting test-counter-pod")
	err = logstream.Cleanup(namespace, k8sClient)
	if err != nil {
		log.Fatal(err)
	}
}

func loadKubeConfigOrDie() (*rest.Config, error) {
	if _, err := os.Stat(clientcmd.RecommendedHomeFile); os.IsNotExist(err) {
		cfg, err := rest.InClusterConfig()
		if err != nil {
			return nil, errors.Wrap(err, "cannot create in-cluster config")
		}
		return cfg, nil
	}

	cfg, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read kubeconfig")
	}
	return cfg, nil
}

func testLogStream(namespace string) error {
	token, err := jwt.GetToken()
	if err != nil {
		return err
	}
	authHeader := jwt.SetAuthHeader(token)
	err = logstream.Test("container", "count", authHeader, 0)
	if err != nil {
		return err
	}
	err = logstream.Test("app", "test-counter-pod", authHeader, 0)
	if err != nil {
		return err
	}
	err = logstream.Test("namespace", namespace, authHeader, 0)
	if err != nil {
		return err
	}
	return nil
}
