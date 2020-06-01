package steps

import (
	"log"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/config"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymahelm"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/overrides"
	errors "github.com/pkg/errors"
	helm "k8s.io/helm/pkg/proto/hapi/release"
)

// stepFactoryCreator is used to create StepFactory instances for installation or uninstallation.
type stepFactoryCreator struct {
	helmClient          kymahelm.ClientInterface
	sourceGetterSupport SourceGetterLegacySupport
}

// NewStepFactoryCreator returns a new stepFactoryCreator instance.
func NewStepFactoryCreator(helmClient kymahelm.ClientInterface, sgls SourceGetterLegacySupport) *stepFactoryCreator {
	return &stepFactoryCreator{
		helmClient:          helmClient,
		sourceGetterSupport: sgls,
	}
}

func (sfc *stepFactoryCreator) getInstalledReleases() (map[string]kymahelm.ReleaseStatus, error) {

	existingReleases := make(map[string]kymahelm.ReleaseStatus)

	list, err := sfc.helmClient.ListReleases()
	if err != nil {
		return nil, errors.New("Helm error: " + err.Error())
	}

	if list != nil {
		log.Println("Helm releases list:")

		for _, release := range list.Releases {
			var lastDeployedRev int32

			statusCode := release.Info.Status.Code
			if statusCode == helm.Status_DEPLOYED {
				lastDeployedRev = release.Version
			} else {
				lastDeployedRev, err = sfc.helmClient.ReleaseDeployedRevision(release.Name)
				if err != nil {
					return nil, errors.New("Helm error: " + err.Error())
				}
			}

			log.Printf("%s status: %s, last deployed revision: %d", release.Name, statusCode, lastDeployedRev)
			existingReleases[release.Name] = kymahelm.ReleaseStatus{
				StatusCode:           statusCode,
				CurrentRevision:      release.Version,
				LastDeployedRevision: lastDeployedRev,
			}
		}
	}
	return existingReleases, nil
}

// ForInstallation returns implementation of StepFactory interface used to install or upgrade Kyma
func (sfc *stepFactoryCreator) ForInstallation(installationData *config.InstallationData, overrideData overrides.OverrideData) (StepLister, error) {

	installedReleases, err := sfc.getInstalledReleases()
	if err != nil {
		return nil, err
	}

	sourceGetter := sfc.sourceGetterSupport.SourceGetterFor(installationData.URL, installationData.KymaVersion)
	return &installStepFactory{
		stepFactory{sfc.helmClient, installedReleases, installationData},
		sourceGetter,
		overrideData,
	}, nil
}

// ForUninstallation returns implementation of StepFactory interface used to uninstall Kyma
func (sfc *stepFactoryCreator) ForUninstallation(installationData *config.InstallationData) (StepLister, error) {

	installedReleases, err := sfc.getInstalledReleases()
	if err != nil {
		return nil, err
	}

	return &uninstallStepFactory{
		stepFactory{sfc.helmClient, installedReleases, installationData},
	}, nil
}

//////////////////////////////////////////
//Code below should be removed ASAP
//////////////////////////////////////////

// SourceGetterLegacySupport exist only to support legacy, now deprecated, mechanism of fetching installation sources. Remove as soon as possible.
type SourceGetterLegacySupport interface {
	SourceGetterFor(kymaURL, kymaVersion string) SourceGetter
}

//TODO: Duplicated interface, see kymasources.SourceGetter
// SourceGetter defines contract for fetching component sources.
type SourceGetter interface {
	// SrcDirFor returns a local filesystem directory path to the component sources.
	SrcDirFor(component v1alpha1.KymaComponent) (string, error)
}
