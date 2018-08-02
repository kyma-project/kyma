package steps

import (
	"log"
	"path"

	actionmanager "github.com/kyma-project/kyma/components/installer/pkg/actionmanager"
	"github.com/kyma-project/kyma/components/installer/pkg/config"
	"github.com/kyma-project/kyma/components/installer/pkg/consts"
	internalerrors "github.com/kyma-project/kyma/components/installer/pkg/errors"
	"github.com/kyma-project/kyma/components/installer/pkg/kymahelm"
	"github.com/kyma-project/kyma/components/installer/pkg/overrides"
	serviceCatalog "github.com/kyma-project/kyma/components/installer/pkg/servicecatalog"
	statusmanager "github.com/kyma-project/kyma/components/installer/pkg/statusmanager"
	"github.com/kyma-project/kyma/components/installer/pkg/toolkit"
	"k8s.io/client-go/kubernetes"
)

//InstallationSteps .
type InstallationSteps struct {
	helmClient          kymahelm.ClientInterface
	kubeClientset       *kubernetes.Clientset
	serviceCatalog      serviceCatalog.ClientInterface
	errorHandlers       internalerrors.ErrorHandlersInterface
	chartDir            string
	statusManager       statusmanager.StatusManager
	actionManager       actionmanager.ActionManager
	kymaCommandExecutor toolkit.CommandExecutor
	kymaPath            string
	kymaPackageClient   KymaPackageInterface
}

// New .
func New(helmClient kymahelm.ClientInterface, kubeClientset *kubernetes.Clientset, serviceCatalog serviceCatalog.ClientInterface, kymaDir string, statusManager statusmanager.StatusManager, actionManager actionmanager.ActionManager, kymaCommandExecutor toolkit.CommandExecutor, kymaPackageClient KymaPackageInterface) *InstallationSteps {
	steps := &InstallationSteps{
		helmClient:          helmClient,
		kubeClientset:       kubeClientset,
		serviceCatalog:      serviceCatalog,
		errorHandlers:       &internalerrors.ErrorHandlers{},
		chartDir:            path.Join(kymaDir, "resources"),
		statusManager:       statusManager,
		actionManager:       actionManager,
		kymaCommandExecutor: kymaCommandExecutor,
		kymaPath:            kymaDir,
		kymaPackageClient:   kymaPackageClient,
	}

	return steps
}

//InstallKyma .
func (steps *InstallationSteps) InstallKyma(installationData *config.InstallationData) error {

	downloadKymaErr := steps.DownloadKyma(installationData)
	if downloadKymaErr != nil {
		return downloadKymaErr
	}

	if installationData.ShouldInstallComponent(consts.ClusterPrerequisitesComponent) {
		if instErr := steps.InstallClusterPrerequisites(installationData); instErr != nil {
			return instErr
		}
	}

	if installationData.ShouldInstallComponent(consts.ClusterEssentialsComponent) {
		if instErr := steps.InstallClusterEssentials(installationData); instErr != nil {
			return instErr
		}
	}

	if installationData.ShouldInstallComponent(consts.IstioComponent) {
		if instErr := steps.InstallIstio(installationData); instErr != nil {
			return instErr
		}
	}
	if installationData.ShouldInstallComponent(consts.PrometheusComponent) {
		if instErr := steps.InstallPrometheus(installationData); instErr != nil {
			return instErr
		}
	}

	if installationData.ShouldInstallComponent(consts.ProvisionBundlesComponent) {
		if bundlesErr := steps.ProvisionBundles(installationData); bundlesErr != nil {
			return bundlesErr
		}
	}

	if installationData.ShouldInstallComponent(consts.DexComponent) {
		if dexErr := steps.InstallDex(installationData); dexErr != nil {
			return dexErr
		}
	}

	if installationData.ShouldInstallComponent(consts.CoreComponent) {
		if instErr := steps.InstallCore(installationData); instErr != nil {
			return instErr
		}

		if upgradeErr := steps.UpgradeCore(installationData); upgradeErr != nil {
			return upgradeErr
		}
	}

	if installationData.ShouldInstallComponent(consts.RemoteEnvironments) {
		if instErr := steps.InstallHmcDefaultRemoteEnvironments(installationData); instErr != nil {
			return instErr
		}

		if instErr := steps.InstallEcDefaultRemoteEnvironments(installationData); instErr != nil {
			return instErr
		}
	}

	if instErr := steps.RemoveKymaSources(installationData); instErr != nil {
		return instErr
	}

	err := steps.actionManager.RemoveActionLabel(installationData.Context.Name, installationData.Context.Namespace, "action")
	if steps.errorHandlers.CheckError("Error on removing label: ", err) {
		return err
	}

	err = steps.statusManager.InstallDone(installationData.URL, installationData.KymaVersion)
	if err != nil {
		return err
	}

	return nil
}

