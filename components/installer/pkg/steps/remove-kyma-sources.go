package steps

import (
	"log"

	"github.com/kyma-project/kyma/components/installer/pkg/config"
)

//RemoveKymaSources .
func (steps *InstallationSteps) RemoveKymaSources(installationData *config.InstallationData) error {
	const stepName string = "Removing Kyma sources"
	steps.PrintInstallationStep(stepName)
	steps.statusManager.InProgress(stepName)

	err := steps.kymaPackageClient.RemoveDir(steps.kymaPath)

	if steps.errorHandlers.CheckError("Mkdir error: ", err) {
		steps.statusManager.Error(stepName)
		return err
	}

	log.Println(stepName + "...DONE")
	return nil
}
