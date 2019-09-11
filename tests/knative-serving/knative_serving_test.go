package knative_serving_acceptance

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"knative.dev/serving/pkg/apis/serving/v1beta1"

	"github.com/avast/retry-go"
	corev1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	// allow client authentication against GKE clusters
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	serving "knative.dev/serving/pkg/apis/serving/v1alpha1"
	servingtyped "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"

	"github.com/kyma-project/kyma/common/ingressgateway"
)

const (
	// Domain name used to build the URL tests send HTTP requests to.
	// `test-service.knative-serving.DOMAIN` is expected to resolve to the IP of the Istio Ingress Gateway.
	domainNameEnvVar = "DOMAIN_NAME"

	// Message that should be printed by the sample Hello World service.
	// https://knative.dev/docs/serving/samples/hello-world/helloworld-go/index.html
	targetEnvVar = "TARGET"
)

func TestKnativeServingAcceptance(t *testing.T) {
	domainName := mustGetenv(t, domainNameEnvVar)
	target := mustGetenv(t, targetEnvVar)

	testServiceURL := fmt.Sprintf("https://test-service.knative-serving.%s", domainName)

	ingressClient, err := ingressgateway.FromEnv().Client()
	if err != nil {
		t.Fatalf("Unexpected error when creating ingressgateway client: %s", err)
	}

	kubeConfig := loadKubeConfigOrDie(t)
	serviceClient := servingtyped.NewForConfigOrDie(kubeConfig).Services("knative-serving")

	svc := serving.Service{
		ObjectMeta: meta.ObjectMeta{
			Name: "test-service",
		},
		Spec: serving.ServiceSpec{
			ConfigurationSpec: serving.ConfigurationSpec{
				Template: &serving.RevisionTemplateSpec{
					ObjectMeta: meta.ObjectMeta{
						Name: "test-service-foo",
					},
					Spec: serving.RevisionSpec{
						RevisionSpec: v1beta1.RevisionSpec{
							PodSpec: corev1.PodSpec{
								Containers: []corev1.Container{{
									Image: "gcr.io/knative-samples/helloworld-go",
									Env: []corev1.EnvVar{
										{
											Name: targetEnvVar,
											Value: target,
										},
									},
								},
								},
							},
						},
					},
				},
			},
		},
	}

	if _, err := serviceClient.Create(&svc); err != nil {
		t.Fatalf("Cannot create test Service: %v", err)
	}
	defer deleteService(serviceClient, svc.Name)

	err = retry.Do(func() error {
		fmt.Printf("Calling: %s\n", testServiceURL)
		t.Logf("Calling: %s", testServiceURL)
		resp, err := ingressClient.Get(testServiceURL)
		if err != nil {
			return err
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		msg := strings.TrimSpace(string(body))
		expectedMsg := fmt.Sprintf("Hello %s!", target)

		t.Logf("Received %d: %q", resp.StatusCode, msg)

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}
		if msg != expectedMsg {
			return fmt.Errorf("expected response to be %q, got %q", expectedMsg, msg)
		}

		return nil
	}, retry.OnRetry(func(n uint, err error) {
		t.Logf("[%v] try failed: %s", n, err)
	}), retry.Attempts(20),
	)

	if err != nil {
		t.Fatalf("Cannot get test service response: %s", err)
	}
}

func mustGetenv(t *testing.T, name string) string {
	t.Helper()

	env := os.Getenv(name)
	if env == "" {
		t.Fatalf("Missing %q variable", name)
	}
	return env
}

func loadKubeConfigOrDie(t *testing.T) *rest.Config {
	t.Helper()

	if _, err := os.Stat(clientcmd.RecommendedHomeFile); os.IsNotExist(err) {
		cfg, err := rest.InClusterConfig()
		if err != nil {
			t.Fatalf("Cannot create in-cluster config: %v", err)
		}
		return cfg
	}

	var err error
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		t.Fatalf("Cannot read kubeconfig: %s", err)
	}
	return kubeConfig
}

func deleteService(servingClient servingtyped.ServiceInterface, name string) {
	var deleteImmediately int64
	_ = servingClient.Delete(name, &meta.DeleteOptions{
		GracePeriodSeconds: &deleteImmediately,
	})
}
