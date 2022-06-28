package istio

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/cucumber/godog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/util/wait"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/kubectl/pkg/util/podutils"
)

const (
	kubeSystemNamespace = "kube-system"
	proxyName           = "istio-proxy"
)

//go:embed test/httpbin.yaml
var httpbin []byte

func InitializeScenarioKubeSystemSidecar(ctx *godog.ScenarioContext) {
	installedCase := istioInstallledCase{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		err := installedCase.getIstioPods()
		return ctx, err
	})
	ctx.Step(`^Istio component is installed$`, installedCase.istioComponentIsInstalled)
	ctx.Step(`^there should be no pods with istio sidecar in kube-system namespace$`, installedCase.kubeSystemPodsShouldNotHaveSidecar)
	ctx.Step(`^Httpbin deployment is created in kube-system$`, installedCase.deployHttpBinInKubeSystem)
	ctx.Step(`^Httpbin deployment should be deployed and ready$`, installedCase.waitForHttpBinInKubeSystem)
	ctx.Step(`^there should be no pods with istio sidecar in kube-system namespace$`, installedCase.kubeSystemPodsShouldNotHaveSidecar)
	ctx.Step(`^Httpbin deployment is deleted from kube-system$`, installedCase.deleteHttpBinInKubeSystem)
}

func (i *istioInstallledCase) kubeSystemPodsShouldNotHaveSidecar() error {
	pods, err := k8sClient.CoreV1().Pods(kubeSystemNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			if container.Name == proxyName {
				return fmt.Errorf("istio sidecars should not be deployed in %s", kubeSystemNamespace)
			}
		}
	}
	return nil
}

func (i *istioInstallledCase) deployHttpBinInKubeSystem() error {
	resources, err := readManifestToUnstructured()
	if err != nil {
		return err
	}

	for _, r := range resources {
		_, err := dynamicClient.Resource(getGroupVersionResource(r)).Namespace(kubeSystemNamespace).Create(context.Background(), &r, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("could not deploy httpbin deployment in %s: %s", kubeSystemNamespace, err)
		}
	}
	return nil
}

func (i *istioInstallledCase) waitForHttpBinInKubeSystem() error {
	err := wait.Poll(1*time.Second, 1*time.Minute, func() (done bool, err error) {
		pods, err := k8sClient.CoreV1().Pods(kubeSystemNamespace).List(context.Background(), metav1.ListOptions{
			LabelSelector: "app=httpbin",
		})
		if err != nil {
			return false, err
		}
		for _, pod := range pods.Items {
			ready := podutils.IsPodReady(&pod)
			return ready, nil
		}
		return false, err
	})
	if err != nil {
		return fmt.Errorf("httpbin is not ready: %s", err)
	}
	return nil
}

func (i *istioInstallledCase) deleteHttpBinInKubeSystem() error {
	resources, err := readManifestToUnstructured()
	if err != nil {
		return err
	}

	for _, r := range resources {
		err = dynamicClient.Resource(getGroupVersionResource(r)).Namespace(kubeSystemNamespace).Delete(context.Background(), r.GetName(), metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("could not delete httpbin deployment from %s", kubeSystemNamespace)
		}
	}
	return nil
}

func readManifestToUnstructured() ([]unstructured.Unstructured, error) {
	var err error
	var unstructList []unstructured.Unstructured

	decoder := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(httpbin), 200)
	for {
		var rawObj runtime.RawExtension
		if err = decoder.Decode(&rawObj); err != nil {
			break
		}
		obj, _, err := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(rawObj.Raw, nil, nil)
		if err != nil {
			break
		}
		unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			break
		}
		unstructuredObj := unstructured.Unstructured{Object: unstructuredMap}
		unstructList = append(unstructList, unstructuredObj)
	}
	if err != io.EOF {
		return unstructList, err
	}

	return unstructList, nil
}

func getGroupVersionResource(resource unstructured.Unstructured) schema.GroupVersionResource {
	mapping, err := mapper.RESTMapping(resource.GroupVersionKind().GroupKind(), resource.GroupVersionKind().Version)
	if err != nil {
		log.Fatal(err)
	}
	return mapping.Resource
}
