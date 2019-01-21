package knative_serving_acceptance

import (
	"fmt"
	"github.com/avast/retry-go"
	serving_api "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	serving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	"github.com/kyma-project/kyma/common/ingressgateway"
	"io/ioutil"
	core_api "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestKnativeServing_Acceptance(t *testing.T) {
	domainName := MustGetenv(t, "DOMAIN_NAME")
	target := MustGetenv(t, "TARGET")

	testServiceURL := fmt.Sprintf("https://test-service.knative-serving.%s", domainName)

	ingressClient, err := ingressgateway.FromEnv().Client()
	if err != nil {
		t.Fatalf("Unexpected error when creating ingressgateway client: %s", err)
	}

	kubeConfig := loadKubeConfigOrDie()
	serviceClient := serving.NewForConfigOrDie(kubeConfig).Services("knative-serving")
	service, err := serviceClient.Create(&serving_api.Service{
		ObjectMeta: meta.ObjectMeta{
			Name: "test-service",
		},
		Spec: serving_api.ServiceSpec{
			RunLatest: &serving_api.RunLatestType{
				Configuration: serving_api.ConfigurationSpec{
					RevisionTemplate: serving_api.RevisionTemplateSpec{
						Spec: serving_api.RevisionSpec{
							Container: core_api.Container{
								Image: "gcr.io/knative-samples/helloworld-go",
								Env: []core_api.EnvVar{
									{
										Name:  "TAREGT",
										Value: target,
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
		log.Printf("Received %v: '%s'", resp.StatusCode, msg)

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code: %v", resp.StatusCode)
		}
		if msg != expectedMsg {
			return fmt.Errorf("unexpected response: '%s'", msg)
		}

		return nil
	}, retry.OnRetry(func(n uint, err error) {
		log.Printf("[%v] try failed: %s", n, err)
	}))

	if err != nil {
		t.Fatalf("cannot get test service response: %s", err)
	}
}

func MustGetenv(t *testing.T, name string) string {
	env := os.Getenv(name)
	if env == "" {
		t.Fatalf("Missing '%s' variable", name)
	}
	return env
}

func loadKubeConfigOrDie() *rest.Config {
	if _, err := os.Stat(clientcmd.RecommendedHomeFile); os.IsNotExist(err) {
		cfg, err := rest.InClusterConfig()
		if err != nil {
			log.Fatalf("Cannot create in-cluster config: %v", err)
		}
		return cfg
	}

	var err error
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		log.Fatalf("Cannot read kubeconfig: %s", err)
	}
	return kubeConfig
}

func deleteService(servingClient serving.ServiceInterface, service *serving_api.Service) {
	var deleteImmediately int64
	_ = servingClient.Delete(service.Name, &meta.DeleteOptions{
		GracePeriodSeconds: &deleteImmediately,
	})
}
