package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"time"

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
	log.Println("Deploying logging test pod")
	loggingStartTime := time.Now()
	err = logstream.DeployDummyPod(loggingPodSpec, k8sClient)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Waiting for logging test pod to run...")
	err = logstream.WaitForDummyPodToRun(loggingPodSpec, k8sClient)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Test if logs from test-counter-pod are streamed by Loki")
	err = testLogStream(loggingPodSpec, loggingStartTime)
	if err != nil {
		logstream.Cleanup(loggingPodSpec, k8sClient)
		log.Fatal(err)
	}
	log.Println("Deleting test-counter-pod")
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

func testLogStream(spec logstream.PodSpec, loggingStartTime time.Time) error {
	httpClient := getHttpClient()
	token, domain, err := jwt.GetToken()
	if err != nil {
		return err
	}
	authHeader := jwt.SetAuthHeader(token)
	err = logstream.Test(domain, authHeader, "container", spec.ContainerName, loggingStartTime, httpClient)
	if err != nil {
		return err
	}
	err = logstream.Test(domain, authHeader, "app", spec.PodName, loggingStartTime, httpClient)
	if err != nil {
		return err
	}
	err = logstream.Test(domain, authHeader, "namespace", spec.Namespace, loggingStartTime, httpClient)
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
