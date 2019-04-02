package testkit

import (
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/release"
	rls "k8s.io/helm/pkg/proto/hapi/services"
	"k8s.io/helm/pkg/tlsutil"

	"time"
)

type HelmClient interface {
	CheckReleaseStatus(rlsName string) (*rls.GetReleaseStatusResponse, error)
	CheckReleaseExistence(name string) (bool, error)
	IsInstalled(rlsName string) bool
	TestRelease(rlsName string) (<-chan *rls.TestReleaseResponse, <-chan error)
}

type helmClient struct {
	helm          *helm.Client
	retryCount    int
	retryWaitTime time.Duration
}

func NewHelmClient(host, tlsKeyFile, tlsCertFile string, skipVerify bool) (HelmClient, error) {
	tlsOpts := tlsutil.Options{
		KeyFile:            tlsKeyFile,
		CertFile:           tlsCertFile,
		InsecureSkipVerify: skipVerify,
	}

	tlsCfg, err := tlsutil.ClientConfig(tlsOpts)
	if err != nil {
		return nil, err
	}

	return &helmClient{
		helm: helm.NewClient(helm.Host(host), helm.WithTLS(tlsCfg)),
	}, nil
}

func (hc *helmClient) CheckReleaseStatus(rlsName string) (*rls.GetReleaseStatusResponse, error) {
	return hc.helm.ReleaseStatus(rlsName)
}

func (hc *helmClient) IsInstalled(rlsName string) bool {
	status, err := hc.CheckReleaseStatus(rlsName)
	return err == nil && status.Info.Status.Code == release.Status_DEPLOYED
}

func (hc *helmClient) CheckReleaseExistence(rlsName string) (bool, error) {
	listResponse, err := hc.helm.ListReleases(helm.ReleaseListStatuses([]release.Status_Code{
		release.Status_DELETED,
		release.Status_DELETING,
		release.Status_DEPLOYED,
		release.Status_FAILED,
		release.Status_PENDING_INSTALL,
		release.Status_PENDING_ROLLBACK,
		release.Status_PENDING_UPGRADE,
		release.Status_SUPERSEDED,
		release.Status_UNKNOWN,
	}))
	if err != nil {
		return false, err
	}

	for _, rel := range listResponse.Releases {
		if rel.Name == rlsName {
			return true, nil
		}
	}
	return false, nil
}

func (hc *helmClient) TestRelease(rlsName string) (<-chan *rls.TestReleaseResponse, <-chan error) {
	return hc.helm.RunReleaseTest(rlsName)
}
