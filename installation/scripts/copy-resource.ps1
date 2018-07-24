$CURRENT_DIR = Split-Path $MyInvocation.MyCommand.Path
$LOCAL_DIR = "${CURRENT_DIR}\..\..".Substring(2)
$INSTALLER_NS = "kyma-installer"
$INSTALLER_POD = "kyma-installer"
$REMOTE_DIR = "/kyma"

$cmd = "kubectl.exe -n ${INSTALLER_NS} get pods -o jsonpath='{.items[*].metadata.name}'"
$POD_NAME = (Invoke-Expression -Command $cmd | Out-String).ToString().Trim()

Write-Output "Copying kyma sources from ${LOCAL_DIR} into ${POD_NAME}:${REMOTE_DIR} ..."

$cmd = "${CURRENT_DIR}\is-ready.ps1 -ns ${INSTALLER_NS} name ${INSTALLER_POD}"
Invoke-Expression -Command $cmd

$cmd = "kubectl.exe exec -n ${INSTALLER_NS} ${POD_NAME} -- /bin/rm -rf ${REMOTE_DIR}"
Invoke-Expression -Command $cmd

$cmd = "kubectl.exe exec -n ${INSTALLER_NS} ${POD_NAME} -- /bin/mkdir ${REMOTE_DIR}"
Invoke-Expression -Command $cmd

$cmd = "kubectl.exe cp ${LOCAL_DIR}\resources ${INSTALLER_NS}/${POD_NAME}:${REMOTE_DIR}/resources"
Invoke-Expression -Command $cmd

$cmd = "kubectl.exe cp ${LOCAL_DIR}\installation ${INSTALLER_NS}/${POD_NAME}:${REMOTE_DIR}/installation"
Invoke-Expression -Command $cmd

$cmd = "kubectl.exe cp ${LOCAL_DIR}\docs ${INSTALLER_NS}/${POD_NAME}:${REMOTE_DIR}/docs"
Invoke-Expression -Command $cmd

$cmd = "kubectl.exe exec -n ${INSTALLER_NS} ${POD_NAME} -- /bin/chmod -R +x ${REMOTE_DIR}"
Invoke-Expression -Command $cmd
