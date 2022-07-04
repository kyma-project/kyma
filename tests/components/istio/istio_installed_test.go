package istio

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/tidwall/pretty"
	"gitlab.com/rodrigoodhin/gocure/models"
	"gitlab.com/rodrigoodhin/gocure/pkg/gocure"
	"gitlab.com/rodrigoodhin/gocure/report/html"
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

var t *testing.T
var goDogOpts = godog.Options{
	Output:   colors.Colored(os.Stdout),
	Format:   "pretty",
	TestingT: t,
}

func init() {
	godog.BindCommandLineFlags("godog.", &goDogOpts)
}

func generateHTMLReport() {
	html := gocure.HTML{
		Config: html.Data{
			InputJsonPath:    cucumberFileName,
			OutputHtmlFolder: "reports/",
			Title:            "Kyma Istio component tests",
			Metadata: models.Metadata{
				TestEnvironment: os.Getenv(deployedKymaProfileVar),
				Platform:        runtime.GOOS,
				Parallel:        "Scenarios",
				Executed:        "Remote",
				AppVersion:      "main",
				Browser:         "default",
			},
		},
	}
	err := html.Generate()
	if err != nil {
		log.Fatalf(err.Error())
	}
}

func TestMain(m *testing.M) {
	pflag.Parse()
	goDogOpts.Paths = pflag.Args()
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
	evalOpts := goDogOpts
	evalOpts.Paths = []string{"features/istio_evaluation.feature", "features/kube_system_sidecar.feature", "features/namespace_disabled_sidecar.feature", "features/kyma_system_sidecar.feature"}

	suite := godog.TestSuite{
		Name:                evalProfile,
		ScenarioInitializer: InitializeScenarioEvalProfile,
		Options:             &evalOpts,
	}

	if suite.Name != os.Getenv(deployedKymaProfileVar) {
		t.Skip()
	}
	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
	if os.Getenv(exportResultVar) == "true" {
		generateHTMLReport()
	}
}

func TestIstioInstalledProduction(t *testing.T) {
	prodOpts := goDogOpts
	prodOpts.Paths = []string{"features/istio_production.feature", "features/kube_system_sidecar.feature", "features/namespace_disabled_sidecar.feature", "features/kyma_system_sidecar.feature"}

	suite := godog.TestSuite{
		Name:                prodProfile,
		ScenarioInitializer: InitializeScenarioProdProfile,
		Options:             &prodOpts,
	}

	if suite.Name != os.Getenv(deployedKymaProfileVar) {
		t.Skip()
	}
	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
	if os.Getenv(exportResultVar) == "true" {
		generateHTMLReport()
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
	InitializeScenarioKubeSystemSidecar(ctx)
	InitializeScenarioTargetNamespaceSidecar(ctx)
	InitializeScenarioKymaSystemSidecar(ctx)
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
	InitializeScenarioKubeSystemSidecar(ctx)
	InitializeScenarioTargetNamespaceSidecar(ctx)
	InitializeScenarioKymaSystemSidecar(ctx)
}

func listPodsIstioNamespace(istiodPodsSelector metav1.ListOptions) (*corev1.PodList, error) {
	return k8sClient.CoreV1().Pods(istioNamespace).List(context.Background(), istiodPodsSelector)
}

func getPodListReport(list *corev1.PodList) string {
	type returnedPodList struct {
		PodList []struct {
			Metadata struct {
				Name              string `json:"name"`
				CreationTimestamp string `json:"creationTimestamp"`
			} `json:"metadata"`
			Status struct {
				Phase string `json:"phase"`
			} `json:"status"`
		} `json:"items"`
	}

	p := returnedPodList{}
	toMarshal, _ := json.Marshal(list)
	json.Unmarshal(toMarshal, &p)
	toPrint, _ := json.Marshal(p)
	return string(pretty.Pretty(toPrint))
}
