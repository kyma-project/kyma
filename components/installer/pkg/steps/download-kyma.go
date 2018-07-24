package steps

import (
	"errors"
	"log"

	"github.com/kyma-project/kyma/components/installer/pkg/config"
)

//DownloadKyma .
func (steps InstallationSteps) DownloadKyma(installationData *config.InstallationData) error {
	const stepName string = "Downloading Kyma"
	steps.PrintInstallationStep(stepName)
	steps.statusManager.InProgress(stepName)

	if steps.kymaPackageClient.NeedDownload(steps.kymaPath) {

		if installationData.KymaVersion == "" {
			validationErr := errors.New("Set version for Kyma package")
			steps.errorHandlers.LogError("Validation error: ", validationErr)
			steps.statusManager.Error(stepName)
			return validationErr
		}

		if installationData.URL == "" {
			validationErr := errors.New("Set url to Kyma package")
			steps.errorHandlers.LogError("Validation error: ", validationErr)
			steps.statusManager.Error(stepName)
			return validationErr
		}

		log.Println("Downloading Kyma... Version: " + installationData.KymaVersion + " url: " + installationData.URL)

		err := steps.kymaPackageClient.CreateDir(steps.kymaPath)

		if steps.errorHandlers.CheckError("Mkdir error: ", err) {
			steps.statusManager.Error(stepName)
			return err
		}

		err = steps.kymaCommandExecutor.RunCommand("curl", "-Lks", installationData.URL, "-o", installationData.KymaVersion+".tar.gz")

		if steps.errorHandlers.CheckError("Download Kyma error: ", err) {
			steps.statusManager.Error(stepName)
			return err
		}

		err = steps.kymaCommandExecutor.RunCommand("tar", "xz", "-C", steps.kymaPath, "--strip-components=1", "-f", installationData.KymaVersion+".tar.gz")

		if steps.errorHandlers.CheckError("Unpack Kyma error: ", err) {
			steps.statusManager.Error(stepName)
			return err
		}
	} else {
		log.Println("Local Kyma sources provided. Download of sources not required.")
	}
	log.Println(stepName + "...DONE")
	return nil
}
