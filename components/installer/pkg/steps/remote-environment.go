package steps

import (
	"log"
	"path"
	"strings"

	"github.com/kyma-project/kyma/components/installer/pkg/config"
	"github.com/kyma-project/kyma/components/installer/pkg/consts"
	"github.com/kyma-project/kyma/components/installer/pkg/overrides"
)

const (
	ecRemoteEnvironmentComponent  = "ec-default"
	hmcRemoteEnvironmentComponent = "hmc-default"
)

//InstallHmcDefaultRemoteEnvironments function will install Hmc Remote Environments
func (steps *InstallationSteps) InstallHmcDefaultRemoteEnvironments(installationData *config.InstallationData) error {
	const stepName string = "Installing Hmc Remote Environment"
	steps.PrintInstallationStep(stepName)

	steps.statusManager.InProgress(stepName)
	chartDir := path.Join(steps.chartDir, consts.RemoteEnvironments)
	hmcOverrides := steps.getHmcOverrides(installationData, chartDir)

	installErr := steps.installRemoteEnvironment(hmcRemoteEnvironmentComponent, chartDir, hmcOverrides)

	if steps.errorHandlers.CheckError("Install Error: ", installErr) {
		steps.statusManager.Error(stepName)
		return installErr
	}

	log.Println(stepName + "...DONE")

	return nil
}

//UpdateHmcDefaultRemoteEnvironments function will install Hmc Remote Environments
func (steps *InstallationSteps) UpdateHmcDefaultRemoteEnvironments(installationData *config.InstallationData) error {
	const stepName string = "Updating Hmc Remote Environment"
	steps.PrintInstallationStep(stepName)

	steps.statusManager.InProgress(stepName)

	chartDir := path.Join(steps.chartDir, consts.RemoteEnvironments)
	hmcOverrides := steps.getHmcOverrides(installationData, chartDir)

	upgradeErr := steps.updateRemoteEnvironment(hmcRemoteEnvironmentComponent, chartDir, hmcOverrides)

	if steps.errorHandlers.CheckError("Update Error: ", upgradeErr) {
		steps.statusManager.Error(stepName)
		return upgradeErr
	}

	log.Println(stepName + "...DONE")

	return nil
}

//InstallEcDefaultRemoteEnvironments function will install EC Remote Environments
func (steps *InstallationSteps) InstallEcDefaultRemoteEnvironments(installationData *config.InstallationData) error {
	const stepName string = "Installing Ec Remote Environment"
	steps.PrintInstallationStep(stepName)

	steps.statusManager.InProgress(stepName)
	chartDir := path.Join(steps.chartDir, consts.RemoteEnvironments)
	ecOverrides := steps.getEcOverrides(installationData, chartDir)

	installErr := steps.installRemoteEnvironment(ecRemoteEnvironmentComponent, chartDir, ecOverrides)

	if steps.errorHandlers.CheckError("Install Error: ", installErr) {
		steps.statusManager.Error(stepName)
		return installErr
	}

	log.Println(stepName + "...DONE")

	return nil
}

//UpdateEcDefaultRemoteEnvironments function will install EC Remote Environments
func (steps *InstallationSteps) UpdateEcDefaultRemoteEnvironments(installationData *config.InstallationData) error {
	const stepName string = "Updating Ec Remote Environment"
	steps.PrintInstallationStep(stepName)

	steps.statusManager.InProgress(stepName)
	chartDir := path.Join(steps.chartDir, consts.RemoteEnvironments)
	ecOverrides := steps.getEcOverrides(installationData, chartDir)

	upgradeErr := steps.updateRemoteEnvironment(ecRemoteEnvironmentComponent, chartDir, ecOverrides)

	if steps.errorHandlers.CheckError("Update Error: ", upgradeErr) {
		steps.statusManager.Error(stepName)
		return upgradeErr
	}

	log.Println(stepName + "...DONE")

	return nil
}

func (steps *InstallationSteps) getHmcOverrides(installationData *config.InstallationData, chartDir string) string {
	var allOverrides []string

	globalOverrides, err := overrides.GetGlobalOverrides(installationData)
	steps.errorHandlers.LogError("Couldn't get global overrides: ", err)
	allOverrides = append(allOverrides, globalOverrides)

	hmcDefaultOverride := overrides.GetHmcDefaultOverrides()
	steps.errorHandlers.LogError("Couldn't get Hmc default overrides: ", err)
	allOverrides = append(allOverrides, hmcDefaultOverride)

	fileOverrides := steps.getStaticFileOverrides(installationData, chartDir)
	if fileOverrides.HasOverrides() == true {
		fileOverridesStr, err := fileOverrides.GetOverrides()
		steps.errorHandlers.LogError("Couldn't get additional overrides: ", err)
		allOverrides = append(allOverrides, *fileOverridesStr)
	}

	return strings.Join(allOverrides, "\n")
}

func (steps *InstallationSteps) getEcOverrides(installationData *config.InstallationData, chartDir string) string {
	var allOverrides []string

	globalOverrides, err := overrides.GetGlobalOverrides(installationData)
	steps.errorHandlers.LogError("Couldn't get global overrides: ", err)
	allOverrides = append(allOverrides, globalOverrides)

	ecDefaultOverride := overrides.GetEcDefaultOverrides()
	steps.errorHandlers.LogError("Couldn't get Ec default overrides: ", err)
	allOverrides = append(allOverrides, ecDefaultOverride)

	fileOverrides := steps.getStaticFileOverrides(installationData, chartDir)
	if fileOverrides.HasOverrides() == true {
		fileOverridesStr, err := fileOverrides.GetOverrides()
		steps.errorHandlers.LogError("Couldn't get additional overrides: ", err)
		allOverrides = append(allOverrides, *fileOverridesStr)
	}

	return strings.Join(allOverrides, "\n")
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
