param (
    [string]$CR_PATH = "",
    [switch]$SKIP_MINIKUBE_START = $false,
    [string]$VM_DRIVER = "hyperv"
)

$CURRENT_DIR = Split-Path $MyInvocation.MyCommand.Path
$SCRIPTS_DIR = "${CURRENT_DIR}\..\scripts"
$DOMAIN = "kyma.local"

if ($SKIP_MINIKUBE_START -eq $false) {
    Invoke-Expression -Command "${SCRIPTS_DIR}\minikube.ps1 -vm_driver ${VM_DRIVER} -domain ${DOMAIN}"

    if($LastExitCode -gt 0){
        exit
    }
}

Invoke-Expression -Command "${SCRIPTS_DIR}\build-kyma-installer.ps1 -vm_driver ${VM_DRIVER}"

Invoke-Expression -Command "${SCRIPTS_DIR}\generate-local-config.ps1"

$CR_PATH = (New-TemporaryFile).FullName

$cmd = "${SCRIPTS_DIR}\create-cr.ps1 -output ${CR_PATH} -domain ${DOMAIN}"
Invoke-Expression -Command $cmd

$cmd = "${SCRIPTS_DIR}\installer.ps1 -local -cr_path ${CR_PATH}"
Invoke-Expression -Command $cmd
