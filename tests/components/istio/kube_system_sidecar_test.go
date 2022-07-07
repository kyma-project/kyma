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
		if hasIstioProxy(pod.Spec.Containers) {
			return fmt.Errorf("istio sidecars should not be deployed in %s", kubeSystemNamespace)
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
