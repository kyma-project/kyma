package istio

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/cucumber/godog"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/kubectl/pkg/util/podutils"
)

var k8sClient kubernetes.Interface
var dynamicClient dynamic.Interface
var mapper *restmapper.DeferredDiscoveryRESTMapper

const (
	istioNamespace         = "istio-system"
	evalProfile            = "evaluation"
	prodProfile            = "production"
	deployedKymaProfileVar = "KYMA_PROFILE"
	exportResultVar        = "EXPORT_RESULT"
	junitFileName          = "junit-report.xml"
)

func TestMain(m *testing.M) {
	k8sClient, dynamicClient, mapper = initK8sClient()
	os.Exit(m.Run())
}

func initK8sClient() (kubernetes.Interface, dynamic.Interface, *restmapper.DeferredDiscoveryRESTMapper) {
	var kubeconfig string
	if kConfig, ok := os.LookupEnv("KUBECONFIG"); !ok {
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")

		}
	} else {
		kubeconfig = kConfig
	}
	_, err := os.Stat(kubeconfig)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Fatalf("kubeconfig %s does not exist", kubeconfig)
		}
		log.Fatalf(err.Error())
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf(err.Error())
	}
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf(err.Error())
	}
	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf(err.Error())
	}
	dc, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		log.Fatalf(err.Error())
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))
	return k8sClient, dynClient, mapper
}

func TestIstioInstalledEvaluation(t *testing.T) {
	suite := godog.TestSuite{
		Name:                evalProfile,
		ScenarioInitializer: InitializeScenarioEvalProfile,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features/istio_evaluation.feature"},
			TestingT: t,
		},
	}

	if suite.Name != os.Getenv(deployedKymaProfileVar) {
		t.Skip()
	}
	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func TestIstioInstalledProduction(t *testing.T) {
	suite := godog.TestSuite{
		Name:                prodProfile,
		ScenarioInitializer: InitializeScenarioProdProfile,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features/istio_production.feature"},
			TestingT: t,
		},
	}

	if suite.Name != os.Getenv(deployedKymaProfileVar) {
		t.Skip()
	}
	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}

}

type istioInstallledCase struct {
	pilotPods     *corev1.PodList
	ingressGwPods *corev1.PodList
}

func (i *istioInstallledCase) getIstioPods() error {
	istiodPods, err := listPodsIstioNamespace(metav1.ListOptions{
		LabelSelector: "istio=pilot",
	})
	if err != nil {
		return err
	}
	i.pilotPods = istiodPods

	for _, p := range istiodPods.Items {
		fmt.Printf("[5525] name:%v timestamp: %v\n, status:%v", p.Name, p.CreationTimestamp, p.Status)
	}

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
		return fmt.Errorf("number of deployed IngressGW pods %d does not equal %d", len(i.ingressGwPods.Items), numberOfPodsRequired)
	}
	return nil
}

func (i *istioInstallledCase) thereIsPodForPilot(numberOfPodsRequired int) error {
	if len(i.pilotPods.Items) != numberOfPodsRequired {
		return fmt.Errorf("number of deployed Pilot pods %d does not equal %d", len(i.pilotPods.Items), numberOfPodsRequired)
	}
	return nil
}

func InitializeScenarioEvalProfile(ctx *godog.ScenarioContext) {
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
}

func InitializeScenarioProdProfile(ctx *godog.ScenarioContext) {
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
	ctx.Step(`^HPA is deployed$`, installedCase.hPAIsDeployed)
}

func listPodsIstioNamespace(istiodPodsSelector metav1.ListOptions) (*corev1.PodList, error) {
	return k8sClient.CoreV1().Pods(istioNamespace).List(context.Background(), istiodPodsSelector)
}
