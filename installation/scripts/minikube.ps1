param (
    [string]$VM_DRIVER = "hyperv",
    [string]$DOMAIN = "kyma.local",
    [string]$DISK_SIZE = "30g",
    [string]$MEMORY = "8192"
)

Write-Output @"This script is deprecated and will be removed with Kyma release 1.14"@

$CURRENT_DIR = Split-Path $MyInvocation.MyCommand.Path
$KUBERNETES_VERSION = "1.12.5"

Write-Output @"
################################################################################
# Minikube setup with driver ${VM_DRIVER} and kubernetes ${KUBERNETES_VERSION}
################################################################################
"@

function CheckIfMinikubeIsInitialized() {
    $cmd = "minikube status --format '{{.Host}}'"
    $minikubeStatus = (Invoke-Expression -Command $cmd) | Out-String

    if ($minikubeStatus -ne "") {
        Write-Output "Minikube is already initialized"
        $deleteMinikube = Read-Host "Do you want to remove previous minikube cluster [y/N]"
        if ($deleteMinikube -eq "y") {
            $cmd = "minikube delete"
            Invoke-Expression -Command $cmd
        }
    }
}

function InitializeMinikubeConfig () {
    $cmd = "minikube config unset ingress"
    Invoke-Expression -Command $cmd
}

function ConfigureMinikubeAddons () {
    $cmd = "minikube addons enable metrics-server"
    Invoke-Expression -Command $cmd
}

function StartMinikube() {
    $cmd = "minikube start"`
        + " --memory ${MEMORY}"`
        + " --cpus 4"`
        + " --extra-config=apiserver.authorization-mode=RBAC"`
	+ " --extra-config=apiserver.cors-allowed-origins='http://*'"`
        + " --extra-config=apiserver.enable-admission-plugins='DefaultStorageClass,LimitRanger,MutatingAdmissionWebhook,NamespaceExists,NamespaceLifecycle,ResourceQuota,ServiceAccount,ValidatingAdmissionWebhook'"`
        + " --kubernetes-version=v${KUBERNETES_VERSION}"`
        + " --disk-size=${DISK_SIZE}"`
        + " --vm-driver=${VM_DRIVER}"`
        + " --bootstrapper=kubeadm"

    if ($VM_DRIVER -eq "hyperv") {
        $cmd += " --hyperv-virtual-switch='${env.HYPERV_VIRTUAL_SW}'"
    }

    Invoke-Expression -Command $cmd
}

function WaitForMinikubeToBeUp() {
    Write-Output "Waiting for minikube to be up..."

    $limit = 15
    $counter = 0
    $clusterStatus = ""

    while ($counter -lt $limit) {
        $counter += 1
        Write-Output "Keep calm, there are ${limit} possibilities and so far it is attempt number ${counter}"

        $cmd = "minikube status --format '{{.Host}}'"
        $clusterStatus = (Invoke-Expression -Command $cmd) | Out-String
        $clusterStatus = $clusterStatus.Trim()
        if ($clusterStatus -eq "Running") {
            break
        }

        Start-Sleep -Seconds 1
    }
}

function IncreaseFsInotifyMaxUserInstances() {
    # Default value of 128 is not enough to perform “kubectl log -f” from pods, hence increased to 524288
    $cmd = "minikube ssh -- `"sudo sysctl -w fs.inotify.max_user_instances=524288`""
    Invoke-Expression -Command $cmd
}

function AddDevDomainsToEtcHosts([string[]]$hostnamesPrefixes) {
    $n = 6 # 7 hostnames in one line, others in next line. Windows can't read more than 9 hostnames in the same line.
    $hostnames = $hostnamesPrefixes | ForEach-Object {"$_.${DOMAIN}"} # for minikube ssh
    $hostnames1 = $hostnamesPrefixes[0..$n] | ForEach-Object {"$_.${DOMAIN}"}
    $hostnames2 = $hostnamesPrefixes[ - ($n + 1)..($n * 2 + 1)] | ForEach-Object {"$_.${DOMAIN}"}
    $hostnames3 = $hostnamesPrefixes[ - ($n * 2 + 2)..-1] | ForEach-Object {"$_.${DOMAIN}"}

    $cmd = "minikube ip"
    $minikubeIp = (Invoke-Expression -Command $cmd | Out-String).Trim()

    Write-Output "Minikube IP address: ${minikubeIp}"

    $cmd = "minikube ssh 'echo `"127.0.0.1 ${hostnames}`" | sudo tee -a /etc/hosts'"
    Invoke-Expression -Command $cmd

    $winHostsPath = "C:\Windows\system32\drivers\etc\hosts"

    Get-Content -Path $winHostsPath | Select-String -Pattern $DOMAIN -NotMatch | Out-File -FilePath $winHostsPath

    "${minikubeIp} ${hostnames1}" | Out-File $winHostsPath -Append
    "${minikubeIp} ${hostnames2}" | Out-File $winHostsPath -Append
    "${minikubeIp} ${hostnames3}" | Out-File $winHostsPath -Append
}

CheckIfMinikubeIsInitialized
InitializeMinikubeConfig
StartMinikube
WaitForMinikubeToBeUp
ConfigureMinikubeAddons
AddDevDomainsToEtcHosts "apiserver", "console", "catalog", "instances", "brokers", "dex", "docs", "add-ons", "lambdas-ui", "console-backend", "minio", "jaeger", "grafana", "configurations-generator", "gateway", "connector-service"
IncreaseFsInotifyMaxUserInstances
