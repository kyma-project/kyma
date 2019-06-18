package steps

import (
	"log"

	"github.com/kyma-project/kyma/components/installer/pkg/actionmanager"
	"github.com/kyma-project/kyma/components/installer/pkg/config"
	internalerrors "github.com/kyma-project/kyma/components/installer/pkg/errors"
	"github.com/kyma-project/kyma/components/installer/pkg/kymahelm"
	"github.com/kyma-project/kyma/components/installer/pkg/kymainstallation"
	"github.com/kyma-project/kyma/components/installer/pkg/kymasources"
	"github.com/kyma-project/kyma/components/installer/pkg/overrides"
	serviceCatalog "github.com/kyma-project/kyma/components/installer/pkg/servicecatalog"
	"github.com/kyma-project/kyma/components/installer/pkg/statusmanager"
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

	currentPackage, downloadKymaErr := steps.EnsureKymaSources(installationData)
	if downloadKymaErr != nil {
		return downloadKymaErr
	}
	steps.currentPackage = currentPackage

	err := steps.processComponents(installationData, overrideData, currentPackage.GetChartsDirPath())
	if steps.errorHandlers.CheckError("install/update error: ", err) {
		return err
	}

	err = steps.statusManager.InstallDone(installationData.URL, installationData.KymaVersion)
	if err != nil {
		return err
	}

	return nil
}

//UninstallKyma .
func (steps *InstallationSteps) UninstallKyma(installationData *config.InstallationData) error {

	err := steps.DeprovisionAzureResources(DefaultDeprovisionConfig(), installationData.Context)
	steps.errorHandlers.LogError("An error during deprovisioning: ", err)

	err = steps.processComponents(installationData, nil, "")
	if steps.errorHandlers.CheckError("uninstall error: ", err) {
		return err
	}

	err = steps.statusManager.UninstallDone()
	if err != nil {
		return err
	}

	return nil
}

//PrintStep .
func (steps *InstallationSteps) PrintStep(stepName string) {
	log.Println("---------------------------")
	log.Println(stepName)
	log.Println("---------------------------")
}

func (steps *InstallationSteps) processComponents(installationData *config.InstallationData, overrideData overrides.OverrideData, chartsPathDir string) error {

	_ = steps.statusManager.InProgress("Verify installed components")
	stepsFactory, factoryErr := kymainstallation.NewStepFactory(chartsPathDir, steps.helmClient, overrideData)
	if factoryErr != nil {
		_ = steps.statusManager.Error("installer", "Verify installed components", factoryErr)
		return factoryErr
	}

	log.Println("Processing Kyma components")

	logPrefix := installationData.Action

	for _, component := range installationData.Components {

		stepName := logPrefix + " component " + component.GetReleaseName()
		_ = steps.statusManager.InProgress(stepName)

		step := stepsFactory.NewStep(component)
		steps.PrintStep(stepName)

		processErr := step.Run()
		if steps.errorHandlers.CheckError("Step error: ", processErr) {
			_ = steps.statusManager.Error(component.GetReleaseName(), stepName, processErr)
			return processErr
		}

		log.Println(stepName + "...DONE!")
	}

	err := steps.actionManager.RemoveActionLabel(installationData.Context.Name, installationData.Context.Namespace, "action")
	if steps.errorHandlers.CheckError("Error on removing label: ", err) {
		return err
	}

	log.Println(logPrefix + " Kyma components ...DONE!")

	return nil
}
