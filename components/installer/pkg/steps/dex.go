package steps

import (
	"log"
	"path"

	"github.com/kyma-project/kyma/components/installer/pkg/config"
	"github.com/kyma-project/kyma/components/installer/pkg/consts"
	"github.com/kyma-project/kyma/components/installer/pkg/overrides"
)

//InstallDex installs Dex component
func (steps *InstallationSteps) InstallDex(installationData *config.InstallationData, overrideData overrides.OverrideData) error {

	const stepName string = "Installing Dex"
	const namespace string = "kyma-system"

	chartDir := path.Join(steps.currentPackage.GetChartsDirPath(), consts.DexComponent)
	dexOverrides, err := steps.getDexOverrides(installationData, chartDir, overrideData)

	if steps.errorHandlers.CheckError("Install Overrides Error: ", err) {
		steps.statusManager.Error(stepName)
		return err
	}

	steps.PrintInstallationStep(stepName)
	steps.statusManager.InProgress(stepName)

	installResp, installErr := steps.helmClient.InstallRelease(
		chartDir,
		namespace,
		consts.DexComponent,
		dexOverrides)

	if steps.errorHandlers.CheckError("Install Error: ", installErr) {
		steps.statusManager.Error(stepName)
		return installErr
	}

	steps.helmClient.PrintRelease(installResp.Release)
	log.Println(stepName + "...Done")

	return nil
}

// UpdateDex updates Dex component
func (steps *InstallationSteps) UpdateDex(installationData *config.InstallationData, overrideData overrides.OverrideData) error {

	const stepName string = "Updating Dex"
	const namespace string = "kyma-system"

	chartDir := path.Join(steps.currentPackage.GetChartsDirPath(), consts.DexComponent)
	dexOverrides, err := steps.getDexOverrides(installationData, chartDir, overrideData)

	if steps.errorHandlers.CheckError("Upgrade Overrides Error: ", err) {
		steps.statusManager.Error(stepName)
		return err
	}

	steps.PrintInstallationStep(stepName)
	steps.statusManager.InProgress(stepName)

	upgradeResp, upgradeErr := steps.helmClient.UpgradeRelease(
		chartDir,
		consts.DexComponent,
		dexOverrides)

	if steps.errorHandlers.CheckError("Upgrade Error: ", upgradeErr) {
		steps.statusManager.Error(stepName)
		return upgradeErr
	}

	steps.helmClient.PrintRelease(upgradeResp.Release)
	log.Println(stepName + "...Done")

	return nil
}

func (steps *InstallationSteps) getDexOverrides(installationData *config.InstallationData, chartDir string, overrideData overrides.OverrideData) (string, error) {

	allOverrides := overrides.Map{}
	overrides.MergeMaps(allOverrides, overrideData.Common())
	overrides.MergeMaps(allOverrides, overrideData.ForComponent(consts.DexComponent))

	globalOverrides, err := overrides.GetGlobalOverrides(installationData, allOverrides)
	if steps.errorHandlers.CheckError("Couldn't get global overrides: ", err) {
		return "", err
	}
	overrides.MergeMaps(allOverrides, globalOverrides)

	staticOverrides := steps.getStaticFileOverrides(installationData, chartDir)
	if staticOverrides.HasOverrides() == true {
		fileOverrides, err := staticOverrides.GetOverrides()
		if steps.errorHandlers.CheckError("Couldn't get additional overrides: ", err) {
			return "", err
		}
		overrides.MergeMaps(allOverrides, fileOverrides)
	}

	return overrides.ToYaml(allOverrides)
}
