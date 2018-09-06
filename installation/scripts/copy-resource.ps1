$CURRENT_DIR = Split-Path $MyInvocation.MyCommand.Path
$LOCAL_DIR = "${CURRENT_DIR}\..\..".Substring(2)
$INSTALLER_NS = "kyma-installer"
$INSTALLER_POD = "kyma-installer"
$INSTALLER_CONTAINER = "kyma-installer"
$REMOTE_DIR = "/kyma/injected"

$cmd = "${CURRENT_DIR}\is-ready.ps1 -ns ${INSTALLER_NS} name ${INSTALLER_POD}"
Invoke-Expression -Command $cmd

$cmd = "kubectl.exe -n ${INSTALLER_NS} get pods -l name=${INSTALLER_POD} -o jsonpath='{.items[*].metadata.name}'"
$POD_NAME = (Invoke-Expression -Command $cmd | Out-String).ToString().Trim()

Write-Output "Copying kyma sources from ${LOCAL_DIR} into ${POD_NAME}:${REMOTE_DIR} ..."

$cmd = "kubectl.exe exec -n ${INSTALLER_NS} ${POD_NAME} -c ${INSTALLER_CONTAINER} -- /bin/rm -rf ${REMOTE_DIR}"
Invoke-Expression -Command $cmd

$cmd = "kubectl.exe exec -n ${INSTALLER_NS} ${POD_NAME} -c ${INSTALLER_CONTAINER} -- /bin/mkdir -p ${REMOTE_DIR}"
Invoke-Expression -Command $cmd

$cmd = "kubectl.exe cp ${LOCAL_DIR}\resources ${INSTALLER_NS}/${POD_NAME}:${REMOTE_DIR}/resources -c ${INSTALLER_CONTAINER}"
Invoke-Expression -Command $cmd

$cmd = "kubectl.exe cp ${LOCAL_DIR}\installation ${INSTALLER_NS}/${POD_NAME}:${REMOTE_DIR}/installation -c ${INSTALLER_CONTAINER}"
Invoke-Expression -Command $cmd

$cmd = "kubectl.exe cp ${LOCAL_DIR}\docs ${INSTALLER_NS}/${POD_NAME}:${REMOTE_DIR}/docs -c ${INSTALLER_CONTAINER}"
Invoke-Expression -Command $cmd

$cmd = "kubectl.exe exec -n ${INSTALLER_NS} ${POD_NAME} -- /bin/chmod -R +x ${REMOTE_DIR}"
Invoke-Expression -Command $cmd