//UpdateKyma .
func (steps *InstallationSteps) UpdateKyma(installationData *config.InstallationData) error {
	downloadKymaErr := steps.DownloadKyma(installationData)
	if downloadKymaErr != nil {
		return downloadKymaErr
	}

	upgradeErr := steps.UpdateClusterPrerequisites(installationData)
	if upgradeErr != nil {
		return upgradeErr
	}

	upgradeErr = steps.UpdateClusterEssentials(installationData)
	if upgradeErr != nil {
		return upgradeErr
	}

	upgradeErr = steps.UpdateIstio(installationData)
	if upgradeErr != nil {
		return upgradeErr
	}

	upgradeErr = steps.UpdatePrometheus(installationData)
	if upgradeErr != nil {
		return upgradeErr
	}

	bundlesErr := steps.UpdateBundles(installationData)
	if bundlesErr != nil {
		return bundlesErr
	}

	dexErr := steps.UpdateDex(installationData)
	if dexErr != nil {
		return dexErr
	}

	upgradeErr = steps.UpgradeCore(installationData)
	if upgradeErr != nil {
		return upgradeErr
	}

	upgradeErr = steps.UpdateHmcDefaultRemoteEnvironments(installationData)
	if upgradeErr != nil {
		return upgradeErr
	}

	upgradeErr = steps.UpdateEcDefaultRemoteEnvironments(installationData)
	if upgradeErr != nil {
		return upgradeErr
	}

	upgradeErr = steps.RemoveKymaSources(installationData)
	if upgradeErr != nil {
		return upgradeErr
	}

	err := steps.actionManager.RemoveActionLabel(installationData.Context.Name, installationData.Context.Namespace, "action")
	if steps.errorHandlers.CheckError("Error on removing label: ", err) {
		return err
	}

	err = steps.statusManager.UpdateDone(installationData.URL, installationData.KymaVersion)
	if err != nil {
		return err
	}

	return nil
}

//UninstallKyma .
func (steps *InstallationSteps) UninstallKyma(installationData *config.InstallationData) error {
	err := steps.DeprovisionAzureResources(DefaultDeprovisionConfig(), installationData.Context)
	steps.errorHandlers.LogError("An error during deprovisioning: ", err)
	steps.RemoveKymaComponents()

	err = steps.actionManager.RemoveActionLabel(installationData.Context.Name, installationData.Context.Namespace, "action")
	if steps.errorHandlers.CheckError("Error on removing label: ", err) {
		return err
	}

	err = steps.statusManager.UninstallDone()
	if err != nil {
		return err
	}

	return nil
}

//PrintInstallationStep .
func (steps *InstallationSteps) PrintInstallationStep(stepName string) {
	log.Println("---------------------------")
	log.Println(stepName)
	log.Println("---------------------------")
}

func (steps *InstallationSteps) getStaticFileOverrides(installationData *config.InstallationData, chartDir string) overrides.StaticFile {
	if installationData.IsLocalInstallation() == true {
		return overrides.NewLocalStaticFile()
	}

	return overrides.NewClusterStaticFile(chartDir)
}
