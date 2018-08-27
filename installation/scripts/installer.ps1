param (
    [string]$CR_PATH = $null,
    [switch]$LOCAL = $false
)

$CURRENT_DIR = Split-Path $MyInvocation.MyCommand.Path

$cmd = "kubectl.exe create ns kyma-installer"
Invoke-Expression -Command $cmd

$cmd = "kubectl apply -f ${CURRENT_DIR}\..\resources\default-sa-rbac-role.yaml"
Invoke-Expression -Command $cmd

$cmd = "kubectl apply -f ${CURRENT_DIR}\..\resources\limit-range-installer.yaml"
Invoke-Expression -Command $cmd

$cmd = "kubectl apply -f ${CURRENT_DIR}\..\resources\resource-quotas-installer.yaml"
Invoke-Expression -Command $cmd

$cmd = "${CURRENT_DIR}\install-tiller.ps1"
Invoke-Expression -Command $cmd

$cmd = "kubectl.exe apply -f ${CURRENT_DIR}\..\resources\installer.yaml -n kyma-installer"
Invoke-Expression -Command $cmd

$cmd = "${CURRENT_DIR}\is-ready.ps1 -ns kube-system -label k8s-app -value kube-dns"
Invoke-Expression -Command $cmd

if ($LOCAL) {
    $cmd = "${CURRENT_DIR}\generate-local-config.ps1"
    Invoke-Expression -Command $cmd

    $cmd = "${CURRENT_DIR}\copy-resource.ps1"
    Invoke-Expression -Command $cmd
}
else {
    Write-Output "Non-local run is not supported!"
    exit
}

if ($CR_PATH -ne "") {
    $cmd = "kubectl.exe apply -f ${CR_PATH}"
    Invoke-Expression -Command $cmd
}