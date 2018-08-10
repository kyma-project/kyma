package steps

import (
	"log"
	"path"

	"github.com/kyma-project/kyma/components/installer/pkg/config"
	"github.com/kyma-project/kyma/components/installer/pkg/consts"
	"github.com/kyma-project/kyma/components/installer/pkg/overrides"
)

//InstallPrometheus .
func (steps *InstallationSteps) InstallPrometheus(installationData *config.InstallationData) error {
	const stepName string = "Installing Prometheus operator"
	const namespace = "kyma-system"

	chartDir := path.Join(steps.chartDir, consts.PrometheusComponent)
	overrides, err := steps.getPrometheusOverrides(installationData, chartDir)

	if steps.errorHandlers.CheckError("Install Overrides Error: ", err) {
		steps.statusManager.Error(stepName)
		return err
	}

	steps.PrintInstallationStep(stepName)
	steps.statusManager.InProgress(stepName)

	installResp, installErr := steps.helmClient.InstallRelease(
		chartDir,
		namespace,
		consts.PrometheusComponent,
		overrides)

	if steps.errorHandlers.CheckError("Install Error: ", installErr) {
		steps.statusManager.Error(stepName)
		return installErr
	}

	steps.helmClient.PrintRelease(installResp.Release)
	log.Println(stepName + "...DONE")

	return nil
}

//UpdatePrometheus .
func (steps *InstallationSteps) UpdatePrometheus(installationData *config.InstallationData) error {
	const stepName string = "Updating Prometheus operator"
	const namespace = "kyma-system"

	chartDir := path.Join(steps.chartDir, consts.PrometheusComponent)
	overrides, err := steps.getPrometheusOverrides(installationData, chartDir)

	if steps.errorHandlers.CheckError("Upgrade Overrides Error: ", err) {
		steps.statusManager.Error(stepName)
		return err
	}

	steps.PrintInstallationStep(stepName)
	steps.statusManager.InProgress(stepName)

	upgradeResp, upgradeErr := steps.helmClient.UpgradeRelease(
		chartDir,
		consts.PrometheusComponent,
		overrides)

	if steps.errorHandlers.CheckError("Upgrade Error: ", upgradeErr) {
		steps.statusManager.Error(stepName)
		return upgradeErr
	}

	steps.helmClient.PrintRelease(upgradeResp.Release)
	log.Println(stepName + "...DONE")

	return nil
}
func (steps *InstallationSteps) getPrometheusOverrides(installationData *config.InstallationData, chartDir string) (string, error) {
	//TODO: this does not get globalOverrides... Is that a problem if global will carry all external ones (overrides.yaml + from configMaps/secrets?)
	allOverrides := overrides.OverridesMap{}

	staticOverrides := steps.getStaticFileOverrides(installationData, chartDir)
	if staticOverrides.HasOverrides() == true {
		fileOverrides, err := staticOverrides.GetOverrides()
		steps.errorHandlers.LogError("Couldn't get additional overrides: ", err)
		overrides.MergeMaps(allOverrides, fileOverrides)
	}

	return overrides.ToYaml(allOverrides)
}
