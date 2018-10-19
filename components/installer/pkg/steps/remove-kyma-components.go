package steps

import (
	"log"

	"github.com/kyma-project/kyma/components/installer/pkg/config"
)

//RemoveKymaComponents .
func (steps InstallationSteps) RemoveKymaComponents(installationData *config.InstallationData) {

	log.Println("Removing Kyma resources...")

	if err := steps.uninstallReleases(installationData); err != nil {
		log.Println("Error. Some releases may have not been removed.")
		return
	}

	log.Println("All Kyma resources have been successfully removed")
}

func (steps InstallationSteps) uninstallReleases(installationData *config.InstallationData) error {
	log.Println("Uninstalling releases...")
	installedReleases := make(map[string]bool)

	releasesRes, err := steps.helmClient.ListReleases()
	if err != nil {
		return err
	}

	if releasesRes != nil {
		for _, release := range releasesRes.Releases {
			installedReleases[release.Name] = true
		}
	}

	for _, component := range installationData.Components {
		if _, ok := installedReleases[component.GetReleaseName()]; ok {
			log.Println("Uninstalling release", component.GetReleaseName())
			uninstallReleaseResponse, uninstallError := steps.helmClient.DeleteRelease(component.GetReleaseName())

			if !steps.errorHandlers.CheckError("Uninstall Error: ", uninstallError) {
				steps.helmClient.PrintRelease(uninstallReleaseResponse.Release)
			}
		} else {
			log.Println(component.GetReleaseName(), "not found")
		}
	}

	log.Println("All releases have been successfully uninstalled!")
	return nil
}
