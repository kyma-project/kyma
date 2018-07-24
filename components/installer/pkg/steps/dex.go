package steps

import (
	"log"
	"path"
	"strings"

	"github.com/kyma-project/kyma/components/installer/pkg/config"
	"github.com/kyma-project/kyma/components/installer/pkg/overrides"
)

//InstallDex installs Dex component
func (steps *InstallationSteps) InstallDex(installationData *config.InstallationData) error {

	const stepName string = "Installing Dex"
	const namespace string = "kyma-system"

	releaseName := "dex"
	chartDir := path.Join(steps.chartDir, releaseName)
	overrides := steps.getDexOverrides(installationData, chartDir)

	steps.PrintInstallationStep(stepName)
	steps.statusManager.InProgress(stepName)

	installResp, installErr := steps.helmClient.InstallRelease(
		chartDir,
		namespace,
		releaseName,
		overrides)

	if steps.errorHandlers.CheckError("Install Error: ", installErr) {
		steps.statusManager.Error(stepName)
		return installErr
	}

	steps.helmClient.PrintRelease(installResp.Release)
	log.Println(stepName + "...Done")

	return nil
}

// UpdateDex updates Dex component
func (steps *InstallationSteps) UpdateDex(installationData *config.InstallationData) error {

	const stepName string = "Updating Dex"
	const namespace string = "kyma-system"

	releaseName := "dex"
	chartDir := path.Join(steps.chartDir, releaseName)
	overrides := steps.getDexOverrides(installationData, chartDir)

	steps.PrintInstallationStep(stepName)
	steps.statusManager.InProgress(stepName)

	upgradeResp, upgradeErr := steps.helmClient.UpgradeRelease(
		chartDir,
		releaseName,
		overrides)

	if steps.errorHandlers.CheckError("Install Error: ", upgradeErr) {
		steps.statusManager.Error(stepName)
		return upgradeErr
	}

	steps.helmClient.PrintRelease(upgradeResp.Release)
	log.Println(stepName + "...Done")

	return nil
}

func (steps *InstallationSteps) getDexOverrides(installationData *config.InstallationData, chartDir string) string {

	var allOverrides []string

	globalOverrides, err := overrides.GetGlobalOverrides(installationData)
	steps.errorHandlers.LogError("Couldn't get global overrides: ", err)
	allOverrides = append(allOverrides, globalOverrides)

	fileOverrides := steps.getStaticFileOverrides(installationData, chartDir)
	if fileOverrides.HasOverrides() == true {
		fileOverridesStr, err := fileOverrides.GetOverrides()
		steps.errorHandlers.LogError("Couldn't get additional overrides: ", err)
		allOverrides = append(allOverrides, *fileOverridesStr)
	}

	return strings.Join(allOverrides, "\n")
}
