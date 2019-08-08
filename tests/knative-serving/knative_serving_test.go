package knative_serving_acceptance

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	retry "github.com/avast/retry-go"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	servingtyped "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"

	"github.com/kyma-project/kyma/common/ingressgateway"
)

func TestKnativeServingAcceptance(t *testing.T) {
	domainName := mustGetenv(t, "DOMAIN_NAME")
	target := mustGetenv(t, "TARGET")

	testServiceURL := fmt.Sprintf("https://test-service.knative-serving.%s", domainName)

	ingressClient, err := ingressgateway.FromEnv().Client()
	if err != nil {
		t.Fatalf("Unexpected error when creating ingressgateway client: %s", err)
	}

	kubeConfig := loadKubeConfigOrDie(t)
	serviceClient := servingtyped.NewForConfigOrDie(kubeConfig).Services("knative-serving")
	service, err := serviceClient.Create(&serving.Service{
		ObjectMeta: meta.ObjectMeta{
			Name: "test-service",
		},
		Spec: serving.ServiceSpec{
			RunLatest: &serving.RunLatestType{
				Configuration: serving.ConfigurationSpec{
					RevisionTemplate: serving.RevisionTemplateSpec{
						Spec: serving.RevisionSpec{
							Container: core.Container{
								Image: "gcr.io/knative-samples/helloworld-go",
								Env: []core.EnvVar{
									{
										Name:  "TARGET",
										Value: target,
									},
								},
								Resources: core.ResourceRequirements{
									Requests: core.ResourceList{
										core.ResourceCPU: *resource.NewMilliQuantity(50, resource.DecimalSI), // 50m
									},
								},
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Cannot create test service: %v", err)
	}
	defer deleteService(serviceClient, service)

	err = retry.Do(func() error {
		t.Logf("Calling: %s", testServiceURL)
		resp, err := ingressClient.Get(testServiceURL)
		if err != nil {
			return err
		}

		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		msg := strings.TrimSpace(string(bytes))
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

func deleteService(servingClient servingtyped.ServiceInterface, service *serving.Service) {
	var deleteImmediately int64
	_ = servingClient.Delete(service.Name, &meta.DeleteOptions{
		GracePeriodSeconds: &deleteImmediately,
	})
}
