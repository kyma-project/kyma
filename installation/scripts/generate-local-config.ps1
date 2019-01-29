$CURRENT_DIR = Split-Path $MyInvocation.MyCommand.Path

$CONFIG_TPL_PATH = "${CURRENT_DIR}\..\resources\installer-config-local.yaml.tpl"
$CONFIG_OUTPUT_PATH = (New-TemporaryFile).FullName

$VERSIONS_FILE_PATH = "${CURRENT_DIR}\..\versions-overrides.env"

Copy-Item -Path $CONFIG_TPL_PATH -Destination $CONFIG_OUTPUT_PATH

##########

Write-Output "Applying configuration"

$cmd = "kubectl.exe create ns kyma-installer"
Invoke-Expression -Command $cmd

$cmd = "kubectl apply -f ${CONFIG_OUTPUT_PATH}"
Invoke-Expression -Command $cmd

##########

Write-Output "Configuring sub-components ..."

$cmd = "${CURRENT_DIR}\configure-components.ps1"
Invoke-Expression -Command $cmd

##########

Write-Output "Configuring versions ..."

if([System.IO.File]::Exists($VERSIONS_FILE_PATH)){
    $cmd = "kubectl create configmap versions --from-env-file=${VERSIONS_FILE_PATH} -n `"kyma-installer`""
    Invoke-Expression -Command $cmd

    $cmd = "kubectl label configmap/versions installer=overrides -n `"kyma-installer`""
    Invoke-Expression -Command $cmd
}
