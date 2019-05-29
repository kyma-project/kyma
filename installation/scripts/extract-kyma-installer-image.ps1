$CURRENT_DIR = Split-Path $MyInvocation.MyCommand.Path
$ROOT_DIR = "${CURRENT_DIR}\..\.."
$INSTALLER_YAML_PATH = "${ROOT_DIR}\installation\resources\installer-local.yaml"

if([System.IO.File]::Exists($INSTALLER_YAML_PATH)) {
    $VERSION = (Get-Content -Path $INSTALLER_YAML_PATH | Select-String -Pattern "kyma-installer" -Encoding utf8 | Select-String -Pattern "image:" -Encoding utf8 | Out-String ).Split(":", 2)[1].Trim()
    Write-Output $VERSION
} else {
    Write-Output "${INSTALLER_YAML_PATH} not found"
    exit 1
}
