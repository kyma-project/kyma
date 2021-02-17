package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/kyma-project/kyma/tests/integration/logging/pkg/jwt"
	"github.com/kyma-project/kyma/tests/integration/logging/pkg/logstream"
	"github.com/kyma-project/kyma/tests/integration/logging/pkg/request"
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
	httpClient *http.Client
)

func main() {
	httpClient = request.GetHttpClient()
	token, domain, err := jwt.GetToken()
	if err != nil {
		log.Fatal(err)
	}
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
	err = testLogStream(loggingPodSpec, token, domain)
	if err != nil {
		logstream.Cleanup(loggingPodSpec, k8sClient)
		log.Fatal(err)
	}
	log.Println("Deleting the logging test pod")
	err = logstream.Cleanup(loggingPodSpec, k8sClient)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Checking that a JWT token is required for accessing Loki")
	err = checkTokenIsRequired(domain)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Checking that sending a request to Loki with a wrong path is forbidden")
	err = checkWrongPathIsForbidden(domain, token)
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

func testLogStream(spec logstream.PodSpec, token string, domain string) error {
	authHeader := jwt.SetAuthHeader(token)

	labelsToSelect := map[string]string{
		"container": spec.ContainerName,
		"app":       spec.PodName,
		"namespace": spec.Namespace,
	}

	err := logstream.Test(domain, authHeader, spec.LogPrefix, labelsToSelect, httpClient)
	if err != nil {
		return err
	}

	return nil
}

func checkTokenIsRequired(domain string) error {
	lokiURL := fmt.Sprintf(`https://loki.%s/api/prom`, domain)
	// sending a request to Loki wihtout a JWT token in the header
	respStatus, _, err := request.DoGet(httpClient, lokiURL, "")
	if err != nil {
		return errors.Wrap(err, "cannot send request to Loki")
	}
	if respStatus != http.StatusForbidden {
		return errors.Errorf("received status code %d instead of %d when accessing Loki wihout a JWT token", respStatus, http.StatusForbidden)
	}
	return nil
}

func checkWrongPathIsForbidden(domain string, token string) error {
	lokiURL := fmt.Sprintf(`https://loki.%s/api/wrongPath`, domain)
	authHeader := jwt.SetAuthHeader(token)
	// sending a request with a wrong path
	respStatus, _, err := request.DoGet(httpClient, lokiURL, authHeader)
	if err != nil {
		return errors.Wrap(err, "cannot send request to Loki")
	}
	if respStatus != http.StatusForbidden {
		return errors.Errorf("received status code %d instead of %d when sending a request to Loki with a wrong path", respStatus, http.StatusForbidden)
	}
	return nil
}
