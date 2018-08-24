package steps

import (
	"log"
	"os/exec"
	"path"

	"github.com/kyma-project/kyma/components/installer/pkg/config"
	"github.com/kyma-project/kyma/components/installer/pkg/consts"
	"github.com/kyma-project/kyma/components/installer/pkg/overrides"
)

const kymaPath = "/kyma"

//InstallCore .
func (steps InstallationSteps) InstallCore(installationData *config.InstallationData) error {
	const stepName string = "Installing core"
	steps.PrintInstallationStep(stepName)
	steps.statusManager.InProgress(stepName)

	chartDir := path.Join(steps.currentPackage.GetChartsDirPath(), consts.CoreComponent)
	coreOverrides, err := steps.getCoreOverrides(installationData, chartDir)

	if steps.errorHandlers.CheckError("Install Overrides Error: ", err) {
		steps.statusManager.Error(stepName)
		logCore(steps)
		return err
	}

	installResp, installErr := steps.helmClient.InstallRelease(
		chartDir,
		"kyma-system",
		consts.CoreComponent,
		coreOverrides)

	if steps.errorHandlers.CheckError("Install Error: ", installErr) {
		steps.statusManager.Error(stepName)
		logCore(steps)
		return installErr
	}

	steps.helmClient.PrintRelease(installResp.Release)
	log.Println(stepName + "...DONE")

	return nil
}

//UpgradeCore .
func (steps InstallationSteps) UpgradeCore(installationData *config.InstallationData) error {
	const stepName string = "Upgrading core"
	steps.PrintInstallationStep(stepName)
	steps.statusManager.InProgress(stepName)

	chartDir := path.Join(steps.currentPackage.GetChartsDirPath(), consts.CoreComponent)
	coreOverrides, err := steps.getCoreOverrides(installationData, chartDir)

	if steps.errorHandlers.CheckError("Upgrade Overrides Error: ", err) {
		steps.statusManager.Error(stepName)
		logCore(steps)
		return err
	}

	upgradeResp, upgradeErr := steps.helmClient.UpgradeRelease(
		chartDir,
		consts.CoreComponent,
		coreOverrides)

	if steps.errorHandlers.CheckError("Upgrade Error: ", upgradeErr) {
		steps.statusManager.Error(stepName)
		logCore(steps)
		return upgradeErr
	}

	steps.helmClient.PrintRelease(upgradeResp.Release)
	log.Println(stepName + "...DONE")

	return nil
}

//Legacy stuff from old build scripts.
//TODO: Provide such logs for every step
func logCore(steps InstallationSteps) {
	//Try to display debug data in case of failure
	status, statusErr := steps.helmClient.ReleaseStatus("core")
	if statusErr != nil {
		log.Println("Cannot get release status: ", statusErr)
	} else {
		log.Println("core status: \n" + status)
	}

	logFailedResources("kyma-system")
}

func logFailedResources(ns string) {
	path := path.Join(kymaPath, "installation/scripts/utils.sh")
	log.Println("\nLooking for failed resources in namespace: " + ns)
	cmd := exec.Command("/bin/bash", "-c", "source "+path+"; showFailedResources "+ns)
	msg, scriptErr := cmd.Output()
	if scriptErr != nil {
		log.Printf("An error occurred while running script: %s (%s)\n", string(msg[:]), scriptErr)
		return
	}
	log.Println(string(msg[:]))
}

func (steps *InstallationSteps) getCoreOverrides(installationData *config.InstallationData, chartDir string) (string, error) {
	allOverrides := overrides.Map{}

	globalOverrides, err := overrides.GetGlobalOverrides(installationData)
	steps.errorHandlers.LogError("Couldn't get global overrides: ", err)
	overrides.MergeMaps(allOverrides, globalOverrides)

	azureBrokerOverrides, err := overrides.EnableAzureBroker(installationData)
	steps.errorHandlers.LogError("Enable azure-broker Error: ", err)
	overrides.MergeMaps(allOverrides, azureBrokerOverrides)

	coreOverrides, err := overrides.GetCoreOverrides(installationData)
	steps.errorHandlers.LogError("Couldn't get Kyma core overrides: ", err)

	overrides.MergeMaps(allOverrides, coreOverrides)

	staticOverrides := steps.getStaticFileOverrides(installationData, chartDir)
	if staticOverrides.HasOverrides() == true {
		fileOverrides, err := staticOverrides.GetOverrides()
		steps.errorHandlers.LogError("Couldn't get additional overrides: ", err)
		overrides.MergeMaps(allOverrides, fileOverrides)
	}

	return overrides.ToYaml(allOverrides)
}
