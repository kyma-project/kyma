package steps

import (
	"errors"
	"log"

	"github.com/kyma-project/kyma/components/installer/pkg/config"
	"github.com/kyma-project/kyma/components/installer/pkg/kymasources"
)

//DownloadKyma .
func (steps InstallationSteps) EnsureKymaSources(installationData *config.InstallationData) (kymasources.KymaPackage, error) {
	const stepName string = "Get Kyma Sources"
	steps.PrintStep(stepName)
	steps.statusManager.InProgress(stepName)

	if steps.kymaPackages.HasInjectedSources() {
		log.Println("Kyma sources available locally.")
		log.Println(stepName + "...DONE")

		return steps.kymaPackages.GetInjectedPackage()
	}

	log.Println("Kyma sources not available. Downloading...")

	if installationData.KymaVersion == "" {
		validationErr := errors.New("Set version for Kyma package")
		steps.errorHandlers.LogError("Validation error: ", validationErr)
		steps.statusManager.Error(stepName)
		return nil, validationErr
	}

	if installationData.URL == "" {
		validationErr := errors.New("Set url to Kyma package")
		steps.errorHandlers.LogError("Validation error: ", validationErr)
		steps.statusManager.Error(stepName)
		return nil, validationErr
	}

	log.Println("Downloading Kyma, Version: " + installationData.KymaVersion + " URL: " + installationData.URL)

	err := steps.kymaPackages.FetchPackage(installationData.URL, installationData.KymaVersion)
	if steps.errorHandlers.CheckError("Fetch Kyma package error: ", err) {
		steps.statusManager.Error(stepName)
		return nil, err
	}

	kymaPackage, err := steps.kymaPackages.GetPackage(installationData.KymaVersion)
	if steps.errorHandlers.CheckError("Get Kyma package error: ", err) {
		steps.statusManager.Error(stepName)
		return nil, err
	}

	log.Println(stepName + "...DONE")

	return kymaPackage, nil
}
