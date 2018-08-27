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
func (steps *InstallationSteps) InstallHmcDefaultRemoteEnvironments(installationData *config.InstallationData, overrideData OverrideData) error {
	const stepName string = "Installing Hmc Remote Environment"
	steps.PrintInstallationStep(stepName)

	steps.statusManager.InProgress(stepName)
	chartDir := path.Join(steps.currentPackage.GetChartsDirPath(), consts.RemoteEnvironments)
	hmcOverrides, err := steps.getHmcOverrides(installationData, chartDir)

	if steps.errorHandlers.CheckError("Install Overrides Error: ", err) {
		steps.statusManager.Error(stepName)
		return err
	}

	allOverrides := overrides.Map{}
	overrides.MergeMaps(allOverrides, overrideData.Common())
	overrides.MergeMaps(allOverrides, hmcOverrides)                                         //TODO: Remove after migration to generic overrides is completed
	overrides.MergeMaps(allOverrides, overrideData.ForComponent(consts.RemoteEnvironments)) //TODO: We need two different components for Hmc and Ec!

	overridesStr, err := overrides.ToYaml(allOverrides)

	if steps.errorHandlers.CheckError("Install Overrides Error: ", err) {
		steps.statusManager.Error(stepName)
		return err
	}

	installErr := steps.installRemoteEnvironment(hmcRemoteEnvironmentComponent, chartDir, overridesStr)

	if steps.errorHandlers.CheckError("Install Error: ", installErr) {
		steps.statusManager.Error(stepName)
		return installErr
	}

	log.Println(stepName + "...DONE")

	return nil
}

//UpdateHmcDefaultRemoteEnvironments function will install Hmc Remote Environments
func (steps *InstallationSteps) UpdateHmcDefaultRemoteEnvironments(installationData *config.InstallationData, overrideData OverrideData) error {
	const stepName string = "Updating Hmc Remote Environment"
	steps.PrintInstallationStep(stepName)

	steps.statusManager.InProgress(stepName)

	chartDir := path.Join(steps.currentPackage.GetChartsDirPath(), consts.RemoteEnvironments)
	hmcOverrides, err := steps.getHmcOverrides(installationData, chartDir)
	if steps.errorHandlers.CheckError("Update Overrides Error: ", err) {
		steps.statusManager.Error(stepName)
		return err
	}

	allOverrides := overrides.Map{}
	overrides.MergeMaps(allOverrides, overrideData.Common())
	overrides.MergeMaps(allOverrides, hmcOverrides)                                         //TODO: Remove after migration to generic overrides is completed
	overrides.MergeMaps(allOverrides, overrideData.ForComponent(consts.RemoteEnvironments)) //TODO: We need two different components for Hmc and Ec!

	overridesStr, err := overrides.ToYaml(allOverrides)

	if steps.errorHandlers.CheckError("Upgrade Overrides Error: ", err) {
		steps.statusManager.Error(stepName)
		return err
	}

	upgradeErr := steps.updateRemoteEnvironment(hmcRemoteEnvironmentComponent, chartDir, overridesStr)

	if steps.errorHandlers.CheckError("Update Error: ", upgradeErr) {
		steps.statusManager.Error(stepName)
		return upgradeErr
	}

	log.Println(stepName + "...DONE")

	return nil
}

//InstallEcDefaultRemoteEnvironments function will install EC Remote Environments
func (steps *InstallationSteps) InstallEcDefaultRemoteEnvironments(installationData *config.InstallationData, overrideData OverrideData) error {
	const stepName string = "Installing Ec Remote Environment"
	steps.PrintInstallationStep(stepName)

	steps.statusManager.InProgress(stepName)
	chartDir := path.Join(steps.currentPackage.GetChartsDirPath(), consts.RemoteEnvironments)
	ecOverrides, err := steps.getEcOverrides(installationData, chartDir)

	if steps.errorHandlers.CheckError("Install Overrides Error: ", err) {
		steps.statusManager.Error(stepName)
		return err
	}

	allOverrides := overrides.Map{}
	overrides.MergeMaps(allOverrides, overrideData.Common())
	overrides.MergeMaps(allOverrides, ecOverrides)                                          //TODO: Remove after migration to generic overrides is completed
	overrides.MergeMaps(allOverrides, overrideData.ForComponent(consts.RemoteEnvironments)) //TODO: We need two different components for Hmc and Ec!

	overridesStr, err := overrides.ToYaml(allOverrides)

	if steps.errorHandlers.CheckError("Install Overrides Error: ", err) {
		steps.statusManager.Error(stepName)
		return err
	}

	installErr := steps.installRemoteEnvironment(ecRemoteEnvironmentComponent, chartDir, overridesStr)

	if steps.errorHandlers.CheckError("Install Error: ", installErr) {
		steps.statusManager.Error(stepName)
		return installErr
	}

	log.Println(stepName + "...DONE")

	return nil
}

