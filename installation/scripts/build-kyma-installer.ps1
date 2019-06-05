param (
    [string]$INSTALLER_VERSION = "",
    [string]$VM_DRIVER
)

Write-Output @"The build-kyma-installer.ps1 script is deprecated and will be removed. Use Kyma CLI instead."@

$CURRENT_DIR = Split-Path $MyInvocation.MyCommand.Path
$ROOT_DIR = "${CURRENT_DIR}\..\.."

$INSTALLER_VERSION_ARG = ""
if($INSTALLER_VERSION -ne "") {
    $INSTALLER_VERSION_ARG = "--build-arg INSTALLER_VERSION=${INSTALLER_VERSION}"
}

$cmd = "${CURRENT_DIR}\extract-kyma-installer-image.ps1"
$IMAGE_NAME = (Invoke-Expression -Command $cmd | Out-String).Trim()

minikube.exe docker-env | Invoke-Expression

Push-Location $ROOT_DIR

$cmd = "docker build -t ${IMAGE_NAME} ${INSTALLER_VERSION_ARG} -f tools\kyma-installer\kyma.Dockerfile ."
Write-Output $cmd
Invoke-Expression -Command $cmd

Pop-Location
