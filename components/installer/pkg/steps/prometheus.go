package steps

import (
	"log"
	"path"

	"github.com/kyma-project/kyma/components/installer/pkg/config"
	"github.com/kyma-project/kyma/components/installer/pkg/consts"
	"github.com/kyma-project/kyma/components/installer/pkg/overrides"
)

//InstallPrometheus .
func (steps *InstallationSteps) InstallPrometheus(installationData *config.InstallationData, overrideData overrides.OverrideData) error {
	const stepName string = "Installing Prometheus operator"
	const namespace = "kyma-system"

	chartDir := path.Join(steps.currentPackage.GetChartsDirPath(), consts.PrometheusComponent)
	prometheusOverrides, err := steps.getPrometheusOverrides(installationData, chartDir)

	if steps.errorHandlers.CheckError("Install Overrides Error: ", err) {
		steps.statusManager.Error(stepName)
		return err
	}

	allOverrides := overrides.Map{}
	overrides.MergeMaps(allOverrides, overrideData.Common())
	overrides.MergeMaps(allOverrides, prometheusOverrides) //TODO: Remove after migration to generic overrides is completed
	overrides.MergeMaps(allOverrides, overrideData.ForComponent(consts.PrometheusComponent))
	overridesStr, err := overrides.ToYaml(allOverrides)

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
		overridesStr)

	if steps.errorHandlers.CheckError("Install Error: ", installErr) {
		steps.statusManager.Error(stepName)
		return installErr
	}

	steps.helmClient.PrintRelease(installResp.Release)
	log.Println(stepName + "...DONE")

	return nil
}

//UpdatePrometheus .
func (steps *InstallationSteps) UpdatePrometheus(installationData *config.InstallationData, overrideData overrides.OverrideData) error {
	const stepName string = "Updating Prometheus operator"
	const namespace = "kyma-system"

	chartDir := path.Join(steps.currentPackage.GetChartsDirPath(), consts.PrometheusComponent)
	prometheusOverrides, err := steps.getPrometheusOverrides(installationData, chartDir)

	if steps.errorHandlers.CheckError("Upgrade Overrides Error: ", err) {
		steps.statusManager.Error(stepName)
		return err
	}

	allOverrides := overrides.Map{}
	overrides.MergeMaps(allOverrides, overrideData.Common())
	overrides.MergeMaps(allOverrides, prometheusOverrides) //TODO: Remove after migration to generic overrides is completed
	overrides.MergeMaps(allOverrides, overrideData.ForComponent(consts.PrometheusComponent))
	overridesStr, err := overrides.ToYaml(allOverrides)

	if steps.errorHandlers.CheckError("Upgrade Overrides Error: ", err) {
		steps.statusManager.Error(stepName)
		return err
	}

	steps.PrintInstallationStep(stepName)
	steps.statusManager.InProgress(stepName)

	upgradeResp, upgradeErr := steps.helmClient.UpgradeRelease(
		chartDir,
		consts.PrometheusComponent,
		overridesStr)

	if steps.errorHandlers.CheckError("Upgrade Error: ", upgradeErr) {
		steps.statusManager.Error(stepName)
		return upgradeErr
	}

	steps.helmClient.PrintRelease(upgradeResp.Release)
	log.Println(stepName + "...DONE")

	return nil
}

func (steps *InstallationSteps) getPrometheusOverrides(installationData *config.InstallationData, chartDir string) (overrides.Map, error) {
	allOverrides := overrides.Map{}

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
