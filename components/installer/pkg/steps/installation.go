package steps

import (
	"log"

	actionmanager "github.com/kyma-project/kyma/components/installer/pkg/actionmanager"
	"github.com/kyma-project/kyma/components/installer/pkg/config"
	internalerrors "github.com/kyma-project/kyma/components/installer/pkg/errors"
	"github.com/kyma-project/kyma/components/installer/pkg/kymahelm"
	"github.com/kyma-project/kyma/components/installer/pkg/kymainstallation"
	"github.com/kyma-project/kyma/components/installer/pkg/kymasources"
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
	statusManager       statusmanager.StatusManager
	actionManager       actionmanager.ActionManager
	kymaCommandExecutor toolkit.CommandExecutor
	kymaPackages        kymasources.KymaPackages

	currentPackage   kymasources.KymaPackage
	installedPackage kymasources.KymaPackage
}

// New .
func New(helmClient kymahelm.ClientInterface, kubeClientset *kubernetes.Clientset,
	serviceCatalog serviceCatalog.ClientInterface, statusManager statusmanager.StatusManager,
	actionManager actionmanager.ActionManager, kymaCommandExecutor toolkit.CommandExecutor,
	kymaPackages kymasources.KymaPackages) *InstallationSteps {
	steps := &InstallationSteps{
		helmClient:          helmClient,
		kubeClientset:       kubeClientset,
		serviceCatalog:      serviceCatalog,
		errorHandlers:       &internalerrors.ErrorHandlers{},
		statusManager:       statusManager,
		actionManager:       actionManager,
		kymaCommandExecutor: kymaCommandExecutor,
		kymaPackages:        kymaPackages,
	}

	return steps
}

//InstallKyma .
func (steps *InstallationSteps) InstallKyma(installationData *config.InstallationData, overrideData overrides.OverrideData) error {

	currentPackage, downloadKymaErr := steps.DownloadKyma(installationData)
	if downloadKymaErr != nil {
		return downloadKymaErr
	}

	steps.currentPackage = currentPackage

	stepsFactory := kymainstallation.NewStepFactory(currentPackage, steps.helmClient, overrides.NewLegacyProvider(overrideData, installationData, currentPackage, steps.errorHandlers))

	for _, component := range installationData.Components {

		stepName := "Installing " + component.GetReleaseName()
		steps.PrintInstallationStep(stepName)
		steps.statusManager.InProgress(stepName)

		if component.Name == "provision-bundles" { // Legacy support for bash provision-bundles script - to be deleted later
			err := steps.ProvisionBundles(installationData)
			if steps.errorHandlers.CheckError("Step installation error: ", err) {
				steps.statusManager.Error(stepName)
				return err
			}
		} else {
			step := stepsFactory.NewStep(component)
			steps.PrintInstallationStep(stepName)

			installErr := step.Install()
			if steps.errorHandlers.CheckError("Step installation error: ", installErr) {
				steps.statusManager.Error(stepName)
				return installErr
			}
		}

		log.Println(stepName + "...DONE")
	}

	err := steps.actionManager.RemoveActionLabel(installationData.Context.Name, installationData.Context.Namespace, "action")

	if steps.errorHandlers.CheckError("Error on removing label: ", err) {
		return err
	}

	err = steps.statusManager.InstallDone(installationData.URL, installationData.KymaVersion)
	if err != nil {
		return err
	}

	log.Println("Kyma Installed")

	return nil
}

//UpdateKyma .
func (steps *InstallationSteps) UpdateKyma(installationData *config.InstallationData, overrideData overrides.OverrideData) error {

	currentPackage, downloadKymaErr := steps.DownloadKyma(installationData)

	if downloadKymaErr != nil {
		return downloadKymaErr
	}

	steps.currentPackage = currentPackage

	stepsFactory := kymainstallation.NewStepFactory(currentPackage, steps.helmClient, overrides.NewLegacyProvider(overrideData, installationData, currentPackage, steps.errorHandlers))

	for _, component := range installationData.Components {

		stepName := "Upgrading " + component.GetReleaseName()
		steps.PrintInstallationStep(stepName)
		steps.statusManager.InProgress(stepName)

		if component.Name == "provision-bundles" { // Legacy support for bash provision-bundles script - to be deleted later
			err := steps.UpdateBundles(installationData)
			if steps.errorHandlers.CheckError("Step installation error: ", err) {
				steps.statusManager.Error(stepName)
				return err
			}
		} else {
			step := stepsFactory.NewStep(component)
			steps.PrintInstallationStep(stepName)

			installErr := step.Upgrade()
			if steps.errorHandlers.CheckError("Step installation error: ", installErr) {
				steps.statusManager.Error(stepName)
				return installErr
			}
		}

		log.Println(stepName + "...DONE")
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
	if installationData.IsLocalInstallation {
		return overrides.NewLocalStaticFile()
	}

	return overrides.NewClusterStaticFile(chartDir)
}
