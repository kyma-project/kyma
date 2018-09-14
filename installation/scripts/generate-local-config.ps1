$CURRENT_DIR = Split-Path $MyInvocation.MyCommand.Path

$CONFIG_TPL_PATH = "${CURRENT_DIR}\..\resources\installer-config-local.yaml.tpl"
$CONFIG_OUTPUT_PATH = (New-TemporaryFile).FullName

$VERSIONS_FILE_PATH = "${CURRENT_DIR}\..\versions.env"

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

##########

if (Test-Path env.AZURE_BROKER_SUBSCRIPTION_ID) {
    Write-Output "Generating secret for Azure Broker ..."
    $AB_SECRET_TPL_PATH = "${CURRENT_DIR}\..\resources\azure-broker-secret.yaml.tpl"
    $AB_SECRET_PATH = (New-TemporaryFile).FullName

    Copy-Item -Path $AB_SECRET_TPL_PATH -Destination $AB_SECRET_PATH

    $AB_SUBSCRIPTION_ID = [System.Convert]::ToBase64String(
        [System.Text.Encoding]::UTF8.GetBytes($env.AZURE_BROKER_SUBSCRIPTION_ID))
    $cmd = "${CURRENT_DIR}\replace-placeholder.ps1 -path ${AB_SECRET_PATH} -placeholder `"__AZURE_BROKER_SUBSCRIPTION_ID__`" -value `"${AB_SUBSCRIPTION_ID}`""
    Invoke-Expression -Command $cmd

    $AB_TENANT_ID = [System.Convert]::ToBase64String(
        [System.Text.Encoding]::UTF8.GetBytes($env.AZURE_BROKER_TENANT_ID))
    $cmd = "${CURRENT_DIR}\replace-placeholder.ps1 -path ${AB_SECRET_PATH} -placeholder `"__AZURE_BROKER_TENANT_ID__`" -value `"${AB_TENANT_ID}`""
    Invoke-Expression -Command $cmd

    $AB_CLIENT_ID = [System.Convert]::ToBase64String(
        [System.Text.Encoding]::UTF8.GetBytes($env.AZURE_BROKER_CLIENT_ID))
    $cmd = "${CURRENT_DIR}\replace-placeholder.ps1 -path ${AB_SECRET_PATH} -placeholder `"__AZURE_BROKER_CLIENT_ID__`" -value `"${AB_CLIENT_ID}`""
    Invoke-Expression -Command $cmd

    $AB_CLIENT_SECRET = [System.Convert]::ToBase64String(
        [System.Text.Encoding]::UTF8.GetBytes($env.AZURE_BROKER_CLIENT_SECRET))
    $cmd = "${CURRENT_DIR}\replace-placeholder.ps1 -path ${AB_SECRET_PATH} -placeholder `"__AZURE_BROKER_CLIENT_SECRET__`" -value `"${AB_CLIENT_SECRET}`""
    Invoke-Expression -Command $cmd

    $cmd = "kubectl apply -f ${AB_SECRET_PATH}"
    Invoke-Expression -Command $cmd
}
