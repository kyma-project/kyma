package istio

import (
	"context"
	"fmt"
	"os"

	"github.com/cucumber/godog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubectl/pkg/util/podutils"
)

func InitializeScenarioIstioInstalled(ctx *godog.ScenarioContext) {
	installedCase := istioInstallledCase{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		err := installedCase.getIstioPods()
		return ctx, err
	})

	ctx.Step(`^a running Kyma cluster with "([^"]*)" profile$`, installedCase.aRunningKymaClusterWithProfile)
	ctx.Step(`^Istio component is installed$`, installedCase.istioComponentIsInstalled)
	ctx.Step(`^there is (\d+) pod for Ingress gateway$`, installedCase.thereIsPodForIngressGateway)
	ctx.Step(`^there is (\d+) pod for Pilot$`, installedCase.thereIsPodForPilot)
	ctx.Step(`^Istio pods are available$`, installedCase.istioPodsAreAvailable)
	ctx.Step(`^HPA is not deployed$`, installedCase.hPAIsNotDeployed)
	ctx.Step(`^HPA is deployed$`, installedCase.hPAIsDeployed)
	InitializeScenarioTargetNamespaceSidecar(ctx)
}

func (i *istioInstallledCase) getIstioPods() error {
	istiodPods, err := listPodsIstioNamespace(metav1.ListOptions{
		LabelSelector: "istio=pilot",
	})
	if err != nil {
		return err
	}
	i.pilotPods = istiodPods

	ingressGwPods, err := listPodsIstioNamespace(metav1.ListOptions{
		LabelSelector: "istio=ingressgateway",
	})
	if err != nil {
		return err
	}
	i.ingressGwPods = ingressGwPods
	return nil
}

func (i *istioInstallledCase) aRunningKymaClusterWithProfile(profile string) error {
	p, ok := os.LookupEnv(deployedKymaProfileVar)
	if !ok {
		return fmt.Errorf("KYMA_PROFILE env variable is not set")
	}
	if profile != p {
		return fmt.Errorf("wanted %s profile but installed %s", profile, p)
	}
	return nil
}

func (i *istioInstallledCase) hPAIsNotDeployed() error {
	list, err := k8sClient.AutoscalingV1().HorizontalPodAutoscalers(istioNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	if len(list.Items) != 0 {
		return fmt.Errorf("hpa should not be deployed in %s", istioNamespace)
	}
	return nil
}

func (i *istioInstallledCase) hPAIsDeployed() error {
	list, err := k8sClient.AutoscalingV1().HorizontalPodAutoscalers(istioNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	if len(list.Items) == 0 {
		return fmt.Errorf("hpa should be deployed in %s", istioNamespace)
	}
	return nil
}

func (i *istioInstallledCase) istioComponentIsInstalled() error {
	if i.pilotPods == nil && i.ingressGwPods == nil {
		return fmt.Errorf("istio is not installed")
	}
	return nil
}

func (i *istioInstallledCase) istioPodsAreAvailable() error {
	var minReadySeconds = 1
	pods := append(i.pilotPods.Items, i.ingressGwPods.Items...)
	for _, pod := range pods {
		isPodAvailable := podutils.IsPodAvailable(&pod, int32(minReadySeconds), metav1.Now())
		if !isPodAvailable {
			return fmt.Errorf("pod %s is not available", pod.Name)
		}
	}
	return nil
}

func (i *istioInstallledCase) thereIsPodForIngressGateway(numberOfPodsRequired int) error {
	if len(i.ingressGwPods.Items) != numberOfPodsRequired {
		return fmt.Errorf("number of deployed IngressGW pods %d does not equal %d\n Pod list: %v", len(i.ingressGwPods.Items), numberOfPodsRequired, getPodListReport(i.ingressGwPods))
	}
	return nil
}

func (i *istioInstallledCase) thereIsPodForPilot(numberOfPodsRequired int) error {
	if len(i.pilotPods.Items) != numberOfPodsRequired {
		return fmt.Errorf("number of deployed Pilot pods %d does not equal %d\n Pod list: %v", len(i.pilotPods.Items), numberOfPodsRequired, getPodListReport(i.pilotPods))
	}
	return nil
}
