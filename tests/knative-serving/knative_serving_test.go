package knative_serving_acceptance

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"knative.dev/serving/pkg/apis/serving/v1beta1"

	"github.com/avast/retry-go"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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
	// Message that should be printed by the sample Hello World service.
	// https://knative.dev/docs/serving/samples/hello-world/helloworld-go/index.html
	targetEnvVar = "TARGET"
)

func TestKnativeServingAcceptance(t *testing.T) {
	target := mustGetenv(t, targetEnvVar)

	ingressClient, err := ingressgateway.FromEnv().Client()

	if err != nil {
		t.Fatalf("Unexpected error when creating ingressgateway client: %s", err)
	}

	kubeConfig := loadKubeConfigOrDie(t)
	serviceClient := servingtyped.NewForConfigOrDie(kubeConfig).Services("knative-serving")

	const podName = "test-service-foo"
	const svcName = "test-service"
	svc := serving.Service{
		ObjectMeta: meta.ObjectMeta{
			Name: svcName,
		},
		Spec: serving.ServiceSpec{
			ConfigurationSpec: serving.ConfigurationSpec{
				Template: &serving.RevisionTemplateSpec{
					ObjectMeta: meta.ObjectMeta{
						Name: podName,
					},
					Spec: serving.RevisionSpec{
						RevisionSpec: v1beta1.RevisionSpec{
							PodSpec: corev1.PodSpec{
								Containers: []corev1.Container{{
									Image: "gcr.io/knative-samples/helloworld-go",
									Env: []corev1.EnvVar{
										{
											Name:  targetEnvVar,
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

	_, err = serviceClient.Create(&svc)

	switch {
	case errors.IsAlreadyExists(err):
		// reuse the existing Knative service
	case err != nil:
		t.Fatalf("Cannot create test Service: %v", err)
	}

	defer deleteService(t, serviceClient, svc.Name)
	var testServiceURL string

	err = retry.Do(func() error {
		ksvc, err := serviceClient.Get(svcName, meta.GetOptions{})
		if err != nil {
			return err
		}
		if ksvc.Status.URL.String() == "" {
			return fmt.Errorf("url not set yet")
		}
		svcurl := ksvc.Status.URL

		//TODO(k15r): ksvc returns a URL with scheme http, but this fails as the client then tries to
		//  open an unencrypted connection on a secure port (443, as probably configured by the ingressgateway.client() )
		svcurl.Scheme = "https"
		testServiceURL = svcurl.String()
		return nil

	}, retry.DelayType(retry.FixedDelay), retry.Delay(5*time.Second), retry.Attempts(10))

	if err != nil {
		t.Fatalf("Cannot get test service url: %s", err)
	}

	err = retry.Do(func() error {
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
		retry.DelayType(retry.FixedDelay),
		retry.Delay(5*time.Second),
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

func deleteService(t *testing.T, servingClient servingtyped.ServiceInterface, name string) {
	t.Helper()

	err := servingClient.Delete(name, &meta.DeleteOptions{})
	if err != nil {
		t.Fatalf("Cannot delete service %v, Error: %v", name, err)
	}
}
