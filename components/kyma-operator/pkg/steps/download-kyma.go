package steps

import (
	"errors"
	"log"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/config"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymasources"
)

//DownloadKyma .
func (steps InstallationSteps) EnsureKymaSources(installationData *config.InstallationData) (kymasources.KymaPackage, error) {
	const stepName string = "Get Kyma Sources"
	steps.PrintStep(stepName)
	_ = steps.statusManager.InProgress(stepName)

	if steps.kymaPackages.HasInjectedSources() {
		log.Println("Kyma sources available locally.")
		log.Println(stepName + "...DONE")

		return steps.kymaPackages.GetInjectedPackage()
	}

	//TODO: Deprecated. Sources can be specified per component with `Source.URL` property.
	log.Println("Kyma sources not available. Downloading...")

	if installationData.KymaVersion == "" {
		validationErr := errors.New("set version for Kyma package")
		steps.errorHandlers.LogError("Validation error: ", validationErr)
		_ = steps.statusManager.Error("Kyma Operator", stepName, validationErr)
		return nil, validationErr
	}

	if installationData.URL == "" {
		validationErr := errors.New("set url to Kyma package")
		steps.errorHandlers.LogError("Validation error: ", validationErr)
		_ = steps.statusManager.Error("Kyma Operator", stepName, validationErr)
		return nil, validationErr
	}

	log.Println("Downloading Kyma, Version: " + installationData.KymaVersion + " URL: " + installationData.URL)

	err := steps.kymaPackages.FetchPackage(installationData.URL, installationData.KymaVersion)
	if steps.errorHandlers.CheckError("Fetch Kyma package error: ", err) {
		_ = steps.statusManager.Error("Kyma Operator", stepName, err)
		return nil, err
	}

	kymaPackage, err := steps.kymaPackages.GetPackage(installationData.KymaVersion)
	if steps.errorHandlers.CheckError("Get Kyma package error: ", err) {
		_ = steps.statusManager.Error("Kyma Operator", stepName, err)
		return nil, err
	}

	log.Println(stepName + "...DONE")

	return kymaPackage, nil
}
