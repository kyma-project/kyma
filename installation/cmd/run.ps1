param (
    [string]$CR_PATH = "",
    [switch]$SKIP_MINIKUBE_START = $false,
    [string]$VM_DRIVER = "hyperv"
)

$CURRENT_DIR = Split-Path $MyInvocation.MyCommand.Path
$SCRIPTS_DIR = "$CURRENT_DIR\..\scripts"
$LOCAL = $true
$DOMAIN = "kyma.local"

if ($CR_PATH -ne "") {
    $LOCAL = $false
}

if ($SKIP_MINIKUBE_START -eq $false) {
    Invoke-Expression -Command "${SCRIPTS_DIR}\minikube.ps1 -vm_driver ${VM_DRIVER} -domain ${DOMAIN}"

    if($LastExitCode -gt 0){
        exit
    }
}

if ($LOCAL -eq $true) {
    $CR_PATH = (New-TemporaryFile).FullName

    $cmd = "$SCRIPTS_DIR\create-cr.ps1 -output ${CR_PATH} -domain ${DOMAIN}"
    Invoke-Expression -Command $cmd
    Get-Content -Path $CR_PATH

    $cmd = "${SCRIPTS_DIR}\installer.ps1 -local -cr_path ${CR_PATH}"
    Invoke-Expression -Command $cmd
}
else {
    Write-Output "Non-local run is not supported!"
    exit
}