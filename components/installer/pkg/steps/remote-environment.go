package steps

import (
	"log"
	"path"

	"github.com/kyma-project/kyma/components/installer/pkg/config"
	"github.com/kyma-project/kyma/components/installer/pkg/consts"
	"github.com/kyma-project/kyma/components/installer/pkg/overrides"
)

const (
	ecRemoteEnvironmentComponent  = "ec-default"
	hmcRemoteEnvironmentComponent = "hmc-default"
)

//InstallHmcDefaultRemoteEnvironments function will install Hmc Remote Environments
func (steps *InstallationSteps) InstallHmcDefaultRemoteEnvironments(installationData *config.InstallationData, overrideData overrides.OverrideData) error {
	const stepName string = "Installing Hmc Remote Environment"
	steps.PrintInstallationStep(stepName)

	steps.statusManager.InProgress(stepName)
	chartDir := path.Join(steps.currentPackage.GetChartsDirPath(), consts.RemoteEnvironments)
	hmcOverrides, err := steps.getHmcOverrides(installationData, chartDir, overrideData)

	if steps.errorHandlers.CheckError("Install Overrides Error: ", err) {
		steps.statusManager.Error(stepName)
		return err
	}

	installErr := steps.installRemoteEnvironment(hmcRemoteEnvironmentComponent, chartDir, hmcOverrides)

	if steps.errorHandlers.CheckError("Install Error: ", installErr) {
		steps.statusManager.Error(stepName)
		return installErr
	}

	log.Println(stepName + "...DONE")

	return nil
}

//UpdateHmcDefaultRemoteEnvironments function will install Hmc Remote Environments
func (steps *InstallationSteps) UpdateHmcDefaultRemoteEnvironments(installationData *config.InstallationData, overrideData overrides.OverrideData) error {
	const stepName string = "Updating Hmc Remote Environment"
	steps.PrintInstallationStep(stepName)

	steps.statusManager.InProgress(stepName)

	chartDir := path.Join(steps.currentPackage.GetChartsDirPath(), consts.RemoteEnvironments)
	hmcOverrides, err := steps.getHmcOverrides(installationData, chartDir, overrideData)
	if steps.errorHandlers.CheckError("Update Overrides Error: ", err) {
		steps.statusManager.Error(stepName)
		return err
	}

	upgradeErr := steps.updateRemoteEnvironment(hmcRemoteEnvironmentComponent, chartDir, hmcOverrides)

	if steps.errorHandlers.CheckError("Update Error: ", upgradeErr) {
		steps.statusManager.Error(stepName)
		return upgradeErr
	}

	log.Println(stepName + "...DONE")

	return nil
}

//InstallEcDefaultRemoteEnvironments function will install EC Remote Environments
func (steps *InstallationSteps) InstallEcDefaultRemoteEnvironments(installationData *config.InstallationData, overrideData overrides.OverrideData) error {
	const stepName string = "Installing Ec Remote Environment"
	steps.PrintInstallationStep(stepName)

	steps.statusManager.InProgress(stepName)
	chartDir := path.Join(steps.currentPackage.GetChartsDirPath(), consts.RemoteEnvironments)
	ecOverrides, err := steps.getEcOverrides(installationData, chartDir, overrideData)

	if steps.errorHandlers.CheckError("Install Overrides Error: ", err) {
		steps.statusManager.Error(stepName)
		return err
	}

	installErr := steps.installRemoteEnvironment(ecRemoteEnvironmentComponent, chartDir, ecOverrides)

	if steps.errorHandlers.CheckError("Install Error: ", installErr) {
		steps.statusManager.Error(stepName)
		return installErr
	}

	log.Println(stepName + "...DONE")

	return nil
}

//UpdateEcDefaultRemoteEnvironments function will install EC Remote Environments
func (steps *InstallationSteps) UpdateEcDefaultRemoteEnvironments(installationData *config.InstallationData, overrideData overrides.OverrideData) error {
	const stepName string = "Updating Ec Remote Environment"
	steps.PrintInstallationStep(stepName)

	steps.statusManager.InProgress(stepName)
	chartDir := path.Join(steps.currentPackage.GetChartsDirPath(), consts.RemoteEnvironments)
	ecOverrides, err := steps.getEcOverrides(installationData, chartDir, overrideData)

	if steps.errorHandlers.CheckError("Update Error: ", err) {
		steps.statusManager.Error(stepName)
		return err
	}

	upgradeErr := steps.updateRemoteEnvironment(ecRemoteEnvironmentComponent, chartDir, ecOverrides)

	if steps.errorHandlers.CheckError("Update Error: ", upgradeErr) {
		steps.statusManager.Error(stepName)
		return upgradeErr
	}

	log.Println(stepName + "...DONE")

	return nil
}

func (steps *InstallationSteps) getHmcOverrides(installationData *config.InstallationData, chartDir string, overrideData overrides.OverrideData) (string, error) {
	allOverrides := overrides.Map{}
	overrides.MergeMaps(allOverrides, overrideData.Common())
	overrides.MergeMaps(allOverrides, overrideData.ForComponent(hmcRemoteEnvironmentComponent))

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

func (steps *InstallationSteps) getEcOverrides(installationData *config.InstallationData, chartDir string, overrideData overrides.OverrideData) (string, error) {
	allOverrides := overrides.Map{}
	overrides.MergeMaps(allOverrides, overrideData.Common())
	overrides.MergeMaps(allOverrides, overrideData.ForComponent(ecRemoteEnvironmentComponent))

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

func (steps *InstallationSteps) installRemoteEnvironment(installationName, chartDir, overrides string) error {
	installResp, installErr := steps.helmClient.InstallRelease(
		chartDir,
		"kyma-integration",
		installationName,
		overrides)

	if installErr != nil {
		return installErr
	}

	steps.helmClient.PrintRelease(installResp.Release)
	return nil
}

func (steps *InstallationSteps) updateRemoteEnvironment(installationName, chartDir, overrides string) error {
	upgradeResp, upgradeErr := steps.helmClient.UpgradeRelease(
		chartDir,
		installationName,
		overrides)

	if upgradeErr != nil {
		return upgradeErr
	}

	steps.helmClient.PrintRelease(upgradeResp.Release)
	return nil
}
