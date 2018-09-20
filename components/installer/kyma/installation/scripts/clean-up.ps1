$cmd = "kubectl.exe delete installation/kyma-installation"
Invoke-Expression -Command $cmd

$cmd = "kubectl.exe delete ns kyma-installer"
Invoke-Expression -Command $cmd

$cmd = "helm.exe del --purge ec-default"
Invoke-Expression -Command $cmd

$cmd = "helm.exe del --purge hmc-default"
Invoke-Expression -Command $cmd

$cmd = "helm.exe del --purge dex"
Invoke-Expression -Command $cmd

$cmd = "helm.exe del --purge core"
Invoke-Expression -Command $cmd

$cmd = "helm.exe del --purge istio"
Invoke-Expression -Command $cmd

$cmd = "helm.exe del --purge cluster-essentials"
Invoke-Expression -Command $cmd

$cmd = "helm.exe del --purge prometheus-operator"
Invoke-Expression -Command $cmd

$cmd = "kubectl.exe delete ns kyma-system"
Invoke-Expression -Command $cmd

$cmd = "kubectl.exe delete ns istio-system"
Invoke-Expression -Command $cmd

$cmd = "kubectl.exe delete ns kyma-integration"
Invoke-Expression -Command $cmd