$CURRENT_DIR = Split-Path $MyInvocation.MyCommand.Path

Write-Output @"The script install-tiller.ps1 is deprecated and will be removed with Kyma release 1.4, please use Kyma CLI instead"@

$cmd = "kubectl apply -f ${CURRENT_DIR}/../resources/tiller.yaml"
Invoke-Expression -Command $cmd

$cmd = "${CURRENT_DIR}\is-ready.ps1 -ns kube-system -label name -value tiller"
Invoke-Expression -Command $cmd