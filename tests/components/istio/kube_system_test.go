package istio

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cucumber/godog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/kubectl/pkg/util/podutils"
)

const (
	kubeSystemNamespace = "kube-system"
	proxyName           = "istio-proxy"
)

func TestKubeSystemSidecar(t *testing.T) {
	suite := godog.TestSuite{
		Name:                kubeSystemNamespace,
		ScenarioInitializer: InitializeScenarioKubeSystem,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features/kube-system.feature"},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func InitializeScenarioKubeSystem(ctx *godog.ScenarioContext) {
	installedCase := istioInstallledCase{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		err := installedCase.getIstioPods()
		return ctx, err
	})
	ctx.Step(`^Istio component is installed$`, installedCase.istioComponentIsInstalled)
	ctx.Step(`^Httpbin deployment is created in kube-system$`, installedCase.deployHttpBinInKubeSystem)
	ctx.Step(`^there should be no pods with istio sidecar in kube-system namespace$`, installedCase.kubeSystemPodsShouldNotHaveSidecar)
	ctx.Step(`^Httpbin deployment is deleted from kube-system$`, installedCase.deleteHttpBinInKubeSystem)
}

func (i *istioInstallledCase) kubeSystemPodsShouldNotHaveSidecar() error {
	pods, err := k8sClient.CoreV1().Pods(kubeSystemNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, pod := range pods.Items {
		if strings.Contains(pod.Name, "httpbin") {
			err := wait.Poll(1*time.Second, 1*time.Minute, func() (done bool, err error) {
				ready := podutils.IsPodReady(&pod)
				return ready, nil
			})
			if err != nil {
				return err
			}
		}
		for _, container := range pod.Spec.Containers {
			if container.Name == proxyName {
				return fmt.Errorf("istio sidecars should not be deployed in %s", kubeSystemNamespace)
			}
		}
	}
	return nil
}

func (i *istioInstallledCase) deployHttpBinInKubeSystem() error {
	resources, err := readManifestToUnstructured("test/httpbin.yaml")
	if err != nil {
		return err
	}

	for _, r := range resources {
		_, err = dynClient.Resource(gvr(r)).Namespace(kubeSystemNamespace).Create(context.Background(), r, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *istioInstallledCase) deleteHttpBinInKubeSystem() error {
	resources, err := readManifestToUnstructured("test/httpbin.yaml")
	if err != nil {
		return err
	}

	for _, r := range resources {
		_ = dynClient.Resource(gvr(r)).Namespace(kubeSystemNamespace).Delete(context.Background(), r.GetName(), metav1.DeleteOptions{})
	}
	return nil
}

func readFile(filepath string) (string, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func readManifestToUnstructured(filepath string) ([]*unstructured.Unstructured, error) {
	manifest, err := readFile(filepath)
	if err != nil {
		return nil, err
	}

	var result []*unstructured.Unstructured
	for _, resourceYAML := range strings.Split(string(manifest), "---") {
		if strings.TrimSpace(resourceYAML) == "" {
			continue
		}
		obj := &unstructured.Unstructured{}
		dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		_, _, err := dec.Decode([]byte(resourceYAML), nil, obj)
		if err != nil {
			return nil, err
		}
		result = append(result, obj)
	}
	return result, nil
}

func gvr(resource *unstructured.Unstructured) schema.GroupVersionResource {
	mapping, err := mapper.RESTMapping(resource.GroupVersionKind().GroupKind(), resource.GroupVersionKind().Version)
	if err != nil {
		log.Fatal(err)
	}
	return mapping.Resource
}
