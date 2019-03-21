param (
    [string]$CR_PATH = $null,
    [switch]$LOCAL = $false,
	[switch]$KNATIVE = $false
)

$CURRENT_DIR = Split-Path $MyInvocation.MyCommand.Path

$cmd = "kubectl apply -f ${CURRENT_DIR}\..\resources\default-sa-rbac-role.yaml"
Invoke-Expression -Command $cmd

$cmd = "${CURRENT_DIR}\install-tiller.ps1"
Invoke-Expression -Command $cmd

$cmd = "kubectl.exe apply -f ${CURRENT_DIR}\..\resources\installer.yaml"
Invoke-Expression -Command $cmd

$cmd = "${CURRENT_DIR}\is-ready.ps1 -ns kube-system -label k8s-app -value kube-dns"
Invoke-Expression -Command $cmd

if ($LOCAL) {
    $cmd = "${CURRENT_DIR}\copy-resource.ps1"
    Invoke-Expression -Command $cmd
}
else {
    Write-Output "Non-local run is not supported!"
    exit
}

if ($CR_PATH -ne "") {
    $cmd = "kubectl.exe apply -f ${CR_PATH}"
    Invoke-Expression -Command $cmd

    if ($KNATIVE -eq $true) {
	    $cmd = @"
kubectl -n kyma-installer patch configmap installation-config-overrides -p '{\"data\": {\"global.knative\": \"true\", \"global.kymaEventBus\": \"false\", \"global.natsStreaming.clusterID\": \"knative-nats-streaming\"}}'
"@
	    Invoke-Expression -Command $cmd
    }

    $cmd = "kubectl.exe label installation/kyma-installation action=install"
    Invoke-Expression -Command $cmd
}