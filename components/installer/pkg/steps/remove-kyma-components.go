package steps

import (
	"log"

	"github.com/kyma-project/kyma/components/installer/pkg/config"
	rls "k8s.io/helm/pkg/proto/hapi/services"
)

//RemoveKymaComponents .
func (steps InstallationSteps) RemoveKymaComponents(installationData *config.InstallationData) {

	log.Println("Removing Kyma resources...")

	if err := steps.uninstallReleases(installationData); err != nil {
		log.Println("Error. Some releases may have not been removed.")
		return
	}

	if err := steps.removeNamespaces(); err != nil {
		log.Println("Error. Some namespaces may not have been removed")
		return
	}

	log.Println("All Kyma resources have been successfully removed")
}

func (steps InstallationSteps) uninstallReleases(installationData *config.InstallationData) error {
	log.Println("Uninstalling releases...")

	releasesRes, err := steps.helmClient.ListReleases()
	if err != nil {
		return err
	}

	releasesToBeDeleted := getReleasesToBeDeleted(releasesRes, installationData)

	for _, releaseName := range releasesToBeDeleted {
		log.Println("Uninstalling release", releaseName)
		uninstallReleaseResponse, uninstallError := steps.helmClient.DeleteRelease(releaseName)
		if !steps.errorHandlers.CheckError("Uninstall Error: ", uninstallError) {
			steps.helmClient.PrintRelease(uninstallReleaseResponse.Release)
		}
	}

	log.Println("All releases have been successfully uninstalled!")
	return nil
}

func (steps InstallationSteps) removeNamespaces() error {
	log.Println("Removing namespaces")
	namespaceManager := steps.kubeClientset.CoreV1().Namespaces()
	namespacesToBeDeleted := []string{"kyma-integration", "kyma-system", "istio-system"}

	for _, namespace := range namespacesToBeDeleted {
		log.Println("Removing namespace", namespace)
		removeError := namespaceManager.Delete(namespace, nil)
		if steps.errorHandlers.CheckError("Remove namespace error: ", removeError) {
			return removeError
		}
		log.Println("Namespace", namespace, "has been successfully removed!")
	}
	log.Println("All namespaces have been successfully removed!")
	return nil
}

func getReleasesToBeDeleted(releasesRes *rls.ListReleasesResponse, installationData *config.InstallationData) []string {
	installedReleases := []string{}
	componentsPresentInCR := []string{}
	releasesToBeDeleted := []string{}

	if releasesRes != nil {
		for _, release := range releasesRes.Releases {
			installedReleases = append(installedReleases, release.Name)
		}

		for _, component := range installationData.Components {
			componentsPresentInCR = append(componentsPresentInCR, component.GetReleaseName())
		}

		for _, release := range installedReleases {
			if !stringSliceContainsString(componentsPresentInCR, release) {
				releasesToBeDeleted = append(releasesToBeDeleted, release)
			}
		}
	}

	if len(releasesToBeDeleted) == 0 {
		releasesToBeDeleted = installedReleases
	}

	return releasesToBeDeleted
}

func stringSliceContainsString(slice []string, str string) bool {
	for _, strInSlice := range slice {
		if str == strInSlice {
			return true
		}
	}
	return false
}
