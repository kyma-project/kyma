package brokerinstaller

import (
	"fmt"
	"os"

	"k8s.io/client-go/kubernetes/typed/core/v1"

	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/helm"
)

const (
	installTimeout    int64 = 240
	tillerConnTimeout int64 = 10
)

type BrokerInstaller struct {
	k8sCli      *v1.CoreV1Client
	helmCli     *helm.Client
	chartPath   string
	releaseName string
	namespace   string
}

func New(chartPath, releaseName, namespace string) (*BrokerInstaller, error) {
	tillerHost := os.Getenv("TILLER_HOST")
	if tillerHost == "" {
		tillerHost = fmt.Sprintf("127.0.0.1:44134")
	}

	cli := helm.NewClient(helm.Host(tillerHost), helm.ConnectTimeout(tillerConnTimeout))
	return &BrokerInstaller{
		helmCli:     cli,
		releaseName: releaseName,
		chartPath:   chartPath,
		namespace:   namespace,
	}, nil
}

func (t *BrokerInstaller) Install() error {
	brokerChart, err := chartutil.Load(t.chartPath)
	if err != nil {
		return err
	}

	_, err = t.helmCli.InstallReleaseFromChart(
		brokerChart,
		t.namespace,
		helm.InstallWait(true),
		helm.InstallTimeout(installTimeout),
		helm.ValueOverrides(nil),
		helm.ReleaseName(t.releaseName),
	)
	if err != nil {
		return err
	}

	return nil
}

func (t *BrokerInstaller) Uninstall() error {
	_, err := t.helmCli.DeleteRelease(t.releaseName, helm.DeletePurge(true), helm.DeleteTimeout(installTimeout))
	return err
}

func (t *BrokerInstaller) ReleaseName() string {
	return t.releaseName
}
