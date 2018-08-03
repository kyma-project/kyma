$CURRENT_DIR = Split-Path $MyInvocation.MyCommand.Path

$CONFIG_TPL_PATH = "${CURRENT_DIR}\..\resources\installer-config.yaml.tpl"
$CONFIG_OUTPUT_PATH = (New-TemporaryFile).FullName

Copy-Item -Path $CONFIG_TPL_PATH -Destination $CONFIG_OUTPUT_PATH

##########

Write-Output "Generating secret for cluster certificate ..."

$TLS_FILE="${CURRENT_DIR}\..\resources\local-tls-certs.yaml"
$TLS_CRT = Get-Content -Path "${TLS_FILE}" | Select-String -Pattern 'tls.crt: .*'
$TLS_CRT = $TLS_CRT.ToString().Replace("tls.crt:", "").Trim()
$TLS_CRT = [System.Convert]::ToBase64String(
    [System.Text.Encoding]::UTF8.GetBytes($TLS_CRT))

$TLS_KEY = Get-Content -Path "${TLS_FILE}" | Select-String -Pattern 'tls.key: .*'
$TLS_KEY = $TLS_KEY.ToString().Replace("tls.key:", "").Trim()
$TLS_KEY = [System.Convert]::ToBase64String(
    [System.Text.Encoding]::UTF8.GetBytes($TLS_KEY))

$cmd = "${SCRIPTS_DIR}\replace-placeholder.ps1 -path ${CONFIG_OUTPUT_PATH} -placeholder `"__TLS_CERT__`" -value `"${TLS_CRT}`""
Invoke-Expression -Command $cmd

$cmd = "${SCRIPTS_DIR}\replace-placeholder.ps1 -path ${CONFIG_OUTPUT_PATH} -placeholder `"__TLS_KEY__`" -value `"${TLS_KEY}`""
Invoke-Expression -Command $cmd

##########

Write-Output "Generating secret for Remote Environemnts"

$cmd = "${SCRIPTS_DIR}\replace-placeholder.ps1 -path ${CONFIG_OUTPUT_PATH} -placeholder `"__REMOTE_ENV_CA__`" -value `"`""
Invoke-Expression -Command $cmd

$cmd = "${SCRIPTS_DIR}\replace-placeholder.ps1 -path ${CONFIG_OUTPUT_PATH} -placeholder `"__REMOTE_ENV_CA_KEY__`" -value `"`""
Invoke-Expression -Command $cmd

##########

Write-Output "Generating config map for installation ..."

$cmd = "minikube.exe ip"
$MINIKUBE_IP = (Invoke-Expression -Command $cmd | Out-String).ToString().Trim()

$MINIKUBE_CA_CRT = Get-Content -Path "${HOME}\.minikube\ca.crt"
$MINIKUBE_CA = [System.Convert]::ToBase64String(
    [System.Text.Encoding]::UTF8.GetBytes($MINIKUBE_CA_CRT))

$cmd = "${SCRIPTS_DIR}\replace-placeholder.ps1 -path ${CONFIG_OUTPUT_PATH} -placeholder `"__DOMAIN__`" -value `"kyma.local`""
Invoke-Expression -Command $cmd

$cmd = "${SCRIPTS_DIR}\replace-placeholder.ps1 -path ${CONFIG_OUTPUT_PATH} -placeholder `"__EXTERNAL_IP_ADDRESS__`" -value `"`""
Invoke-Expression -Command $cmd

$cmd = "${SCRIPTS_DIR}\replace-placeholder.ps1 -path ${CONFIG_OUTPUT_PATH} -placeholder `"__REMOTE_ENV_IP__`" -value `"`""
Invoke-Expression -Command $cmd

$cmd = "${SCRIPTS_DIR}\replace-placeholder.ps1 -path ${CONFIG_OUTPUT_PATH} -placeholder `"__K8S_APISERVER_URL__`" -value `"${MINIKUBE_IP}`""
Invoke-Expression -Command $cmd

$cmd = "${SCRIPTS_DIR}\replace-placeholder.ps1 -path ${CONFIG_OUTPUT_PATH} -placeholder `"__K8S_APISERVER_CA__`" -value `"${MINIKUBE_CA}`""
Invoke-Expression -Command $cmd

$cmd = "${SCRIPTS_DIR}\replace-placeholder.ps1 -path ${CONFIG_OUTPUT_PATH} -placeholder `"__ADMIN_GROUP__`" -value `"`""
Invoke-Expression -Command $cmd

$cmd = "${SCRIPTS_DIR}\replace-placeholder.ps1 -path ${CONFIG_OUTPUT_PATH} -placeholder `"__ENABLE_ETCD_BACKUP_OPERATOR__`" -value `"false`""
Invoke-Expression -Command $cmd

