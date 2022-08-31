package istio

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/cucumber/godog"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/kubectl/pkg/util/podutils"
)

func InitializeScenarioTargetNamespaceSidecar(ctx *godog.ScenarioContext) {
	installedCase := istioInstalledCase{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		err := installedCase.getIstioPods()
		return ctx, err
	})

	ctx.Step(`^Istio component is installed$`, installedCase.istioComponentIsInstalled)
	ctx.Step(`^"([^"]*)" namespace exists`, installedCase.createTargetNamespace)
	ctx.Step(`^"([^"]*)" namespace is labeled with "([^"]*)" "([^"]*)"$`, installedCase.labelTargetNamespace)
	ctx.Step(`^Httpbin deployment is created in "([^"]*)" namespace$`, installedCase.deployHttpBinInTargetNamespace)
	ctx.Step(`^Httpbin deployment is deployed and ready in "([^"]*)" namespace$`, installedCase.waitForHttpBinInTargetNamespace)
	ctx.Step(`^there should be no pods with istio sidecar in "([^"]*)" namespace$`, installedCase.targetNamespacePodsShouldNotHaveSidecar)
	ctx.Step(`^there should be some pods with istio sidecar in "([^"]*)" namespace$`, installedCase.targetNamespacePodsShouldHaveSidecar)
	ctx.Step(`^there should be istio sidecar in httpbin pod in "([^"]*)" namespace$`, installedCase.httpBinPodShouldHaveSidecar)
	ctx.Step(`^Httpbin deployment is deleted from "([^"]*)" namespace$`, installedCase.deleteHttpBinInTargetNamespace)
	ctx.Step(`^"([^"]*)" namespace is deleted$`, installedCase.deleteTargetNamespace)
}

func (i *istioInstalledCase) httpBinPodShouldHaveSidecar(targetNamespace string) error {
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

func (i *istioInstalledCase) targetNamespacePodsShouldHaveSidecar(targetNamespace string) error {
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

func (i *istioInstalledCase) targetNamespacePodsShouldNotHaveSidecar(targetNamespace string) error {
	pods, err := k8sClient.CoreV1().Pods(targetNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, pod := range pods.Items {
		if hasIstioProxy(pod.Spec.Containers) {
			return fmt.Errorf("istio sidecars should not be deployed in %s", targetNamespace)
		}
	}
	return nil
}

func (i *istioInstalledCase) deployHttpBinInTargetNamespace(targetNamespaceName string) error {
	resources, err := readManifestToUnstructured()
	if err != nil {
		return err
	}

	for _, r := range resources {
		_, err := dynamicClient.Resource(getGroupVersionResource(r)).Namespace(targetNamespaceName).Create(context.Background(), &r, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("could not deploy httpbin deployment in %s: %s", targetNamespaceName, err)
		}
	}
	return nil
}

func (i *istioInstalledCase) waitForHttpBinInTargetNamespace(targetNamespace string) error {
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

func (i *istioInstalledCase) deleteTargetNamespace(targetNamespace string) error {
	err := k8sClient.CoreV1().Namespaces().Delete(context.Background(), targetNamespace, metav1.DeleteOptions{})

	if err != nil {
		return fmt.Errorf("could not delete namespace %s", targetNamespace)
	}

	return nil
}

func (i *istioInstalledCase) deleteHttpBinInTargetNamespace(targetNamespace string) error {
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

func (i *istioInstalledCase) createTargetNamespace(targetNamespaceName string) error {
	namespace := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: targetNamespaceName,
		},
	}
	_, err := k8sClient.CoreV1().Namespaces().Create(context.Background(), namespace, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("could not create namespace %s", targetNamespaceName)
	}
	return nil
}

func (i *istioInstalledCase) labelTargetNamespace(targetNamespace string, labelName string, labelValue string) error {
	namespace, err := k8sClient.CoreV1().Namespaces().Get(context.Background(), targetNamespace, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("could not get namespace %s", targetNamespace)
	}

	namespace.ObjectMeta.Labels[labelName] = labelValue

	_, err = k8sClient.CoreV1().Namespaces().Update(context.Background(), namespace, metav1.UpdateOptions{})

	if err != nil {
		return fmt.Errorf("could not label namespace %s", targetNamespace)
	}

	return nil
}
