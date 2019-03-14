param (
    [string]$CR_PATH = "",
    [switch]$SKIP_MINIKUBE_START = $false,
    [switch]$KNATIVE = $false,
    [string]$VM_DRIVER = "hyperv",
    [string]$PASSWORD = ""
)

$CURRENT_DIR = Split-Path $MyInvocation.MyCommand.Path
$SCRIPTS_DIR = "${CURRENT_DIR}\..\scripts"
$DOMAIN = "kyma.local"
$INSTALLER_EXTRA_ARGS = ""
$MINIKUBE_EXTRA_ARGS = ""
$CREATE_CR_EXTRA_ARGS = ""

if ($KNATIVE -eq $true) {
    $MINIKUBE_EXTRA_ARGS = "${MINIKUBE_EXTRA_ARGS} -memory 10240 -disk_size 30g"
    $INSTALLER_EXTRA_ARGS = "${INSTALLER_EXTRA_ARGS} -knative"
}

if ($SKIP_MINIKUBE_START -eq $false) {
    Invoke-Expression -Command "${SCRIPTS_DIR}\minikube.ps1 -vm_driver ${VM_DRIVER} -domain ${DOMAIN} ${MINIKUBE_EXTRA_ARGS}"

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
