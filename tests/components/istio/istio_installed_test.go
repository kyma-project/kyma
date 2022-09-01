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
	installedCase := istioInstalledCase{}
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
	ctx.Step(`^"([^"]*)" has "([^"]*)" set to cpu - "([^"]*)" and memory - "([^"]*)"$`, installedCase.componentHasResourcesSetToCpuAndMemory)
	InitializeScenarioTargetNamespaceSidecar(ctx)
}

func (i *istioInstalledCase) getIstioPods() error {
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

func (i *istioInstalledCase) aRunningKymaClusterWithProfile(profile string) error {
	p, ok := os.LookupEnv(deployedKymaProfileVar)
	if !ok {
		return fmt.Errorf("KYMA_PROFILE env variable is not set")
	}
	if profile != p {
		return fmt.Errorf("wanted %s profile but installed %s", profile, p)
	}
	return nil
}

func (i *istioInstalledCase) hPAIsNotDeployed() error {
	list, err := k8sClient.AutoscalingV1().HorizontalPodAutoscalers(istioNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	if len(list.Items) != 0 {
		return fmt.Errorf("hpa should not be deployed in %s", istioNamespace)
	}
	return nil
}

func (i *istioInstalledCase) hPAIsDeployed() error {
	list, err := k8sClient.AutoscalingV1().HorizontalPodAutoscalers(istioNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	if len(list.Items) == 0 {
		return fmt.Errorf("hpa should be deployed in %s", istioNamespace)
	}
	return nil
}

func (i *istioInstalledCase) istioComponentIsInstalled() error {
	if i.pilotPods == nil && i.ingressGwPods == nil {
		return fmt.Errorf("istio is not installed")
	}
	return nil
}

func (i *istioInstalledCase) istioPodsAreAvailable() error {
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

func (i *istioInstalledCase) thereIsPodForIngressGateway(numberOfPodsRequired int) error {
	if len(i.ingressGwPods.Items) != numberOfPodsRequired {
		return fmt.Errorf("number of deployed IngressGW pods %d does not equal %d\n Pod list: %v", len(i.ingressGwPods.Items), numberOfPodsRequired, getPodListReport(i.ingressGwPods))
	}
	return nil
}

func (i *istioInstalledCase) thereIsPodForPilot(numberOfPodsRequired int) error {
	if len(i.pilotPods.Items) != numberOfPodsRequired {
		return fmt.Errorf("number of deployed Pilot pods %d does not equal %d\n Pod list: %v", len(i.pilotPods.Items), numberOfPodsRequired, getPodListReport(i.pilotPods))
	}
	return nil
}

func (i *istioInstalledCase) componentHasResourcesSetToCpuAndMemory(component, resourceType, cpu, memory string) error {
	operator, err := getIstioOperatorFromCluster()
	if err != nil {
		return err
	}
	resources, err := getResourcesForComponent(operator, component, resourceType)
	if err != nil {
		return err
	}

	if resources.Cpu != cpu {
		return fmt.Errorf("cpu %s for component %s wasn't expected; expected=%s got=%s", resourceType, component, cpu, resources.Cpu)
	}

	if resources.Memory != memory {
		return fmt.Errorf("memory %s for component %s wasn't expected; expected=%s got=%s", resourceType, component, memory, resources.Memory)
	}

	return nil
}
