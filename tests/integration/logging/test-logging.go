package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"

	"github.com/kyma-project/kyma/tests/integration/logging/pkg/jwt"
	"github.com/kyma-project/kyma/tests/integration/logging/pkg/logstream"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	loggingPodSpec = logstream.PodSpec{
		PodName:       "logging-test-app",
		ContainerName: "logging-test-counter",
		Namespace:     "kyma-system",
		LogPrefix:     "logTest",
	}
)

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
	err = logstream.Cleanup(loggingPodSpec, k8sClient)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Deploying the logging test pod")
	err = logstream.DeployDummyPod(loggingPodSpec, k8sClient)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Waiting for the logging test pod to run...")
	err = logstream.WaitForDummyPodToRun(loggingPodSpec, k8sClient)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Test if logs from the logging test pod are streamed by Loki")
	err = testLogStream(loggingPodSpec)
	if err != nil {
		logstream.Cleanup(loggingPodSpec, k8sClient)
		log.Fatal(err)
	}
	log.Println("Deleting the logging test pod")
	err = logstream.Cleanup(loggingPodSpec, k8sClient)
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

func testLogStream(spec logstream.PodSpec) error {
	httpClient := getHttpClient()
	token, domain, err := jwt.GetToken()
	if err != nil {
		return err
	}
	authHeader := jwt.SetAuthHeader(token)

	labelsToSelect := map[string]string{
		"container": spec.ContainerName,
		"app":       spec.PodName,
		"namespace": spec.Namespace,
	}

	err = logstream.Test(domain, authHeader, spec.LogPrefix, labelsToSelect, httpClient)
	if err != nil {
		return err
	}

	return nil
}

func getHttpClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}
	return client
}
