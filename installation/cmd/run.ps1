param (
    [string]$CR_PATH = "",
    [switch]$SKIP_MINIKUBE_START = $false,
    [string]$VM_DRIVER = "hyperv",
    [string]$PASSWORD = ""
)

$CURRENT_DIR = Split-Path $MyInvocation.MyCommand.Path
$SCRIPTS_DIR = "${CURRENT_DIR}\..\scripts"
$DOMAIN = "kyma.local"
$CREATE_CR_EXTRA_ARGS = ""

if ($SKIP_MINIKUBE_START -eq $false) {
    Invoke-Expression -Command "${SCRIPTS_DIR}\minikube.ps1 -vm_driver ${VM_DRIVER} -domain ${DOMAIN}

    if($LastExitCode -gt 0){
        exit
    }
}

Invoke-Expression -Command "${SCRIPTS_DIR}\build-kyma-installer.ps1 -vm_driver ${VM_DRIVER}"

Invoke-Expression -Command "${SCRIPTS_DIR}\generate-local-config.ps1 -password '${PASSWORD}'"

if ([string]::IsNullOrEmpty($CR_PATH)) {
    $CR_PATH = (New-TemporaryFile).FullName

    $cmd = "${SCRIPTS_DIR}\create-cr.ps1 -output ${CR_PATH} -domain ${DOMAIN} ${CREATE_CR_EXTRA_ARGS}"
    Invoke-Expression -Command $cmd
}


$cmd = "${SCRIPTS_DIR}\installer.ps1 -local -cr_path ${CR_PATH} ${INSTALLER_EXTRA_ARGS}"
Invoke-Expression -Command $cmd
