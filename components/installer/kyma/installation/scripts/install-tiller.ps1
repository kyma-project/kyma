$CURRENT_DIR = Split-Path $MyInvocation.MyCommand.Path

$cmd = "kubectl apply -f ${CURRENT_DIR}/../resources/tiller.yaml"
Invoke-Expression -Command $cmd

$cmd = "${CURRENT_DIR}\is-ready.ps1 -ns kube-system -label name -value tiller"
Invoke-Expression -Command $cmd