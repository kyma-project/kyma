package steps

import (
	"log"
	"path"

	"github.com/kyma-project/kyma/components/installer/pkg/config"
	"github.com/kyma-project/kyma/components/installer/pkg/consts"
	"github.com/kyma-project/kyma/components/installer/pkg/overrides"
)

// InstallIstio .
func (steps *InstallationSteps) InstallIstio(installationData *config.InstallationData, overrideData overrides.OverrideData) error {
	const stepName string = "Installing istio"
	steps.PrintInstallationStep(stepName)
	steps.statusManager.InProgress(stepName)

	chartDir := path.Join(steps.currentPackage.GetChartsDirPath(), consts.IstioComponent, "istio")
	istioOverrides, err := steps.getIstioOverrides(installationData, chartDir, overrideData)

	if steps.errorHandlers.CheckError("Install Overrides Error: ", err) {
		steps.statusManager.Error(stepName)
		return err
	}

	//helm install
	installResp, installErr := steps.helmClient.InstallReleaseWithoutWait(
		chartDir,
		"istio-system",
		consts.IstioComponent,
		istioOverrides)

	if steps.errorHandlers.CheckError("Install Error: ", installErr) {
		steps.statusManager.Error(stepName)
		return installErr
	}

	steps.helmClient.PrintRelease(installResp.Release)
	log.Println(stepName + "...DONE")

	return nil
}

// UpdateIstio .
func (steps *InstallationSteps) UpdateIstio(installationData *config.InstallationData, overrideData overrides.OverrideData) error {
	const stepName string = "Updating istio"
	steps.PrintInstallationStep(stepName)
	steps.statusManager.InProgress(stepName)

	chartDir := path.Join(steps.currentPackage.GetChartsDirPath(), consts.IstioComponent, "istio")
	istioOverrides, err := steps.getIstioOverrides(installationData, chartDir, overrideData)

	if steps.errorHandlers.CheckError("Upgrade Overrides Error: ", err) {
		steps.statusManager.Error("Updating istio")
		return err
	}

	upgradeResp, upgradeErr := steps.helmClient.UpgradeRelease(
		chartDir,
		consts.IstioComponent,
		istioOverrides)

	if steps.errorHandlers.CheckError("Upgrade Error: ", upgradeErr) {
		steps.statusManager.Error("Updating istio")
		return upgradeErr
	}

	steps.helmClient.PrintRelease(upgradeResp.Release)
	log.Println(stepName + "...DONE")

	return nil
}

func (steps *InstallationSteps) getIstioOverrides(installationData *config.InstallationData, chartDir string, overrideData overrides.OverrideData) (string, error) {
	allOverrides := overrides.Map{}
	overrides.MergeMaps(allOverrides, overrideData.Common())
	overrides.MergeMaps(allOverrides, overrideData.ForComponent(consts.IstioComponent))

	globalOverrides, err := overrides.GetGlobalOverrides(installationData, allOverrides)
	if steps.errorHandlers.CheckError("Couldn't get global overrides: ", err) {
		return "", err
	}
	overrides.MergeMaps(allOverrides, globalOverrides)

	istioOverrides, err := overrides.GetIstioOverrides(installationData, allOverrides)
	if steps.errorHandlers.CheckError("Couldn't get Istio overrides: ", err) {
		return "", err
	}
	overrides.MergeMaps(allOverrides, istioOverrides)

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
