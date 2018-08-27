package steps

import (
	"log"
	"path"

	"github.com/kyma-project/kyma/components/installer/pkg/config"
	"github.com/kyma-project/kyma/components/installer/pkg/consts"
	"github.com/kyma-project/kyma/components/installer/pkg/overrides"
)

// InstallClusterEssentials .
func (steps *InstallationSteps) InstallClusterEssentials(installationData *config.InstallationData, overrideData OverrideData) error {
	const stepName string = "Installing cluster-essentials"
	steps.PrintInstallationStep(stepName)
	steps.statusManager.InProgress(stepName)

	chartDir := path.Join(steps.currentPackage.GetChartsDirPath(), consts.ClusterEssentialsComponent)
	clusterEssentialsOverrides, err := steps.getClusterEssentialsOverrides(installationData, chartDir)

	if steps.errorHandlers.CheckError("Install Overrides Error: ", err) {
		steps.statusManager.Error(stepName)
		return err
	}

	allOverrides := overrides.Map{}
	overrides.MergeMaps(allOverrides, overrideData.Common())
	overrides.MergeMaps(allOverrides, clusterEssentialsOverrides) //TODO: Remove after migration to generic overrides is completed
	overrides.MergeMaps(allOverrides, overrideData.ForComponent(consts.ClusterEssentialsComponent))
	overridesStr, err := overrides.ToYaml(allOverrides)

	if steps.errorHandlers.CheckError("Install Overrides Error: ", err) {
		steps.statusManager.Error(stepName)
		return err
	}

	installResp, installErr := steps.helmClient.InstallRelease(
		chartDir,
		"kyma-system",
		consts.ClusterEssentialsComponent,
		overridesStr)

	if steps.errorHandlers.CheckError("Install Error: ", installErr) {
		steps.statusManager.Error(stepName)
		return installErr
	}

	steps.helmClient.PrintRelease(installResp.Release)
	log.Println(stepName + "...DONE")

	return nil
}

// UpdateClusterEssentials .
func (steps *InstallationSteps) UpdateClusterEssentials(installationData *config.InstallationData, overrideData OverrideData) error {
	const stepName string = "Updating cluster-essentials"
	steps.PrintInstallationStep(stepName)
	steps.statusManager.InProgress(stepName)

	chartDir := path.Join(steps.currentPackage.GetChartsDirPath(), consts.ClusterEssentialsComponent)
	clusterEssentialsOverrides, err := steps.getClusterEssentialsOverrides(installationData, chartDir)

	if steps.errorHandlers.CheckError("Upgrade Overrides Error: ", err) {
		steps.statusManager.Error(stepName)
		return err
	}

	allOverrides := overrides.Map{}
	overrides.MergeMaps(allOverrides, overrideData.Common())
	overrides.MergeMaps(allOverrides, clusterEssentialsOverrides) //TODO: Remove after migration to generic overrides is completed
	overrides.MergeMaps(allOverrides, overrideData.ForComponent(consts.ClusterEssentialsComponent))

	overridesStr, err := overrides.ToYaml(allOverrides)

	if steps.errorHandlers.CheckError("Upgrade Overrides Error: ", err) {
		steps.statusManager.Error(stepName)
		return err
	}

	upgradeResp, upgradeErr := steps.helmClient.UpgradeRelease(
		chartDir,
		consts.ClusterEssentialsComponent,
		overridesStr)

	if steps.errorHandlers.CheckError("Upgrade Error: ", upgradeErr) {
		steps.statusManager.Error(stepName)
		return upgradeErr
	}

	steps.helmClient.PrintRelease(upgradeResp.Release)
	log.Println(stepName + "...DONE")

	return nil
}

func (steps *InstallationSteps) getClusterEssentialsOverrides(installationData *config.InstallationData, chartDir string) (overrides.Map, error) {
	allOverrides := overrides.Map{}

	globalOverrides, err := overrides.GetGlobalOverrides(installationData)
	if steps.errorHandlers.CheckError("Couldn't get global overrides: ", err) {
		return nil, err
	}

	overrides.MergeMaps(allOverrides, globalOverrides)

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
