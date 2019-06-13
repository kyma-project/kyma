package steps

import (
	"log"

	"github.com/kyma-project/kyma/components/installer/pkg/config"
)

//RemoveKymaComponents .
func (steps InstallationSteps) RemoveKymaComponents(installationData *config.InstallationData) {
	log.Println("Uninstalling releases...")

	opts := []DeleteOption{DeletePurge(true)}
	for _, component := range installationData.Components {
		log.Println("Uninstalling release", component.GetReleaseName())
		uninstallReleaseResponse, uninstallError := steps.helmClient.DeleteRelease(component.GetReleaseName(), opts)

		if !steps.errorHandlers.CheckError("Uninstall Error: ", uninstallError) {
			steps.helmClient.PrintRelease(uninstallReleaseResponse.Release)
		}
	}

	log.Println("Uninstall finished")
}
