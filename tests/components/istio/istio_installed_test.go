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
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/kubectl/pkg/util/podutils"
)

var k8sClient kubernetes.Interface

const (
	istioNamespace         = "istio-system"
	evalProfile            = "evaluation"
	deployedKymaProfileVar = "KYMA_PROFILE"
)

func TestMain(m *testing.M) {
	k8sClient = initK8sClient()
	os.Exit(m.Run())
}

func initK8sClient() kubernetes.Interface {
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
	return k8sClient
}

func TestIstioInstalled(t *testing.T) {

	suite := godog.TestSuite{
		Name:                evalProfile,
		ScenarioInitializer: InitializeScenarioEvalProfile,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
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

	ingressGwPods, err := listPodsIstioNamespace(metav1.ListOptions{
		LabelSelector: "istio=ingressgateway",
	})
	if err != nil {
		return err
	}
	i.ingressGwPods = ingressGwPods
	return nil
}

func (i *istioInstallledCase) installedIstioProfile(profile string) error {
	p, ok := os.LookupEnv(deployedKymaProfileVar)
	if !ok {
		return fmt.Errorf("KYMA_PROFILE env variable is not set")
	}
	if profile != p {
		return fmt.Errorf("wanted %s profile but installed %s", profile, p)
	}
	return nil
}

func (i *istioInstallledCase) hPAShouldNotBeDeployed() error {
	list, err := k8sClient.AutoscalingV1().HorizontalPodAutoscalers(istioNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	if len(list.Items) != 0 {
		return fmt.Errorf("hpa should not be deployed in %s", istioNamespace)
	}
	return nil
}

func (i *istioInstallledCase) thereShouldBePodForPilotAndPodForIngressGateway(numberOfPilotPods, numberOfIGWPods int) error {
	if len(i.pilotPods.Items) != numberOfPilotPods {
		return fmt.Errorf("number of deployed Pilot pods %d does not equal %d", len(i.pilotPods.Items), numberOfPilotPods)
	}

	if len(i.ingressGwPods.Items) != numberOfIGWPods {
		return fmt.Errorf("number of deployed ingressgw pods %d does not equal %d", len(i.ingressGwPods.Items), numberOfIGWPods)
	}
	return nil
}

func (i *istioInstallledCase) podsShouldBeAvailableForAtLeastSeconds(minReadySeconds int) error {
	pods := append(i.pilotPods.Items, i.ingressGwPods.Items...)
	for _, pod := range pods {
		isPodAvailable := podutils.IsPodAvailable(&pod, int32(minReadySeconds), metav1.Now())
		if !isPodAvailable {
			return fmt.Errorf("pod %s is not available", pod.Name)
		}
	}
	return nil
}

func listPodsIstioNamespace(istiodPodsSelector metav1.ListOptions) (*corev1.PodList, error) {
	return k8sClient.CoreV1().Pods(istioNamespace).List(context.Background(), istiodPodsSelector)
}

func InitializeScenarioEvalProfile(ctx *godog.ScenarioContext) {
	installedCase := istioInstallledCase{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		err := installedCase.getIstioPods()
		return ctx, err
	})
	ctx.Step(`^installed Istio "([^"]*)" profile$`, installedCase.installedIstioProfile)
	ctx.Step(`^HPA should not be deployed$`, installedCase.hPAShouldNotBeDeployed)
	ctx.Step(`^Istio pods should be available for at least (\d+) seconds$`, installedCase.podsShouldBeAvailableForAtLeastSeconds)
	ctx.Step(`^there should be (\d+) pod for pilot and (\d+) pod for ingress gateway$`, installedCase.thereShouldBePodForPilotAndPodForIngressGateway)
}
