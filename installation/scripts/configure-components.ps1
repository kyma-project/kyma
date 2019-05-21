$CURRENT_DIR = Split-Path $MyInvocation.MyCommand.Path
$FILE_NAME = "components.env"
$FILE_PATH = "${CURRENT_DIR}\..\${FILE_NAME}"
$CM_NAME = "kyma-sub-components"
$CM_NS = "kyma-installer"

Write-Output @"The script configure-components.ps1 is deprecated and will be removed with Kyma release 1.14, please use Kyma CLI instead"@

# Do nothing if the components.env file is empty or does not exist at all
if(![System.IO.File]::Exists($FILE_PATH) -or ((Get-Content $FILE_PATH).Length -eq 0)) {
    exit
}

$cmd = "kubectl.exe create cm ${CM_NAME} --from-env-file ${FILE_PATH} -n ${CM_NS}"
Invoke-Expression -Command $cmd

$cmd = "kubectl.exe label cm ${CM_NAME} -n ${CM_NS} installer=overrides"
Invoke-Expression -Command $cmd