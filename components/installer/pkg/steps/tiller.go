package steps

import (
	"log"
	"path"

	"github.com/kyma-project/kyma/components/installer/pkg/config"
	"github.com/kyma-project/kyma/components/installer/pkg/toolkit"
)

// InstallTiller will install Tiller
func (steps InstallationSteps) InstallTiller(installationData *config.InstallationData) error {
	const stepName string = "Installing Tiller"
	steps.PrintInstallationStep(stepName)

	steps.statusManager.InProgress(stepName)
	tillerScriptPath := path.Join(steps.chartDir, "tiller/install.sh")
	err := steps.kymaCommandExecutor.RunCommand("/bin/bash", tillerScriptPath)

	if steps.errorHandlers.CheckError("Script error: ", err) {
		steps.statusManager.Error(stepName)
		return err
	}

	log.Println(stepName + "...DONE")

	return nil
}

// UpdateTiller will update Tiller
func (steps InstallationSteps) UpdateTiller(installationData *config.InstallationData) error {
	const stepName string = "Updating Tiller"
	steps.PrintInstallationStep(stepName)

	steps.statusManager.InProgress(stepName)
	tillerScriptPath := path.Join(steps.chartDir, "tiller/update.sh")
	err := steps.kymaCommandExecutor.RunBashCommand(tillerScriptPath, toolkit.EmptyArgs)

	if steps.errorHandlers.CheckError("Script error: ", err) {
		steps.statusManager.Error(stepName)
		return err
	}

	log.Println(stepName + "...DONE")

	return nil
}
