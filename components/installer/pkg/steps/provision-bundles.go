package steps

import (
	"log"
	"path"

	"github.com/kyma-project/kyma/components/installer/pkg/config"
	"github.com/kyma-project/kyma/components/installer/pkg/toolkit"
)

// ProvisionBundles .
func (steps InstallationSteps) ProvisionBundles(installationData *config.InstallationData) error {
	const stepName string = "Provisioning helm broker repo bundles"
	steps.PrintInstallationStep(stepName)
	steps.statusManager.InProgress(stepName)

	helmBrokerScriptPath := path.Join(steps.chartDir, "helm-broker-repo/install.sh")

	err := steps.kymaCommandExecutor.RunCommand("/bin/bash", helmBrokerScriptPath)

	if steps.errorHandlers.CheckError("Script Error: ", err) {
		steps.statusManager.Error(stepName)
		return err
	}

	log.Println(stepName + "...DONE")

	return nil
}

// UpdateBundles .
func (steps InstallationSteps) UpdateBundles(installationData *config.InstallationData) error {
	const stepName string = "Updating helm broker repo bundles"
	steps.PrintInstallationStep(stepName)
	steps.statusManager.InProgress(stepName)

	helmBrokerScriptPath := path.Join(steps.chartDir, "helm-broker-repo/update.sh")

	err := steps.kymaCommandExecutor.RunBashCommand(helmBrokerScriptPath, toolkit.EmptyArgs)

	if steps.errorHandlers.CheckError("Script Error: ", err) {
		steps.statusManager.Error(stepName)
		return err
	}

	log.Println(stepName + "...DONE")

	return nil
}
