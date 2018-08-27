package steps

import (
	"log"
	"path"

	"github.com/kyma-project/kyma/components/installer/pkg/config"
	"github.com/kyma-project/kyma/components/installer/pkg/consts"
	"github.com/kyma-project/kyma/components/installer/pkg/overrides"
)

// InstallIstio .
func (steps *InstallationSteps) InstallIstio(installationData *config.InstallationData, overrideData *overrides.Overrides) error {
	const stepName string = "Installing istio"
	steps.PrintInstallationStep(stepName)
	steps.statusManager.InProgress(stepName)

	chartDir := path.Join(steps.currentPackage.GetChartsDirPath(), consts.IstioComponent, "istio")
	istioOverrides, err := steps.getIstioOverrides(installationData, chartDir)

	if steps.errorHandlers.CheckError("Install Overrides Error: ", err) {
		steps.statusManager.Error(stepName)
		return err
	}

	allOverrides := overrides.Map{}
	overrides.MergeMaps(allOverrides, overrideData.Common())
	overrides.MergeMaps(allOverrides, istioOverrides) //TODO: Remove after migration to generic overrides is completed
	overrides.MergeMaps(allOverrides, overrideData.ForComponent(consts.IstioComponent))
	overridesStr, err := overrides.ToYaml(allOverrides)

	if steps.errorHandlers.CheckError("Install Overrides Error: ", err) {
		steps.statusManager.Error(stepName)
		return err
	}

	//helm install
	installResp, installErr := steps.helmClient.InstallReleaseWithoutWait(
		chartDir,
		"istio-system",
		consts.IstioComponent,
		overridesStr)

	if steps.errorHandlers.CheckError("Install Error: ", installErr) {
		steps.statusManager.Error(stepName)
		return installErr
	}

	steps.helmClient.PrintRelease(installResp.Release)
	log.Println(stepName + "...DONE")

	return nil
}

// UpdateIstio .
func (steps *InstallationSteps) UpdateIstio(installationData *config.InstallationData, overrideData *overrides.Overrides) error {
	const stepName string = "Updating istio"
	steps.PrintInstallationStep(stepName)
	steps.statusManager.InProgress(stepName)

	chartDir := path.Join(steps.currentPackage.GetChartsDirPath(), consts.IstioComponent, "istio")
	istioOverrides, err := steps.getIstioOverrides(installationData, chartDir)

	if steps.errorHandlers.CheckError("Upgrade Overrides Error: ", err) {
		steps.statusManager.Error("Updating istio")
		return err
	}

	allOverrides := overrides.Map{}
	overrides.MergeMaps(allOverrides, overrideData.Common())
	overrides.MergeMaps(allOverrides, istioOverrides) //TODO: Remove after migration to generic overrides is completed
	overrides.MergeMaps(allOverrides, overrideData.ForComponent(consts.IstioComponent))
	overridesStr, err := overrides.ToYaml(allOverrides)

	if steps.errorHandlers.CheckError("Upgrade Overrides Error: ", err) {
		steps.statusManager.Error(stepName)
		return err
	}

	upgradeResp, upgradeErr := steps.helmClient.UpgradeRelease(
		chartDir,
		consts.IstioComponent,
		overridesStr)

	if steps.errorHandlers.CheckError("Upgrade Error: ", upgradeErr) {
		steps.statusManager.Error("Updating istio")
		return upgradeErr
	}

	steps.helmClient.PrintRelease(upgradeResp.Release)
	log.Println(stepName + "...DONE")

	return nil
}

func (steps *InstallationSteps) getIstioOverrides(installationData *config.InstallationData, chartDir string) (overrides.Map, error) {
	allOverrides := overrides.Map{}

	globalOverrides, err := overrides.GetGlobalOverrides(installationData)

	if steps.errorHandlers.CheckError("Couldn't get global overrides: ", err) {
		return nil, err
	}

	overrides.MergeMaps(allOverrides, globalOverrides)

	istioOverrides, err := overrides.GetIstioOverrides(installationData)
	if steps.errorHandlers.CheckError("Couldn't get Istio overrides: ", err) {
		return nil, err
	}
	overrides.MergeMaps(allOverrides, istioOverrides)

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
