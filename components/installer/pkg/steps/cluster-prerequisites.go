package steps

import (
	"log"
	"path"

	"github.com/kyma-project/kyma/components/installer/pkg/config"
	"github.com/kyma-project/kyma/components/installer/pkg/toolkit"
)

//InstallClusterPrerequisites will install all needed before Kyma installation resources
func (steps *InstallationSteps) InstallClusterPrerequisites(installationData *config.InstallationData) error {
	const stepName string = "Installing cluster prerequisites"
	steps.PrintInstallationStep(stepName)

	steps.statusManager.InProgress(stepName)
	prerequisitesScriptPath := path.Join(steps.chartDir, "cluster-prerequisites/install.sh")
	err := steps.kymaCommandExecutor.RunCommand("/bin/bash", prerequisitesScriptPath)

	if steps.errorHandlers.CheckError("Script Error: ", err) {
		steps.statusManager.Error(stepName)
		return err
	}

	log.Println(stepName + "...DONE")

	return nil
}

//UpdateClusterPrerequisites will update all needed before Kyma installation resources
func (steps *InstallationSteps) UpdateClusterPrerequisites(installationData *config.InstallationData) error {
	const stepName string = "Updating cluster prerequisites"
	steps.PrintInstallationStep(stepName)

	steps.statusManager.InProgress(stepName)
	prerequisitesScriptPath := path.Join(steps.chartDir, "cluster-prerequisites/update.sh")
	err := steps.kymaCommandExecutor.RunBashCommand(prerequisitesScriptPath, toolkit.EmptyArgs)

	if steps.errorHandlers.CheckError("Script Error: ", err) {
		steps.statusManager.Error(stepName)
		return err
	}

	log.Println(stepName + "...DONE")

	return nil
}