$cmd = "${SCRIPTS_DIR}\replace-placeholder.ps1 -path ${CONFIG_OUTPUT_PATH} -placeholder `"__ETCD_BACKUP_ABS_CONTAINER_NAME__`" -value `"`""
Invoke-Expression -Command $cmd

##########

Write-Output "Applying configuration"

$cmd = "kubectl.exe create ns kyma-installer"
Invoke-Expression -Command $cmd

$cmd = "kubectl apply -f ${CONFIG_OUTPUT_PATH}"
Invoke-Expression -Command $cmd

##########

Write-Output "Generating secret for UI Test ..."

$UI_TEST_USER = [System.Convert]::ToBase64String(
    [System.Text.Encoding]::UTF8.GetBytes("admin@kyma.cx"))
$UI_TEST_PASSWORD = [System.Convert]::ToBase64String(
    [System.Text.Encoding]::UTF8.GetBytes("nimda123"))

$UI_TEST_SECRET_TPL_PATH = "${CURRENT_DIR}\..\resources\ui-test-secret.yaml.tpl"
$UI_TEST_SECRET_PATH = (New-TemporaryFile).FullName

Copy-Item -Path $UI_TEST_SECRET_TPL_PATH -Destination $UI_TEST_SECRET_PATH

$cmd = "${SCRIPTS_DIR}\replace-placeholder.ps1 -path ${UI_TEST_SECRET_PATH} -placeholder `"__UI_TEST_USER__`" -value `"${UI_TEST_USER}`""
Invoke-Expression -Command $cmd

$cmd = "${SCRIPTS_DIR}\replace-placeholder.ps1 -path ${UI_TEST_SECRET_PATH} -placeholder `"__UI_TEST_PASSWORD__`" -value `"${UI_TEST_PASSWORD}`""
Invoke-Expression -Command $cmd

$cmd = "kubectl apply -f ${UI_TEST_SECRET_PATH}"
Invoke-Expression -Command $cmd

##########

if (Test-Path env.AZURE_BROKER_SUBSCRIPTION_ID) {
    Write-Output "Generating secret for Azure Broker ..."
    $AB_SECRET_TPL_PATH = "${CURRENT_DIR}\..\resources\azure-broker-secret.yaml.tpl"
    $AB_SECRET_PATH = (New-TemporaryFile).FullName

    Copy-Item -Path $AB_SECRET_TPL_PATH -Destination $AB_SECRET_PATH

    $AB_SUBSCRIPTION_ID = [System.Convert]::ToBase64String(
        [System.Text.Encoding]::UTF8.GetBytes($env.AZURE_BROKER_SUBSCRIPTION_ID))
    $cmd = "${SCRIPTS_DIR}\replace-placeholder.ps1 -path ${AB_SECRET_PATH} -placeholder `"__AZURE_BROKER_SUBSCRIPTION_ID__`" -value `"${AB_SUBSCRIPTION_ID}`""
    Invoke-Expression -Command $cmd

    $AB_TENANT_ID = [System.Convert]::ToBase64String(
        [System.Text.Encoding]::UTF8.GetBytes($env.AZURE_BROKER_TENANT_ID))
    $cmd = "${SCRIPTS_DIR}\replace-placeholder.ps1 -path ${AB_SECRET_PATH} -placeholder `"__AZURE_BROKER_TENANT_ID__`" -value `"${AB_TENANT_ID}`""
    Invoke-Expression -Command $cmd

    $AB_CLIENT_ID = [System.Convert]::ToBase64String(
        [System.Text.Encoding]::UTF8.GetBytes($env.AZURE_BROKER_CLIENT_ID))
    $cmd = "${SCRIPTS_DIR}\replace-placeholder.ps1 -path ${AB_SECRET_PATH} -placeholder `"__AZURE_BROKER_CLIENT_ID__`" -value `"${AB_CLIENT_ID}`""
    Invoke-Expression -Command $cmd

    $AB_CLIENT_SECRET = [System.Convert]::ToBase64String(
        [System.Text.Encoding]::UTF8.GetBytes($env.AZURE_BROKER_CLIENT_SECRET))
    $cmd = "${SCRIPTS_DIR}\replace-placeholder.ps1 -path ${AB_SECRET_PATH} -placeholder `"__AZURE_BROKER_CLIENT_SECRET__`" -value `"${AB_CLIENT_SECRET}`""
    Invoke-Expression -Command $cmd

    $cmd = "kubectl apply -f ${AB_SECRET_PATH}"
    Invoke-Expression -Command $cmd
}
