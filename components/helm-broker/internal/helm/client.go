package helm

import (
	"time"

	"github.com/ghodss/yaml"
	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/proto/hapi/release"
	hapi_release5 "k8s.io/helm/pkg/proto/hapi/release"
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

// ReleaseResponse keeps information about release after install or update helm release
type ReleaseResponse struct {
	Release *hapi_release5.Release `protobuf:"bytes,1,opt,name=release" json:"release,omitempty"`
}

// checks if release with conrete name and in namespace is already deployed
func (cli *Client) isReleaseDeployed(namespace internal.Namespace, releaseName internal.ReleaseName) (bool, error) {
	charts, err := cli.helmClient().ListReleases()
	if err != nil {
		return false, errors.Wrapf(err, "while finding release [%s] in namespace [%s] on releases list", releaseName, namespace)
	}

	for _, rel := range charts.GetReleases() {
		if string(namespace) != rel.Namespace {
			continue
		}
		if string(releaseName) != rel.Name {
			continue
		}
		if rel.Info.Status.GetCode() == release.Status_DEPLOYED {
			return true, nil
		}
	}

	return false, nil
}

// InstallOrUpdate is installing chart release or update release if it already exist
func (cli *Client) InstallOrUpdate(c *chart.Chart, values internal.ChartValues, releaseName internal.ReleaseName, namespace internal.Namespace) (*ReleaseResponse, error) {
	cli.log.Infof("Installing chart with release name [%s] in namespace [%s]", releaseName, namespace)
	byteValues, err := yaml.Marshal(values)

	if err != nil {
		return nil, errors.Wrapf(err, "while marshalling chart values: [%v]", values)
	}

	isDeployed, err := cli.isReleaseDeployed(namespace, releaseName)
	if err != nil {
		return nil, errors.Wrap(err, "while checking release is deployed already")
	}
	if isDeployed {
		resp, err := cli.helmClient().UpdateReleaseFromChart(string(releaseName), c)
		if err != nil {
			return nil, errors.Wrapf(err, "while updateing release from chart with name [%s] in namespace [%s]", releaseName, namespace)
		}
		return &ReleaseResponse{Release: resp.Release}, nil
	}

	resp, err := cli.helmClient().InstallReleaseFromChart(c, string(namespace),
		helm.InstallWait(true),
		helm.InstallTimeout(int64(installTimeout.Seconds())),
		helm.ValueOverrides(byteValues),
		helm.ReleaseName(string(releaseName)))
	if err != nil {
		return nil, errors.Wrapf(err, "while installing release from chart with name [%s] in namespace [%s]", releaseName, namespace)
	}
	return &ReleaseResponse{Release: resp.Release}, nil
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
