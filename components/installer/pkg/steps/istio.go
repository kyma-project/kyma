package steps

import (
	"log"
	"path"
	"strings"

	"github.com/kyma-project/kyma/components/installer/pkg/config"
	"github.com/kyma-project/kyma/components/installer/pkg/consts"
	"github.com/kyma-project/kyma/components/installer/pkg/overrides"
)

// InstallIstio .
func (steps *InstallationSteps) InstallIstio(installationData *config.InstallationData) error {
	const stepName string = "Installing istio"
	steps.PrintInstallationStep(stepName)
	steps.statusManager.InProgress(stepName)

	chartDir := path.Join(steps.chartDir, consts.IstioComponent, "istio")
	overrides := steps.getIstioOverrides(installationData, chartDir)

	//helm install
	installResp, installErr := steps.helmClient.InstallReleaseWithoutWait(
		chartDir,
		"istio-system",
		consts.IstioComponent,
		overrides)

	if steps.errorHandlers.CheckError("Install Error: ", installErr) {
		steps.statusManager.Error(stepName)
		return installErr
	}

	steps.helmClient.PrintRelease(installResp.Release)
	log.Println(stepName + "...DONE")

	return nil
}

// UpdateIstio .
func (steps *InstallationSteps) UpdateIstio(installationData *config.InstallationData) error {
	const stepName string = "Updating istio"
	steps.PrintInstallationStep(stepName)
	steps.statusManager.InProgress(stepName)

	chartDir := path.Join(steps.chartDir, consts.IstioComponent, "istio")
	overrides := steps.getIstioOverrides(installationData, chartDir)

	upgradeResp, upgradeErr := steps.helmClient.UpgradeRelease(
		chartDir,
		consts.IstioComponent,
		overrides)

	if steps.errorHandlers.CheckError("Upgrade Error: ", upgradeErr) {
		steps.statusManager.Error("Updating istio")
		return upgradeErr
	}

	steps.helmClient.PrintRelease(upgradeResp.Release)
	log.Println(stepName + "...DONE")

	return nil
}

func (steps *InstallationSteps) getIstioOverrides(installationData *config.InstallationData, chartDir string) string {
	var allOverrides []string

	globalOverrides, err := overrides.GetGlobalOverrides(installationData)
	steps.errorHandlers.LogError("Couldn't get global overrides: ", err)
	allOverrides = append(allOverrides, globalOverrides)

	istioOverrides, err := overrides.GetIstioOverrides(installationData)
	steps.errorHandlers.LogError("Couldn't get Istio overrides: ", err)
	allOverrides = append(allOverrides, istioOverrides)

	fileOverrides := steps.getStaticFileOverrides(installationData, chartDir)
	if fileOverrides.HasOverrides() == true {
		fileOverridesStr, err := fileOverrides.GetOverrides()
		steps.errorHandlers.LogError("Couldn't get additional overrides: ", err)
		allOverrides = append(allOverrides, *fileOverridesStr)
	}

	return strings.Join(allOverrides, "\n")
}
