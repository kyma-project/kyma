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
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/kubectl/pkg/util/podutils"
)

var k8sClient kubernetes.Interface

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

func TestIstio(t *testing.T) {

	t.Run("list istio pilot and ingresgw pods", func(t *testing.T) {
		listOptions := metav1.ListOptions{
			LabelSelector: "istio in (pilot, ingressgateway)",
		}
		podList, err := k8sClient.CoreV1().Pods("istio-system").List(context.Background(), listOptions)
		require.NoError(t, err)
		if os.Getenv("KYMA_PROFILE") == "production" {
			require.Len(t, podList.Items, 5)
		} else {
			require.Len(t, podList.Items, 2)
		}

		minReadySeconds := int32(3)
		t.Run(fmt.Sprintf("and make sure they are available at least for %d seconds", minReadySeconds), func(t *testing.T) {
			for _, pod := range podList.Items {
				require.True(t, podutils.IsPodAvailable(&pod, minReadySeconds, metav1.Now()))
			}
		})
	})
}

// TODO: test should be using Gherkin files as inputs in followup task
// see some suggestions: https://github.com/kyma-project/kyma/pull/14133#issuecomment-1114565277
func installedIstio() error {
	return godog.ErrPending
}

func thereShouldBeAtLeastPods(arg1 int) error {
	return godog.ErrPending
}

func theyShouldBeAvailableForAtLeastSeconds(arg1 int) error {
	return godog.ErrPending
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^installed istio$`, installedIstio)
	ctx.Step(`^there should be at least (\d+) pods$`, thereShouldBeAtLeastPods)
	ctx.Step(`^they should be available for at least (\d+) seconds$`, theyShouldBeAvailableForAtLeastSeconds)
}
