package helm

import (
	"time"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/chart"
	rls "k8s.io/helm/pkg/proto/hapi/services"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
)

const (
	installTimeout  = time.Hour
	deleteWithPurge = true
)

// NewClient creates Tiller client
func NewClient(cfg Config, log *logrus.Entry) *Client {
	return &Client{
		tillerHost:        cfg.TillerHost,
		tillerConnTimeout: int64(cfg.TillerConnectionTimeout),
		log:               log.WithField("service", "helm_client"),
	}
}

// Client provide communication with Tiller
type Client struct {
	tillerHost        string
	tillerConnTimeout int64
	log               *logrus.Entry
}

// Install is installing chart release
func (cli *Client) Install(c *chart.Chart, values internal.ChartValues, releaseName internal.ReleaseName, namespace internal.Namespace) (*rls.InstallReleaseResponse, error) {
	cli.log.Infof("Installing chart with release name [%s] in namespace [%s]", releaseName, namespace)
	byteValues, err := yaml.Marshal(values)

	if err != nil {
		return nil, errors.Wrapf(err, "while marshalling chart values: [%v]", values)
	}
	resp, err := cli.helmClient().InstallReleaseFromChart(c, string(namespace),
		helm.InstallWait(true),
		helm.InstallTimeout(int64(installTimeout.Seconds())),
		helm.ValueOverrides(byteValues),
		helm.ReleaseName(string(releaseName)))
	if err != nil {
		return nil, errors.Wrapf(err, "while installing release from chart with name [%s] in namespace [%s]", releaseName, namespace)
	}
	return resp, nil
}

// Delete is deleting release of the chart
func (cli *Client) Delete(releaseName internal.ReleaseName) error {
	cli.log.WithField("purge", deleteWithPurge).Infof("Deleting chart with release name [%s]", releaseName)
	if _, err := cli.helmClient().DeleteRelease(string(releaseName), helm.DeletePurge(deleteWithPurge)); err != nil {
		return errors.Wrapf(err, "while deleting release name: [%s]", releaseName)
	}
	return nil
}

func (cli *Client) helmClient() helmDeleteInstaller {
	// helm client is not thread safe -
	//
	// helm.ConnectTimeout option is REQUIRED, because of this issue:
	// https://github.com/kubernetes/helm/issues/3658
	return helm.NewClient(helm.Host(cli.tillerHost), helm.ConnectTimeout(cli.tillerConnTimeout))
}
