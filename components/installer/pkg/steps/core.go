package steps

import (
	"log"
	"os/exec"
	"path"
	"strings"

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

	chartDir := path.Join(steps.chartDir, consts.CoreComponent)
	coreOverrides := steps.getCoreOverrides(installationData, chartDir)

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

	chartDir := path.Join(steps.chartDir, consts.CoreComponent)
	coreOverrides := steps.getCoreOverrides(installationData, chartDir)

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

func (steps *InstallationSteps) getCoreOverrides(installationData *config.InstallationData, chartDir string) string {
	var allOverrides []string

	globalOverrides, err := overrides.GetGlobalOverrides(installationData)
	steps.errorHandlers.LogError("Couldn't get global overrides: ", err)
	allOverrides = append(allOverrides, globalOverrides)

	azureBrokerOverrides, err := overrides.EnableAzureBroker(installationData)
	steps.errorHandlers.LogError("Enable azure-broker Error: ", err)
	allOverrides = append(allOverrides, azureBrokerOverrides)

	coreOverrides, err := overrides.GetCoreOverrides(installationData)
	steps.errorHandlers.LogError("Couldn't get Kyma core overrides: ", err)

	allOverrides = append(allOverrides, coreOverrides)

	fileOverrides := steps.getStaticFileOverrides(installationData, chartDir)
	if fileOverrides.HasOverrides() == true {
		fileOverridesStr, err := fileOverrides.GetOverrides()
		steps.errorHandlers.LogError("Couldn't get additional overrides: ", err)
		allOverrides = append(allOverrides, *fileOverridesStr)
	}

	return strings.Join(allOverrides, "\n")
}
