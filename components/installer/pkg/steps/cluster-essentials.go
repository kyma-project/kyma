package steps

import (
	"log"
	"path"

	"github.com/kyma-project/kyma/components/installer/pkg/config"
	"github.com/kyma-project/kyma/components/installer/pkg/consts"
	"github.com/kyma-project/kyma/components/installer/pkg/overrides"
)

// InstallClusterEssentials .
func (steps *InstallationSteps) InstallClusterEssentials(installationData *config.InstallationData) error {
	const stepName string = "Installing cluster-essentials"
	steps.PrintInstallationStep(stepName)
	steps.statusManager.InProgress(stepName)

	chartDir := path.Join(steps.chartDir, consts.ClusterEssentialsComponent)
	clusterEssentialsOverrides, err := steps.getClusterEssentialsOverrides(installationData, chartDir)

	if steps.errorHandlers.CheckError("Install Overrides Error: ", err) {
		steps.statusManager.Error(stepName)
		return err
	}

	installResp, installErr := steps.helmClient.InstallRelease(
		chartDir,
		"kyma-system",
		consts.ClusterEssentialsComponent,
		clusterEssentialsOverrides)

	if steps.errorHandlers.CheckError("Install Error: ", installErr) {
		steps.statusManager.Error(stepName)
		return installErr
	}

	steps.helmClient.PrintRelease(installResp.Release)
	log.Println(stepName + "...DONE")

	return nil
}

// UpdateClusterEssentials .
func (steps *InstallationSteps) UpdateClusterEssentials(installationData *config.InstallationData) error {
	const stepName string = "Updating cluster-essentials"
	steps.PrintInstallationStep(stepName)
	steps.statusManager.InProgress(stepName)

	chartDir := path.Join(steps.chartDir, consts.ClusterEssentialsComponent)
	clusterEssentailsOverrides, err := steps.getClusterEssentialsOverrides(installationData, chartDir)

	if steps.errorHandlers.CheckError("Upgrade Overrides Error: ", err) {
		steps.statusManager.Error(stepName)
		return err
	}

	upgradeResp, upgradeErr := steps.helmClient.UpgradeRelease(
		chartDir,
		consts.ClusterEssentialsComponent,
		clusterEssentailsOverrides)

	if steps.errorHandlers.CheckError("Upgrade Error: ", upgradeErr) {
		steps.statusManager.Error(stepName)
		return upgradeErr
	}

	steps.helmClient.PrintRelease(upgradeResp.Release)
	log.Println(stepName + "...DONE")

	return nil
}

func (steps *InstallationSteps) getClusterEssentialsOverrides(installationData *config.InstallationData, chartDir string) (string, error) {
	allOverrides := overrides.OverridesMap{}

	globalOverrides, err := overrides.GetGlobalOverrides(installationData)
	steps.errorHandlers.LogError("Couldn't get global overrides: ", err)
	overrides.MergeMaps(allOverrides, globalOverrides)

	staticOverrides := steps.getStaticFileOverrides(installationData, chartDir)
	if staticOverrides.HasOverrides() == true {
		fileOverrides, err := staticOverrides.GetOverrides()
		steps.errorHandlers.LogError("Couldn't get additional overrides: ", err)
		overrides.MergeMaps(allOverrides, fileOverrides)
	}

	return overrides.ToYaml(allOverrides)
}
