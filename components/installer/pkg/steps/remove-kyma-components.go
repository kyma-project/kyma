package steps

import (
	"log"
)

//RemoveKymaComponents .
func (steps InstallationSteps) RemoveKymaComponents() {

	log.Println("Removing Kyma resources...")

	if err := steps.uninstallReleases(); err != nil {
		log.Println("Error. Some releases may have not been removed.")
		return
	}

	if err := steps.removeNamespaces(); err != nil {
		log.Println("Error. Some namespaces may not have been removed")
		return
	}

	log.Println("All Kyma resources have been successfully removed")
}

func (steps InstallationSteps) uninstallReleases() error {
	log.Println("Uninstalling releases...")
	releasesToBeDeleted := []string{"ec-default", "hmc-default", "dex", "core", "prometheus-operator", "istio", "cluster-essentials"}

	for _, releaseName := range releasesToBeDeleted {
		log.Println("Uninstalling release", releaseName)
		uninstallReleaseResponse, uninstallError := steps.helmClient.DeleteRelease(releaseName)
		if steps.errorHandlers.CheckError("Uninstall Error: ", uninstallError) {
			return uninstallError
		}
		steps.helmClient.PrintRelease(uninstallReleaseResponse.Release)
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
