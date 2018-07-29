param (
    [string]$DOMAIN = "",
    [string]$IP_ADDRESS = "",
    [string]$REMOTE_ENV_IP = "",
    [string]$K8S_APISERVER_URL = "",
    [string]$K8S_APISERVER_CA = "",
    [string]$ADMIN_GROUP = "",
    [string]$ENABLE_ETCD_BACKUP_OPERATOR = "",
    [string]$ETCD_BACKUP_ABS_CONTAINER_NAME = "",
    [string]$OUTPUT = ""
)

$CURRENT_DIR = Split-Path $MyInvocation.MyCommand.Path
$TPL_PATH = "${CURRENT_DIR}\..\resources\installation-config.yaml.tpl"

Copy-Item -Path $TPL_PATH -Destination $OUTPUT

(Get-Content $OUTPUT).replace("__EXTERNAL_IP_ADDRESS__", $IP_ADDRESS) | Set-Content $OUTPUT
(Get-Content $OUTPUT).replace("__DOMAIN__", $DOMAIN) | Set-Content $OUTPUT
(Get-Content $OUTPUT).replace("__REMOTE_ENV_IP__", $REMOTE_ENV_IP) | Set-Content $OUTPUT
(Get-Content $OUTPUT).replace("__K8S_APISERVER_URL__", $K8S_APISERVER_URL) | Set-Content $OUTPUT
(Get-Content $OUTPUT).replace("__K8S_APISERVER_CA__", $K8S_APISERVER_CA) | Set-Content $OUTPUT
(Get-Content $OUTPUT).replace("__ADMIN_GROUP__", $ADMIN_GROUP) | Set-Content $OUTPUT
(Get-Content $OUTPUT).replace("__ENABLE_ETCD_BACKUP_OPERATOR__", $ENABLE_ETCD_BACKUP_OPERATOR) | Set-Content $OUTPUT
(Get-Content $OUTPUT).replace("__ETCD_BACKUP_ABS_CONTAINER_NAME__", $ETCD_BACKUP_ABS_CONTAINER_NAME) | Set-Content $OUTPUT