//UpdateEcDefaultRemoteEnvironments function will install EC Remote Environments
func (steps *InstallationSteps) UpdateEcDefaultRemoteEnvironments(installationData *config.InstallationData, overrideData OverrideData) error {
	const stepName string = "Updating Ec Remote Environment"
	steps.PrintInstallationStep(stepName)

	steps.statusManager.InProgress(stepName)
	chartDir := path.Join(steps.currentPackage.GetChartsDirPath(), consts.RemoteEnvironments)
	ecOverrides, err := steps.getEcOverrides(installationData, chartDir)

	if steps.errorHandlers.CheckError("Update Error: ", err) {
		steps.statusManager.Error(stepName)
		return err
	}

	allOverrides := overrides.Map{}
	overrides.MergeMaps(allOverrides, overrideData.Common())
	overrides.MergeMaps(allOverrides, ecOverrides)                                          //TODO: Remove after migration to generic overrides is completed
	overrides.MergeMaps(allOverrides, overrideData.ForComponent(consts.RemoteEnvironments)) //TODO: We need two different components for Hmc and Ec!

	overridesStr, err := overrides.ToYaml(allOverrides)

	if steps.errorHandlers.CheckError("Install Upgrade Error: ", err) {
		steps.statusManager.Error(stepName)
		return err
	}

	upgradeErr := steps.updateRemoteEnvironment(ecRemoteEnvironmentComponent, chartDir, overridesStr)

	if steps.errorHandlers.CheckError("Update Error: ", upgradeErr) {
		steps.statusManager.Error(stepName)
		return upgradeErr
	}

	log.Println(stepName + "...DONE")

	return nil
}

func (steps *InstallationSteps) getHmcOverrides(installationData *config.InstallationData, chartDir string) (overrides.Map, error) {
	allOverrides := overrides.Map{}

	globalOverrides, err := overrides.GetGlobalOverrides(installationData)
	if steps.errorHandlers.CheckError("Couldn't get global overrides: ", err) {
		return nil, err
	}
	overrides.MergeMaps(allOverrides, globalOverrides)

	hmcDefaultOverride, err := overrides.GetHmcDefaultOverrides()

	if steps.errorHandlers.CheckError("Couldn't get Hmc default overrides: ", err) {
		return nil, err
	}
	overrides.MergeMaps(allOverrides, hmcDefaultOverride)

	staticOverrides := steps.getStaticFileOverrides(installationData, chartDir)
	if staticOverrides.HasOverrides() == true {
		fileOverrides, err := staticOverrides.GetOverrides()
		if steps.errorHandlers.CheckError("Couldn't get additional overrides: ", err) {
			return nil, err
		}
		overrides.MergeMaps(allOverrides, fileOverrides)
	}

	return allOverrides, nil
}

func (steps *InstallationSteps) getEcOverrides(installationData *config.InstallationData, chartDir string) (overrides.Map, error) {
	allOverrides := overrides.Map{}

	globalOverrides, err := overrides.GetGlobalOverrides(installationData)
	if steps.errorHandlers.CheckError("Couldn't get global overrides: ", err) {
		return nil, err
	}
	overrides.MergeMaps(allOverrides, globalOverrides)

	ecDefaultOverride, err := overrides.GetEcDefaultOverrides()
	if steps.errorHandlers.CheckError("Couldn't get Ec default overrides: ", err) {
		return nil, err
	}
	overrides.MergeMaps(allOverrides, ecDefaultOverride)

	staticOverrides := steps.getStaticFileOverrides(installationData, chartDir)
	if staticOverrides.HasOverrides() == true {
		fileOverrides, err := staticOverrides.GetOverrides()
		if steps.errorHandlers.CheckError("Couldn't get additional overrides: ", err) {
			return nil, err
		}
		overrides.MergeMaps(allOverrides, fileOverrides)
	}

	return allOverrides, nil
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
