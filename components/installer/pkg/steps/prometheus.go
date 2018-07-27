package steps

import (
	"log"
	"path"
	"strings"

	"github.com/kyma-project/kyma/components/installer/pkg/config"
	"github.com/kyma-project/kyma/components/installer/pkg/consts"
)

//InstallPrometheus .
func (steps *InstallationSteps) InstallPrometheus(installationData *config.InstallationData) error {
	const stepName string = "Installing Prometheus operator"
	const namespace = "kyma-system"

	chartDir := path.Join(steps.chartDir, consts.PrometheusComponent)
	overrides := steps.getPrometheusOverrides(installationData, chartDir)

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
	overrides := steps.getPrometheusOverrides(installationData, chartDir)

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
func (steps *InstallationSteps) getPrometheusOverrides(installationData *config.InstallationData, chartDir string) string {
	var allOverrides []string

	fileOverrides := steps.getStaticFileOverrides(installationData, chartDir)
	if fileOverrides.HasOverrides() == true {
		fileOverridesStr, err := fileOverrides.GetOverrides()
		steps.errorHandlers.LogError("Couldn't get additional overrides: ", err)
		allOverrides = append(allOverrides, *fileOverridesStr)
	}

	return strings.Join(allOverrides, "\n")
}
