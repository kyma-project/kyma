package istio

import (
	"context"
	"fmt"
	"time"

	"github.com/cucumber/godog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/kubectl/pkg/util/podutils"
)

func InitializeScenarioKymaSystemSidecar(ctx *godog.ScenarioContext) {
	installedCase := istioInstallledCase{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		err := installedCase.getIstioPods()
		return ctx, err
	})
	ctx.Step(`^Istio component is installed$`, installedCase.istioComponentIsInstalled)
	ctx.Step(`^there should be pods with istio sidecar in "([^"]*)" namespace$`, installedCase.kymaSystemPodsShouldHaveSidecar)
	ctx.Step(`^Httpbin deployment is created in "([^"]*)"$`, installedCase.deployHttpBinInKymaSystem)
	ctx.Step(`^Httpbin deployment should be deployed and ready in "([^"]*)" namespace$`, installedCase.waitForHttpBinInKymaSystem)
	ctx.Step(`^there should be istio sidecar in httpbin pod in "([^"]*)" namespace$`, installedCase.httpBinPodShouldHaveSidecar)
	ctx.Step(`^Httpbin deployment is deleted from "([^"]*)"$`, installedCase.deleteHttpBinInKymaSystem)
}

func (i *istioInstallledCase) kymaSystemPodsShouldHaveSidecar(targetNamespace string) error {
	pods, err := k8sClient.CoreV1().Pods(targetNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	var proxies []string
	for _, pod := range pods.Items {
		if !metav1.HasAnnotation(pod.ObjectMeta, "sidecar.istio.io/inject") {
			for _, container := range pod.Spec.Containers {
				if container.Name == proxyName {
					proxies = append(proxies, pod.Name)
				}
			}
			if len(proxies) == 0 {
				return fmt.Errorf("istio sidecars should be deployed in %s", targetNamespace)
			}
		}
	}
	return nil
}

func (i *istioInstallledCase) httpBinPodShouldHaveSidecar(targetNamespace string) error {
	pods, err := k8sClient.CoreV1().Pods(targetNamespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: "app=httpbin",
	})
	if err != nil {
		return err
	}
	for _, pod := range pods.Items {
		if !hasIstioProxy(pod.Spec.Containers) {
			return fmt.Errorf("istio sidecars should be deployed in %s", targetNamespace)
		}
	}

	return nil
}

func (i *istioInstallledCase) deployHttpBinInKymaSystem(targetNamespace string) error {
	resources, err := readManifestToUnstructured()
	if err != nil {
		return err
	}

	for _, r := range resources {
		_, err := dynamicClient.Resource(getGroupVersionResource(r)).Namespace(targetNamespace).Create(context.Background(), &r, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("could not deploy httpbin deployment in %s: %s", targetNamespace, err)
		}
	}
	return nil
}

func (i *istioInstallledCase) waitForHttpBinInKymaSystem(targetNamespace string) error {
	err := wait.Poll(1*time.Second, 1*time.Minute, func() (done bool, err error) {
		pods, err := k8sClient.CoreV1().Pods(targetNamespace).List(context.Background(), metav1.ListOptions{
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

func (i *istioInstallledCase) deleteHttpBinInKymaSystem(targetNamespace string) error {
	resources, err := readManifestToUnstructured()
	if err != nil {
		return err
	}

	for _, r := range resources {
		err = dynamicClient.Resource(getGroupVersionResource(r)).Namespace(targetNamespace).Delete(context.Background(), r.GetName(), metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("could not delete httpbin deployment from %s", targetNamespace)
		}
	}
	return nil
}
